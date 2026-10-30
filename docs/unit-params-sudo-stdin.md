# P0.9 — unit 參數不可偵測 / 不可用：sudo 的 stdin 被當資料、狀態讀取不必要地要 root

> 日期：2026-10-30（Dev VM `a1` 分析 + SteamOS 真機 `deck@192.168.1.232` 實證與驗收）
> 症狀：SteamOS 上 UI（native 與 web 皆同）的 **Huge Pages、Shared memory in THP、Compaction Proactiveness、HugePage Defrag、Page Lock Unfairness** 五項「不可偵測與不可用」：狀態永遠 off/unknown，點擊只彈錯誤對話框。
> 受影響代碼：`core/sudo.go` `runSudoWithInput()`、`core/util.go` `getUnitStatus()` / `removeFile()` / `writeFile()`、`core/memory.go` `SetHugePages()`。
> 引入時點：P0.8（2026-10-20）把 `exec.Command("sudo", ...)` 收攏成 `runSudo()` 的那一次重構。**swap 相關操作沒被波及**（`dd`/`mkswap`/`swapon`/`swapoff` 不吃 stdin payload），所以同一天 swap resize 真機驗收是通過的 —— 這是為什麼這個缺陷當時沒被抓到。

---

## 1. 通話鏈（一句話）

`sudo -S` 沒有 `-k` → sudo 的 timestamp 還在有效期內（`RenewAuth`/`TestAuth` 剛驗證過）→ **sudo 完全不讀 stdin** → 我們寫進 stdin 的整段 payload（**第一行是明文 sudo 密碼**，第二行才是要寫的值）被 `tee` 原樣當成資料 → `tee` 先把密碼那一行寫進 sysfs/procfs → kernel 回 **EINVAL** → `tee` exit 1 → `setUnitValue()` 回報錯誤 → UI 彈錯誤對話框。

與此同時，**讀取**路徑 `getUnitStatus()` 也走 sudo，而這些檔案其實 `0644 root:root`、一般使用者可讀 → 沒有密碼時（web 模式永遠沒有密碼可給；native 尚未解鎖時也沒有）直接 `ErrNoSudoPassword` → 五個 `get*Status()` 一律回 false → 狀態顯示成「關閉」而不是「不知道」。

兩件事合起來就是用戶看到的「不可偵測與不可用」。

---

## 2. 真機實證（SteamOS，2026-10-20 17:56，desktop binary md5 `a2bcd4bf0330ef8bc3f2fd2ed66a41bc`）

`~/.cryoutils_ng/cryoutils_ng.log`（密碼已在真機 scrub 成 `[REDACTED-PASSWORD]`）：

```
INFO  Enabling hugepages...
INFO  Removing /proc/sys/vm/nr_hugepages
ERROR rm /proc/sys/vm/nr_hugepages: exit status 1: rm: 無法刪除 '/proc/sys/vm/nr_hugepages': 拒絕不符權限的操作
INFO  Couldn't delete /proc/sys/vm/nr_hugepages , likely missing.
INFO  Writing always for param hugepages to memory.
ERROR tee /sys/kernel/mm/transparent_hugepage/enabled: exit status 1: [REDACTED-PASSWORD]
      alwaystee: /sys/kernel/mm/transparent_hugepage/enabled: 不適用的引數
```

三個要點都在這五行裡：

1. `exit status 1` 的輸出內容以**密碼**開始，接著才是 `always` —— 這是 `tee` 把它收到的 stdin 原封不動 echo 回 stdout，證明 **sudo 沒有消耗密碼那一行**，密碼與值一起到了 `tee`。
2. `不適用的引數` = EINVAL：kernel 對 `/sys/kernel/mm/transparent_hugepage/enabled` 只接受 `always|madvise|never`，寫 `密碼`當然被拒。
3. 五個參數（hugepages / shmem_enabled / compaction_proactiveness / defrag / page_lock_unfairness）**每一次 toggle 都是同一個形狀**，所以「不可用」是必現的，不是機率性。

另外 17:55:47 那一組：

```
ERROR Unable to get current hugepages value
ERROR Unable to get current shmem_enabled
ERROR Unable to get current compaction_proactiveness
ERROR Unable to get current defrag
ERROR Unable to get current page_lock_unfairness
```

就是「不可偵測」那 half —— 讀取走到 sudo 而當下沒有密碼。

### 2.1 Dev VM 先用真實 sudo 重現（非 root 帳號 `lawliet`）

| 情境 | 命令 | 目標檔案最終內容 |
|---|---|---|
| timestamp 已快取、**無** `-k`（P0.8 的寫法） | `printf 'pw\nVALUE\n' \| sudo -S tee target` | `pw\nVALUE\n` ← **密碼被寫進檔案** |
| 加 `-k`（每次都要驗證） | `printf 'pw\nVALUE\n' \| sudo -S -k tee target` | `VALUE\n` |
| 讀取、無 `-k`、當下要驗證 | `sudo -S cat secret`（stdout+stderr 合併） | `[sudo] password for lawliet: SECRET_42` ← **提示語混進讀值** |

