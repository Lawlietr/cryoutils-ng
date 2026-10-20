# Bug 報告與根因：swap resize 會摧毀 SteamOS 的 zram（P0.7）

- **發現日期**：2026-10-20（真實 Steam Deck / SteamOS，用戶實測回報）
- **嚴重性**：P0 —— 破壞 SteamOS 原廠記憶體架構，且**會自我固化**（一次壞掉後每次都不再嘗試恢復），只有重開機能回來了
- **狀態**：根因已定位（有 journal 實證）；**修法已拍板 = 方法 A**，程式碼修正待回 Dev VM 施作
- **受影響版本**：現行 `core/swap.go`（含 Phase 6.6 binary，md5 `a9ab4bf61afd96397ee650ccf7b60817`）

---

## 1. 症狀

只要用本 app 改 swap size（CLI `swap <GB>` 或 UI 的 Swap 區塊），改完之後 zram 就消失：

- `/proc/swaps` 只剩 `/home/swapfile`，總 swap 從 23 GiB 掉到 16 GB
- `zramctl` 輸出為空、`/sys/block/zram0/disksize` 變成 `0`
- `systemctl status systemd-zram-setup@zram0.service` → `inactive (dead)`

## 2. 根因（已用 journal 實證確定，非推測）

`core/swap.go:109` `DisableSwap()` 用的是 `sudo swapoff -a`。在 SteamOS 3.6+ 上，zram 是由 **zram-generator**（systemd generator）管理：`swapoff -a` 把 zram0 上的 swap 關掉後，systemd 會**沿著單位依賴把整個 zram 裝置拆掉**，包括把 `disksize` 寫 0（reset device）：

```
16:33:30  sudo /usr/bin/swapoff -a                       ← core/swap.go:109 DisableSwap()
16:33:42  systemd[1]: home-swapfile.swap: Deactivated successfully.
16:33:51  systemd[1]: dev-disk-by\x2dlabel-zram0.swap: Deactivated successfully.
16:33:51  systemd[1]: dev-zram0.swap: Deactivated successfully.
16:33:51  systemd[1]: Stopping Create swap on /dev/zram0...
16:33:51  kernel: zram0: detected capacity change from 15159296 to 0     ← 裝置被 reset
16:33:51  systemd[1]: systemd-zram-setup@zram0.service: Deactivated successfully.
          ExecStop=/usr/lib/systemd/system-generators/zram-generator --reset-device zram0
          (code=exited, status=0/SUCCESS)                                 ← 關鍵：ExecStop 寫 disksize=0
16:33:51  sudo /usr/bin/dd if=/dev/zero of=/home/swapfile bs=1G count=12
16:34:01  sudo /usr/bin/chmod 600 /home/swapfile
16:34:01  sudo /usr/bin/mkswap /home/swapfile
16:34:01  sudo /usr/bin/swapon /home/swapfile
          kernel: Adding 12582908k swap on /home/swapfile.  Priority:-2 ...
16:34:01  sudo /usr/bin/swapon /dev/zram0                    ← core/swap.go:231 P0 補救
          （journal 裡完全沒有對應的 kernel "Adding ... swap on /dev/zram0"）→ 失敗
```

**結論**：`swapon /dev/zram0` 是對一個 **容量 0** 的裝置操作，必然失敗（EINVAL）。失敗又被 `core/swap.go:232` 吞掉（註解 "zram is a nice-to-have"），使用者完全看不到。

### 2.1 bug 會自我固化

第二次 resize（`16:34:31`，`count=16`）的 journal 裡**完全沒有 `swapon /dev/zram0`**：
因為此時 zram 已不在 `/proc/swaps`，`core/swap.go:210` 的 `zramActive, _ := e.GetZramStatus()` 回傳 false，
`core/swap.go:229` 的 `if zramActive` 直接跳過。

→ **第一次弄壞之後，之後每一次 resize 都不再嘗試恢復 zram**，只有重開機才會回來。

### 2.2 時序是非同步的（任何「事後補救」都會 race）

`swapoff -a`（16:33:30）到 zram teardown（16:33:51）中間隔了 **21 秒**，是 systemd 非同步處裡。
所以「先 resize 再補 swapon」這條路本身就會跟 teardown 赛跑，從設計上就不該走這條路。

## 3. 順帶查出的關聯缺陷

| # | 缺陷 | 位置 | 說明 |
|---|---|---|---|
| A-1 | **優先權也是錯的** | `core/swap.go:231` | SteamOS 原廠 zram 是 `swap-priority = 100`（見 §4）。我們的 `swapon /dev/zram0` 不帶 `-p`，會拿到 kernel 預設值（journal 顯示 swapfile 是 `Priority:-2`）。即使 swapon 成功，**zram 優先權會低於硬碟 swapfile**，kernel 永遠先寫碟 → zram 等於沒用。也就是舊 P0 修正的「成功路徑」本身也是錯的。 |
| A-2 | **失敗證據被自我抹除** | `cmd/cryoutilities/main.go:34`、`cmd/desktop/main.go:45` | 兩者都在啟動時 `os.Remove(core.LogFilePath)`，且共用同一個 `~/.cryoutils_ng/cryoutils_ng.log`。16:34 那次 resize 的 log（含 `Failed to re-enable zram:`）已被 16:36 啟動的 desktop 覆蓋 → 永遠查證不到。 |
| A-3 | **`-a` 的爆破半徑過大 + OOM/earlyoom 風險** | `core/swap.go:109` | `-a` 動到所有 swap。該 Deck 有 `earlyoom` 在跑（journal：`earlyoom[941]: mem avail: ... swap free: ...`），實測 resize 前 swap used 11 GiB、RAM available 僅 6.9 GiB → `swapoff -a` 期間 earlyoom 可能直接殺程序。改成只關 swapfile 後，遷移期間 zram 仍在，可吸收大量換頁，風險大幅降低。 |