第三列是第二個缺陷的來源：sudo 的密碼提示寫在 **stderr**，而 `runSudoWithInput` 把 stdout/stderr 塞进同一個 buffer，所以讀到的「值」前面掛了一段提示語，比對永遠不吻合（列表型參數靠括號解析僥倖存活，純數值型如 `compaction_proactiveness`、`swappiness`、`defrag` 會被弄壞）。

---

## 3. 修正規格（已實作）

### 3.1 `core/sudo.go`

- `runSudoWithInput` 一律 **`sudo -S -k`**：`-k` 強制每次驗證 → sudo 必定從 stdin 讀走**恰好一行**密碼 → payload 剩下部分才是命令的資料。**這是本次的核心修正。**
- **stdout / stderr 分開**兩個 buffer；回傳值只取 stdout；錯誤訊息帶 **command line + stderr tail**（不帶 stdout —— `tee` 的 stdout 是輸入的 echo，既無診斷價值又可能含敏感內容）。
- 新增 `(*Engine).redact()`：任何要進 log 或回給 UI 的字串先抹掉 `e.Password`。
- 新增 `useSudoAlways`（僅測試設定）：讓「root 也要用 sudo 包裝」這條路徑在任意機器上都能被單元測試覆蓋，不會因為測試剛好以 root 跑而被跳過。
- 保留 P0.8 的兩個性質：沒有密碼時**不啟動 sudo**（`ErrNoSudoPassword`）、一律不使用裸 `sudo`（永不碰 tty）。

### 3.2 `core/util.go`

- `getUnitStatus()`：先 `unitFileReader`（預設 `os.ReadFile`）直接讀；
  - 成功 → 直接解析（**不碰 root**）
  - `EPERM` → 才退回 `runSudo("cat", path)`
  - 其他錯誤（例如 ENOENT，參數在此 kernel 不存在）→ 直接回報，不浪費一次特權嘗試
- 抽出 `parseUnitValue()`：列表型 sysfs（`always madvise [never]` → `never`）與純數值两条分支，並加單元測試。
- 未知參數名稱明確報錯（原本會 `UnitMatrix[param]` 拿到空字串然後 `cat ""`）。
- `removeFile()` 改用 **`rm -f`**：revert 一個不存在的 `/etc/tmpfiles.d/<param>.conf` 不該被當成失敗。
- `writeFile()` 的暫存檔（在自己 install 目錄內）改用 `os.Remove`，不再為它跑一次 `sudo rm`。

### 3.3 `core/memory.go`

- `SetHugePages()` 拿掉 `removeFile(NHPTestingFile)`：**procfs / sysfs 的檔案不能 unlink，root 也不行**（真機 log 已證明 EPERM）。註解留了說明；若真的要釋放已保留的 huge pages，正確做法是對 `/proc/sys/vm/nr_hugepages` 寫 `0`——**那是行為變更，本輪刻意不做，需用戶拍板**。

### 3.4 log 權限

- `cmd/cryoutilities/main.go`、`cmd/desktop/main.go`：`os.OpenFile(core.LogFilePath, …, 0666)` → **`0600`**。原本是**世界可讀**，而內容一度含明文 sudo 密碼。

---

## 4. 資安影響（同一批代碼的副作用，已處理）

- `tee` 會 echo 它的 stdin，而那份 stdin 的第一行是 **明文 sudo 密碼**；失敗時 `commandError()` 又把輸出帶進 log → 真機 log 內測得 **5 處**含密碼，且 log mode 是 **0666**。
- 已在真機處理：5 處替換成 `[REDACTED-PASSWORD]`；`chmod 600`。
- 修好後的 binary 跑完完整驗收（見 §6），log 内含密碼 **0 處**。
- **仍建議更新 Deck 的 sudo 密碼**：它已在這台機器上一份可讀檔案裡存在了約 10 天。

---

## 5. 迴歸測試（Dev VM，`core` 全套綠）

fake sudo 重寫成**重現 sudo 的 stdin 契約**：只有「真的需要驗證」時（有 `-k`，或 timestamp marker 不存在）才吃掉第一行密碼，否則把整段 stdin 原封不動交給命令 —— 這樣舊 bug 會被測試抓到，而不是靠環境碰運氣。

| 測試 | 鎖定什麼 |
|---|---|
| `TestRunSudoWithInputNeverLeaksPasswordIntoTheCommand` | 連跑兩次（第二次是快取態）：參數檔案必須是 `always\n`；每次 sudo 呼叫都要帶 `-k`；log 內不得出現密碼 |
| `TestRunSudoReturnsStdoutOnly` | 回傳值只有 stdout，不含 stderr 的提示/雜訊 |
| `TestRunSudoFailureCarriesStderrButNotThePassword` | 失敗仍要可診斷（command + stderr），但不得含密碼 |
| `TestRunSudoWithoutPasswordNeverSpawnsSudo` | 沒密碼時不啟動 sudo；`setUnitValue` 回 `ErrNoSudoPassword`；**讀取不需要 sudo** |
| `TestGetUnitStatusReadsWithoutRoot`（`core/unit_test.go`） | 狀態讀取零次 sudo 呼叫 + 列表/數值解析正確 |
| `TestGetUnitStatusFallsBackToSudoWhenDenied` | 模擬 `EPERM` 時仍會退回 sudo 並正確解析 |
| `TestParseUnitValue` / `TestRemoveFileToleratesAMissingPath` / `TestSetUnitValueReportsFailures` | 括號解析、`rm -f` 語意、寫入失敗必須被看見 |

**三個回歸實驗**（把舊行為改回去，新測試必須紅 —— 已跑過確認）：

1. 拿掉 `-k` → `parameter file = "<密碼>\nalways\n", want "always\n"`
2. 合併 stdout/stderr → `runSudo output = "VALUE\nPROMPT_AND_NOISE\n", want only stdout`
3. 讀值改回走 sudo → `getUnitStatus = "", sudo password not provided`

---

## 6. 真機驗收（2026-10-30，Steam Deck，kernel `7.2.7-valve2-2-neptune-72`，sshd 首次使用）

> **md5 對照（2026-10-30 最終佈署）**：CLI `1751459e162dfb1ed15cf1c11686f057` · desktop `cc95f58e23032b23239132a6b5001400` · probe `640cce9a27f7c4af97a0dbb06e7c23b0`，三支的 md5 在 Dev VM 與 Deck 上完全相同（從已提交的來源重建，gofmt 後的版本）。舊版留檔：`cryoutils-ng.bak-p08`、`cryoutils-ng-desktop.bak-p08`、`cryoutils-ng-desktop.bak-425bc04`。驗收腳本：`tools/p09check/p09_check.sh`（唯讀；在 Deck 上 `bash /tmp/p09_check.sh`，實測 RESULT: 全部 PASS）。既有 log 已移到 `cryoutils_ng.log.pre-p09`，驗收從乾淨 log 開始。

佈署：先備份 `cryoutils-ng.bak-p08` / `cryoutils-ng-desktop.bak-p08`，再推送新二進位；兩支 md5 與 Dev VM 建置一致（CLI `9aea9457…`、desktop `d4558b4f…`）。

**讀取 A/B** —— 同一台 Deck、同一指令 `cryoutils-ng status`、deck 身份、**完全不給密碼**：

| | 舊 binary（P0.8） | 新 binary |
|---|---|---|
| HugePages | false | **true** |
| ShMem | false | **true** |
| CompactionProactiveness | false | false |
| Defrag | false | false |
| PageLockUnfairness | false | false |
| log | 5 行 `Unable to get current …` | **零錯誤行** |
| kernel ground truth（deck 直接 `cat`） | — | `[always] madvise never` / `[advise]` / `1` / `20` / `5`（與上表逐項吻合） |

**寫入 round-trip**（`tools/p09probe`，以 deck 身份經**真 sudo**，每個值測完還原）：

- 五個參數各自 `Toggle()` 兩次：`err=<nil>`，值確實改變並回到原值
- `Set()` 後 `/etc/tmpfiles.d/<param>.conf` **確實建立**（`/` 在此 Deck 是 `rw`，tmpfiles.d 持久層正常）
- `Revert()` 後值回 default、conf 被移除
- `probe failures counted: 0`
- 收尾：探測留下的 `hugepages.conf` / `shmem_enabled.conf` 已 `sudo rm -f`，`/etc/tmpfiles.d` 回空；kernel 值全數回原

**副作用檢查**：`/dev/zram0` 仍在（zstd、7.2G、`[SWAP]`，Priority 未變）、swapfile 16 GB 未動、log 內密碼 0 處、log mode 600。

**還差**：用戶在 native UI 實際點五個 toggle（checkbox 即時反映、不再彈錯誤）—— 與 Phase 6.6 的雙解析度驗收一同做。

---

## 7. 教訓（寫進 AGENTS.md 的規則）

1. **透過 stdin 送密碼時，務必 `-k`**。「密碼走 stdin 的第一行」這個約定只在 sudo **真的讀 stdin** 時成立；快取態下它不讀，密碼就成了命令的資料。任何「stdin 前半是認證、後半是資料」的協議都必須強制每次認證（或改用 `SUDO_ASKPASS` / `--stdin` 之外的獨立通道）。
2. **不要把 stdout 與 stderr 合成一個流**去做「讀值」。提示語、進度訊息與值會混在一起。
3. **讀狀態不需要 root**：`/proc/sys/**` 與 `/sys/kernel/mm/**` 的調參檔案幾乎都是 `0644 root:root`。把讀取包進 sudo 只是把「沒有密碼」變成「讀不到」，還附帶每 5 秒一輪的 log 噪音。
4. **`tee` 會 echo 它的輸入**：用它寫值時，stdout 不是診斷資訊；而且只要 stdin 裡出現過敏感字串，就可能被 echo 進 log。任何錯誤字串進 log 前要 redact。
5. **procfs/sysfs 不能 unlink**（root 也不行）。要重設就寫值，不要 `rm`。
6. 同一次重構把「讀」和「寫」一起搬進同一個 helper 時，要**分開驗證讀取與寫入**：swap 那批不吃 stdin 所以沒事，unit 那批吃 stdin 所以全壞 —— 只驗 swap 不足以證明重構無副作用。