## 4. SteamOS 原廠 zram 設定（單一來源，別自己重現）

`/usr/lib/systemd/zram-generator.conf`（Valve 維護，holo 的一部分）：

```
zram-size = ram/2
compression-algorithm = zstd
swap-priority = 100
fs-type = swap
```

相關單位（`systemctl list-unit-files | grep -i zram`）：

```
systemd-zram-setup@.service     static
dev-zram0.swap                  generated
```

`dev-disk-by-label-zram0.swap` / `dev-disk-by-diskseq-2.swap` 等都是同一個 zram0 的 generated 別名單位。

## 5. 修正方法（已拍板：方法 A）

### 方法 A —— 主要修正：不要碰 zram（**採用**）

1. `DisableSwap()`：`sudo swapoff -a` → 改成只關 swap file 本身：
   `sudo swapoff <e.SwapFileLocation>`
2. **移除** `ChangeSwapSize()` 裡的 zram 重啟區塊（`core/swap.go:227-233`）與 `zramActive` 前置偵測（`core/swap.go:210`）
   —— zram 從頭到尾沒被動到，就不需要任何補救；優先權 100 與 systemd 狀態自然保持
3. **新增前置安全檢查**（避免 earlyoom / OOM）：讀 `/proc/swaps` 該 swapfile 的 `Used` 欄位與 `/proc/meminfo` 的 `MemAvailable`，
   若 `used_swap_pages > MemAvailable` 則**明確報錯中止**，不執行 swapoff
4. 所有 swap 相關指令補上 `-k` 與密碼 pipe 一致性（與 `core/sudo.go` 的 `RenewAuth()` 規則一致），並把 `swapoff`/`dd`/`mkswap`/`swapon` 的 **stderr 帶進 error message**（現在所有指令都用 `.Output()` 且把 err 換成無訊息的固定字串，無法查證）

預期結果：`/proc/swaps` 在 resize 前後都保持 zram0（priority 100）+ swapfile，總 swap 不變。

### 方法 B —— 安全網（可選，之後再加）

万一双 zram 還是因為別的原因不见了，用 **systemd** 恢復，不是裸 `swapon`：

```bash
systemctl restart systemd-zram-setup@zram0.service
# ExecStart 會重建 disksize(ram/2) + mkswap + swapon -p 100 + zstd
```
並等 `dev-zram0.swap` 變 active。**絕對不要**只用 `swapon /dev/zram0`（disksize=0 時必然 EINVAL）。

### 方法 C —— 可觀測性（與 A 一起做）

log 改成 `O_APPEND`（拿掉 `os.Remove`），或 CLI / desktop 各用獨立 log 檔，讓失敗留痕。

## 6. 本機（SteamOS 端）已做的復原動作 2026-10-20

```bash
systemctl restart systemd-zram-setup@zram0.service     # deck 於 wheel 群組，polkit 允許，不需密碼
```

復原後驗證（全部通過）：

```
$ cat /sys/block/zram0/disksize
7761559552
$ zramctl
NAME       ALGORITHM DISKSIZE DATA COMPR TOTAL STREAMS MOUNTPOINT
/dev/zram0 zstd          7.2G   4K   64B   20K         [SWAP]
$ cat /proc/swaps
/dev/zram0      partition  7579644  0  100
/home/swapfile  file      16777212  0   -2
$ grep SwapTotal /proc/meminfo
SwapTotal:      24356856 kB
$ systemctl is-active systemd-zram-setup@zram0.service dev-zram0.swap
active
active
```

> 注意：`swapoff -a` 造成的 zram 消失**不是**「等 systemd 裝置掃描約 18 分鐘會自己回來」（舊 P0 階段的假設錯誤）。
> 它不會自己回來，必須 `systemctl restart systemd-zram-setup@zram0.service` 或重開機。

## 7. 驗收條件（回 Dev VM 改完後，真機要逐項驗）

- [ ] `sudo cryoutils-ng swap 16`（與現值相同）後：`/proc/swaps` **zram0 仍在**，Priority 仍為 **100**，Size 仍 7579644
- [ ] `sudo cryoutils-ng swap 12` 再 `swap 16`：兩次都保持 zram 存在；`zramctl` 顯示 zstd / 7.2G
- [ ] `SwapTotal` 在 resize 前後差異只來自 swapfile 大小变化，不含 zram 消失
- [ ] app log 中不再出现 `Failed to re-enable zram`，且日誌可追溯（append 模式）
- [ ] 提前把 swap used 拉高於 MemAvailable 時，resize **中止並報明原因**，而不是把系統壓到 earlyoom 殺程序
- [ ] web UI / native UI / CLI `status` 的 `ZramActive` 在 resize 前後都為 true
- [ ] 原版 CryoUtilities 不受影響（`~/.cryo_utilities/` 內容不變）

## 8. 相關檔案

- `core/swap.go`：`DisableSwap()`、`ResizeSwapFile()`、`InitNewSwapFile()`、`ChangeSwapSize()`、`GetZramStatus()`、`GetZramSizeBytes()`
- `core/sudo.go`：`RenewAuth()`（`sudo -S -k`）、`TestAuth()`
- `core/engine.go`：`GetStatusSummary()`、`getTotalSwapGB()`
- `todo.md`：P0（ZRAM 支援，代碼層已完成但設計錯誤）、P0.7（本檔對應的修正項）
