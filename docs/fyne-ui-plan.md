# Fyne 原生 UI（全新撰寫）— 施工手冊（Phase 6.5）

> **交接文件**：本文記錄新 Fyne 原生 UI 的全部既定決策、已排除方案與施工步驟。
> **給未來的 agent：標記 [DECIDED] 的項目不要再重新決策**，照做即可；只有「§8 未決事項」需要用戶確認。
> 撰寫基礎：2026-08 Dev VM 實地偵察結果（環境、API、檔案結構皆以 repo 現況為準）。
> **禁止自行 commit/push**（AGENTS.md 工作規則）。

---

## 0. 最終決策（先讀這一段）

**背景**：Web UI 的 Chromium `--app=` 無邊框視窗在真實 Steam Deck 上驗證失敗
（2026-08-11：Flatpak Chrome/Brave `--app=` 皆「拒絕連線」，`blink.mojom.Widget` 錯誤，
見 `todo.md` Phase 5.5 Deck 測試結果）。`steam://openurl` 是次佳 fallback（Steam 內建 CEF，外觀不夠好）。

**決策**：**從零撰寫全新 Fyne 原生 UI**（不沿用、不 rebase `internal/` 舊 Fyne UI）。
Web UI **保留**（Phase 7 Decky 要重用 React 前端 + 作為 fallback），兩種 UI 可選啟動。

### [DECIDED] 決策清單（勿再評估）

| # | 決策 | 說明 |
|---|------|------|
| 1 | 技術：**Fyne v2** | 不用 Gio / GTK。用戶偏好 + SteamOS Desktop Mode 是 KDE |
| 2 | Fyne 版本：**v2.7.4**（保守預設） | v2.8.0 是 2026-07-13 剛發布的大版（1000+ commits、GPU shapes、新 window APIs）。PoC（Step 1）通過後可選升級試 v2.8.0，有問題就維持 v2.7.4 |
| 3 | **不沿用 `internal/` 舊 Fyne UI** | 舊 UI 層 1330 行 + handler 層 1170 行；handler 含已知 bug（zram0 `return error`、舊 sudo 無 `-k`）；無 i18n（全硬編碼英文）；5-tab 違反 locked 版面。「只重用 UI 層」也僅省 30–40% 版面工作，核心工作（i18n 重構、dark theme、ZRAM 顯示、sudo 對話框、全部改接 `core.Engine`）一行都省不到 |
| 4 | 版面：**單頁、直式、無 tab**（與 Web UI 一致，locked） | 見 §2.2 |
| 5 | 主題：**dark**（`theme.DarkTheme()`） | locked |
| 6 | i18n：**共享單一來源 + runtime drop-in 資料夾**（見 §3） | 放入 `~/.cryoutils_ng/locales/xx.json` 即生效，不用重編 |
| 7 | ZRAM 顯示：與 Web UI 完全同等 | swap file / ZRAM Size / ZRAM Status / Total Swap |
| 8 | Sudo：Fyne UI 有密碼對話框（首啟 + 失敗時） | Web UI 目前**沒有**密碼輸入（既有缺口，見 §2.4 附註）；Fyne 走自己的對話框。密碼只存記憶體，不落磁碟 |
| 9 | `internal/` 刪除：Phase 6.5 全部接受後，**單獨一個 commit** 移除 | 移除前它是 root module 編譯的參考基準（Step 1 gate 用它確認 Fyne 升級沒炸） |
| 10 | `cmd/desktop` 新增 `-ui web\|native`，**預設 `web`** | native 在真實 Deck 接受完成前，預設值**不切**；接受後才把預設改成 `native`（一個 flag 預設值的 commit） |
| 11 | `cryoutils-ng-desktop` 二進位改為 `CGO_ENABLED=1` 建置 | Fyne 需要 CGO + 系統 GL/X11 庫。**CLI `cryoutils-ng` 維持 `CGO_ENABLED=0` 靜態** |
| 12 | 嵌入 Noto Sans TC（Regular + Bold）確保中文字形 | Fyne 2.5+ 用 go-text/typesetting，CJK shaping 已改進，但字形仍需字体涵蓋 |

### [REJECTED] 已排除方案（理由已評估，不要再開）

| 方案 | 排除理由 |
|------|----------|
| Rebase `internal/` 舊 Fyne UI（「UI 層沿用 + handler 改接 core」） | 見決策 3；另外升級 Fyne 版本會使舊 UI 層也需改 API，重用價值進一步下降 |
| Gio | 曾以「保留 CGO=0」為優勢，但決策 11 已接受 UI 二進位走 CGO，優勢消失；生態系較弱 |
| GTK / libadwaita | 用戶明確排除；CGO + glibc 可攜性風險大；Go 綁定品質弱 |
| Electron / CEF | 二進位 100MB+ 過重 |
| WebView 內嵌（Wails / webview / webkit2gtk） | Stock SteamOS 無 webkit2gtk（Valve issue #1851）+ rootfs 唯讀 |
| Fyne v2.3.1（root go.mod 現行版本） | 2022 年代 textlayout，CJK 渲染差、無 `fyne.Do` 等新 thread API |
| Sudo 密碼落磁碟（如 `~/.cryoutils_ng/auth.json`） | 明文 sudo 密碼安全取捨；本期不做，日后再議 |
| 為 Fyne 單獨出第二個二進位（`cryoutils-ng-native`） | 打包/安裝/更新複雜度翻倍；單二進位 + `-ui` flag 足夠 |

---

## 1. 環境事實（2026-08 偵察結果）

| 項目 | 現況 |
|------|------|
| Dev VM | Ubuntu 24.04.4 LTS（`a1` / `root@192.168.1.15`），Go 1.26.5（`/usr/local/go/bin/go`），gcc 13.3.0，根目錄可用空間 ~15G |
| Fyne v2.3.1 編譯 | `CGO_ENABLED=1 go build ./internal/` **今日可過**（GL/X11 headers 已齊）→ 舊 code 是有效的升級基準 |
| Xvfb | **已裝**（`/usr/bin/Xvfb`、`xvfb-run`） |
| llvmpipe | **缺**：`/usr/lib/x86_64-linux-gnu/dri/swrast_dri.so` 不存在 → 需 `apt-get install libgl1-mesa-dri` |
| 截圖工具 | 缺 scrot → `apt-get install scrot`；另裝 `mesa-utils`（`glxinfo` 驗證 renderer） |
| Fyne 版本 | v2.7.4 = 2.7.x 最後 stable；v2.8.0 = 最新 stable（2026-07-13）。`go get fyne.io/fyne/v2@vX` 即可 |
| SteamOS | 無 sshd（**deck 端主動 `scp root@192.168.1.15:...` 拉取**）、rootfs 唯讀、Desktop Mode = KDE Plasma、glibc 2.39+（Arch 基底） |
| glibc 相容性 | Ubuntu 24.04 也是 glibc 2.39 → 理論相容；**真實 Deck 驗證**。若二進位要求更高 glibc → fallback：在 Arch Linux container/chroot 內建置 |
| 二進位體積預期 | Fyne 約 15–25MB + 嵌入 web dist（小）+ Noto Sans TC 兩個 TTF（每個數 MB）≈ 25–40MB |

**Dev VM 待裝套件（Step 0，先 `df -h /`）**：`libgl1-mesa-dri`、`scrot`、`mesa-utils`（合計 < 200MB，遠低於 3GB 門檻）。

---

## 2. 架構

### 2.1 模組佈局

```
cmd/desktop/
  main.go          + `-ui web|native` flag；native 分支：不啟 HTTP/token/瀏覽器，
                   建立 core.Engine 後直接在 process 內跑 fyneui.Run(e)
  (web 模式既有 code 全部不動)

ui/fyneui/         [新增] package fyneui —— 新原生 UI
  app.go           Run(e *core.Engine)：NewAppWithID / dark theme / 字體 / window / 單頁 scroll
  sections_status.go    系統狀態（含 ZRAM）
  sections_swap.go      Swap 調整（resize 流程）
  sections_memory.go    記憶體參數開關
  sections_vram.go      VRAM 唯讀
  sections_presets.go   Recommended / Stock
  sections_gamedata.go  Steam 遊戲資料 sync / cleanup
  sudo_dialog.go        sudo 密碼對話框
  lang.go               語言選單 + 整頁重繪 + ui.json 持久化
  fonts.go              //go:embed fonts/*.ttf + 設為預設字體
  fonts/                NotoSansTC-Regular.ttf / NotoSansTC-Bold.ttf + OFL.txt

i18n/              [新增] package i18n —— locale 載入（Go 端單一來源）
  i18n.go          //go:embed locales/*.json + Load/Available/T
  i18n_test.go     fallback / override / T 單元測試
  locales/en.json      canonical（git mv 自 web/src/locales/）
  locales/zh-TW.json

core/              **零改動** —— GetStatusSummary() 已含 ZramSizeGB/ZramActive/TotalSwapGB（P0 已完成）
web/               只動 package.json（sync-loc）；src/ 全部不動
```

**import 關係**：`ui/fyneui` → `cryoutils-ng/core`、`cryoutils-ng/i18n`（同 module 直接 import）。
`i18n` → `cryoutils-ng/core`（只為 `core.InstallDirectory` 常數；無迴圈）。
`cmd/desktop` → 兩邊都 import（native 分支）。

**已知取捨（接受）**：`cmd/desktop` 二進位以 `go:embed` 內嵌 web dist，因此**即使 `-ui native` 使用者，
二進位仍含 web UI 資料**（數百 KB，可忽略）。不為省這幾百 KB 拆二進位。

### 2.2 UI 版面（與 Web UI 六區塊同等；單頁、直式、無 tab，locked）

由上而下（全部在一個可捲動欄位內，progress bar 固定在捲動容器**上方**）：

1. **Header**：`CryoUtils NG` 標題 + 版本（`core.CurrentVersionNumber`）+ 語言選單（Select）
2. **Progress bar**：長任務時顯示（swap resize / gamedata sync / cleanup / presets）
3. **System Status**：Swap File 路徑、Swap Size（file）、**ZRAM Size、ZRAM Status（Active/Inactive）、Total Swap**、Swappiness（✓ Recommended / ✗ Default 標記）、VRAM —— 全部來自 `e.GetStatusSummary()`
4. **Swap Settings**：size 下拉（`e.GetAvailableSwapSizesStr()`）+ 現值標示 + **Resize**（先確認對話框 → 警告 swap 會短暫停用 → progress → 完成後刷新狀態）
5. **Memory Settings**：5 參數開關（hugepages / shmem / compaction_proactiveness / defrag / page_lock_unfairness），label 用 `memoryParams.*` 翻譯鍵
6. **VRAM**：唯讀 + 說明文字（與 Web 相同措辭：實際改 VRAM 要走 BIOS）
7. **Presets**：`Recommended` / `Stock` 兩按鈕（執行後刷新全部狀態）
8. **Game Data**：library 選擇（`e.FindDataFolders()`）+ `Sync Game Data` + `Cleanup Orphaned Data`（皆長任務 + progress）

### 2.3 執行緒模型（Fyne 重點，最容易出錯）

- Fyne **所有 UI 更新必須在主 goroutine**。跨 goroutine 一律用 **`fyne.Do(func(){ ... })`**（v2.5+ API；2.7.4 有）。
- 標準模式（每個長任務按鈕照抄）：

```go
func onResize(size string, e *core.Engine, w *fyne.Window) {
    btn.Disable()
    prog.Show()
    go func() {
        err := e.ChangeSwapSize(size) // 方法名/簽名以 core/swap.go 為準
        fyne.Do(func() {
            prog.Hide()
            btn.Enable()
            refreshStatus()           // 刷新狀態區
            if err != nil { dialog.ShowError(err, w) }
        })
    }()
}
```

- **`e.SetProgressCallback(...)` 只在啟動時掛一次**（Fyne 版回呼，內部用 `fyne.Do` 更新 progress bar 數值）。
  注意 Web 模式用的 SSE broadcast 機制（`core/progress.go`）在 native 模式**不使用**。
- 狀態自動刷新：**5 秒 ticker + 每次操作後 + 手動 Refresh 按鈕**（三者皆有，[DECIDED]）。

### 2.4 Sudo 流程

- 啟動時 `os.Geteuid() != 0 && e.Password == ""` → 顯示密碼對話框（`widget.NewPasswordEntry()`）：
  - 輸入 → `e.TestAuth(pw)` 驗證（方法簽名以 `core/sudo.go` 為準）→ 成功：`e.Password = pw`
  - 可「略過」→ 頁頂顯示**警告 banner**（需要權限的行動會失敗），不擋只讀功能
- 操作中 `RenewAuth()` 回錯（密碼錯 / 已過期）→ **重彈對話框**。
- 密碼只存 `e.Password`（記憶體），**不落磁碟**（決策 8）。
- CLI 以 `sudo` 執行時 euid==0 的路徑由 `core/sudo.go` 處理，UI 不需理會。

> **附註（已知缺口，記錄用）**：P0.5 已完成（2026-08）：Web 模式的密碼輸入 UI 與
> `cmd/desktop/api.go` 的 `POST /api/auth` 端點**皆已移除**，`core/sudo.go` 的
> `RenewAuth()` 新增 `os.Geteuid() == 0` 跳過 → web 模式以 root 執行時正常；
> 非 root web 執行時 `e.Password` 恒為空、`RenewAuth()` 會失敗（需權限操作會報錯）。
> 本 phase **不修 web**（native UI 有自己的對話框，不受影響）；
> 是否給 web 補回輸入是**獨立議題**，等用戶指示。

### 2.5 引擎方法對照（實作時的捷徑）

`cmd/desktop/api.go` 是現成的「**動作 → engine 方法**」對照表：每個 HTTP handler 內部
就是對應的 `core.Engine` 呼叫（含簽名與參數）。Fyne 各 section 照同一張表接線即可，
**不要自行猜方法名**。ZRAM 三個欄位直接在 `GetStatusSummary()` 回傳結構裡。

---

## 3. i18n（共享單一來源 + drop-in，用戶要求的核心功能）

### 3.1 檔案佈局 [DECIDED]

```
i18n/locales/*.json                    ← **canonical 單一來源**（git mv 自 web/src/locales/）
web/src/locales/*.json                 ← 生成物（build 前由 sync-loc 複製過來）→ 加入 .gitignore
dist/locales/*.json                    ← 既有：web runtime fetch（機制不變）
~/.cryoutils_ng/locales/*.json         ← **Fyne runtime drop-in**（新語言免重編）
```

- **web 端改動（最小化，src/ 程式碼零改動）**：
  - `package.json`：`"build": "npm run sync-loc && tsc && vite build && (mkdir -p dist/locales && cp src/locales/*.json dist/locales/)"`
  - 新增 `"sync-loc": "cp ../i18n/locales/*.json src/locales/"`
  - `import.meta.glob('/src/locales/*.json', { eager: true })` 維持不變（build 時生成物已存在）
  - `LOCALES_CODES`（`web/src/i18n/locales.ts`）仍是 web 端的語言白名單（web 端要新增語言需在此加一碼）
- **Go 端**：`i18n` package `//go:embed locales/*.json`；啟動時**另外掃描**
  `filepath.Join(core.InstallDirectory, "locales", "*.json")`，檔案 stem = 語言 code，
  **同 code 時 runtime 檔覆蓋 built-in**；JSON 解析失敗的檔跳過 + 記 log（不 crash）。
- **web runtime drop-in 不做**（web dist 是 embed 進二進位的；用戶的 drop-in 要求針對 Fyne UI）。

### 3.2 `i18n` package API

```go
package i18n

// Available 回傳所有可用語言 code（built-in + ~/.cryoutils_ng/locales/，去重排序，en 在前）
func Available() []string

// Load 載入指定語言（built-in 或 runtime；找不到回 nil）
func Load(code string) *Lang

// T dot-notation 查詢："status.recommended" → 值。
// fallback 鏈：所选語言 → en built-in → 回傳 key 本身（與 web resolveKey 行為一致）
func (l *Lang) T(key string) string

// Label 語言顯示名（JSON 內的 "_label" 欄位；缺省用 code）
func (l *Lang) Label() string
```

- JSON schema 沿用 web 既有格式：`_label` + 巢狀 section（`header.*`、`status.*`、`swap.*`、
  `memory.*`、`vram.*`、`presets.*`、`gamedata.*`、`memoryParams.*`、`common.*`）。
- **Fyne 專屬字串放新 section `native.*`**（sudo 對話框、語言選單、警告 banner 等）。
  web 端對未知 section 不敏感（safe）。
- **Fyne UI 字串避免 emoji**（`✓` `✗` 可用；`🔒` 等圖形 emoji Noto Sans TC 未必有 glyph）。
  `native.*` 鍵值不要複製 web 的 `header.locked = "🔒 Locked"` 那類。

### 3.3 社群新增語言流程（零重編）

**Fyne UI**：把 `xx.json`（含 `_label`）放入 `~/.cryoutils_ng/locales/` → 重新啟動 app（或重開語言選單）→ 自動出現在選單。**零程式碼、零重編。**

**Web UI**：`xx.json` 放入 `i18n/locales/` + 在 `web/src/i18n/locales.ts` 的 `LOCALES_CODES` 加 code → rebuild。

**Repo 正式新增語言**：`i18n/locales/xx.json` + 兩端各一行（web 白名單；Fyne 自動）。

---

## 4. Fyne 實作細節

| 項目 | 做法 |
|------|------|
| App | `fyne.NewAppWithID("io.cryoutils-ng")`（App ID 沿用 AGENTS.md naming） |
| 主題 | `app.SetTheme(theme.DarkTheme())` |
| 字體 | `//go:embed fonts/NotoSansTC-Regular.ttf fonts/NotoSansTC-Bold.ttf` → `fyne.NewStaticResource` → `fyne.SetFontWithAlias(res, fyne.TextFont{...})`（v2.5+ 字體 API；**2.7.4 確切簽名照官方 docs 對**，2.5 起 API 有過一次重構）。Bold 對應 `fyne.TextFontBold` style。備援：custom theme 設 `theme.Font` |
| 字體來源 | `https://github.com/google/fonts/raw/main/ofl/notosanstc/NotoSansTC-Regular.ttf`（及 `-Bold.ttf`）；OFL 授權，連 `OFL.txt` 一起放入 `fonts/`（GPLv3 + OFL 相容）。**下載後驗證檔有效（`file` 指令 + 實際渲染）**；repo 路徑有變就換 notofonts/noto-cjk |
| Window | `w.Resize(fyne.NewSize(1280, 800))`（Deck 原生解析度）、resizable、`w.ShowAndRun()` |
| 單頁結構 | `container.NewVBox(header, progressRow, sc)`，其中 `sc := container.NewScroll(contentBox)`；progressRow 預設 height 0 / 隱藏 |
| 語言切換 | 選單 `onChange` → 存 `ui.json` → **重建整個 contentBox**（整頁重繪，最簡單可靠） |
| 語言持久化 | `~/.cryoutils_ng/ui.json`：`{"language":"zh-TW"}`；不存在時預設 `en` |
| 4K | Fyne 跟隨系統 DPI 縮放；**接受條件**：3840×2160 外接時文字/控件清晰不溢出（真實 Deck 驗） |
| 確認對話框 | `dialog.NewConfirm`（resize 前、presets 前）；錯誤 `dialog.ShowError`；成功可 `dialog.ShowInformation` 或不彈（狀態刷新即可） |
| 警告 banner | 一個 `widget.NewRichText`（紅字），sudo 未設定時顯示 |
| 單實例鎖 | **範圍外**（與既有決策一致） |

**Fyne widget 對照表**（常用）：`widget.NewLabel` / `NewRichTextWithText` / `NewSelect(options, cb)` /
`NewCheck(text, cb)` / `NewPasswordEntry` / `NewButton(text, cb)` /
`widget.NewProgressBar()`（indeterminate）+ `widget.NewProgressBarMinMax` /
`container.NewVBox` / `NewGridWrap` / `NewScroll` / `dialog.NewConfirm` / `dialog.ShowError`。

---

## 5. 施工步驟（每步有 gate；依序執行）

> 所有 go 指令用 `/usr/local/go/bin/go`。**每步完成先自測再進下一步；任何 gate 失敗 → 停下回報用戶**（不要自行換路線）。

### Step 0 — Dev VM 準備
```bash
df -h /                                   # 強制規則：< 3GB 就停
apt-get install -y libgl1-mesa-dri scrot mesa-utils
Xvfb :99 -screen 0 1280x800x24 &
DISPLAY=:99 glxinfo | grep -i "opengl renderer"   # 預期：llvmpipe 或 softpipe
```
**gate**：glxinfo 顯示 llvmpipe/softpipe。

### Step 1 — Fyne 升級 gate（v2.3.1 → v2.7.4）
```bash
cd /root/opencode-stuffs/steam-deck-utilities
/usr/local/go/bin/go get fyne.io/fyne/v2@v2.7.4
/usr/local/go/bin/go mod tidy
CGO_ENABLED=1 /usr/local/go/bin/go build ./internal/        # 舊 UI 必須仍可編譯
CGO_ENABLED=0 /usr/local/go/bin/go build ./core/            # core 不受影響
CGO_ENABLED=0 /usr/local/go/bin/go build -o /tmp/cli ./cmd/cryoutilities
```
寫 `/tmp/fynehello/main.go`（獨立 module `go mod init fynehello` + 同版 Fyne）：
一個 window + 一個 button + 一個含 `「繁體中文測試」` 的 label。
```bash
cd /tmp/fynehello && CGO_ENABLED=1 /usr/local/go/bin/go build -o hello .
Xvfb :99 -screen 0 800x600x24 &
DISPLAY=:99 ./hello &
sleep 3 && DISPLAY=:99 scrot /tmp/fynehello.png
```
**gate（go/no-go）**：截圖顯示 window 有渲染。中文字是方框/缺字**符合預期**（字體是 Step 3 的事）；但 **window 完全黑/無法啟動 = native 路線 go/no-go 失敗 → 停、回報、重新評估**。
（gate 通過後**可選**：`go get fyne.io/fyne/v2@v2.8.0` 重測；有異常就回 v2.7.4 並記錄。）

### Step 2 — i18n 單一來源
1. `mkdir i18n && git mv web/src/locales i18n/locales`
2. 寫 `i18n/i18n.go`（§3.2 API）+ `i18n/i18n_test.go`（fallback 鏈、runtime 覆蓋、T 缺鍵、壞 JSON 跳過）
3. `web/package.json` 加 `sync-loc` + 改 `build`（§3.1）
4. `i18n/locales/` 補 `native.*` 新鍵（en + zh-TW 都補）
5. 確認 `.gitignore` 有 `web/src/locales/`（**本次更新文件時已加**；Step 2 實際生效）

**gate**：`cd i18n && CGO_ENABLED=0 /usr/local/go/bin/go test ./...` ✅ + `cd web && npm run build` ✅（dist/locales 齊全、glob 抓到 generated 檔）+ `tsc --noEmit` ✅。

### Step 3 — UI 骨架 + 字體（font gate）
1. `ui/fyneui/app.go`：§4 全部（app、dark theme、window、VBox + Scroll、header、progressRow 隱碼、5s ticker 佔位）
2. 下載 Noto Sans TC → `ui/fyneui/fonts/` + `fonts.go` embed + `SetFontWithAlias`
3. 暫時在 contentBox 放一行 `「CryoUtils NG 繁體中文渲染測試 ✓」`
4. `cmd/desktop/main.go` 加 `-ui` flag（native 分支先只跑骨架）
```bash
cd web && npm run build && cp -r dist ../cmd/desktop/web/dist    # embed 需要 dist
cd ../cmd/desktop && CGO_ENABLED=1 /usr/local/go/bin/go build -o /tmp/cryoutils-ng-desktop .
Xvfb :99 -screen 0 1280x800x24 &
DISPLAY=:99 /tmp/cryoutils-ng-desktop -ui native &
sleep 4 && DISPLAY=:99 scrot /tmp/fyne-skeleton.png
```
**gate（font gate）**：截圖 = 暗色視窗 + **中文字正常顯示**（非方框）。

> **2026-08-25 字體來源定案**：`notofonts/notofonts.github.io` 不含 CJK（無 NotoSansTC）。改用 `google/fonts` 的變數字型 `ofl/notosanstc/NotoSansTC[wght].ttf`，以 `python3-fonttools varLib.instancer` 分別固定 `wght=400/700` 產出靜態 Regular/Bold TTF（各 ~7MB，Fyne 需兩檔分開對應 TextStyle），連同 `OFL.txt` 放 `ui/fyneui/fonts/`。⚠️ 網路上流傳的部分 NotoSansTC 檔是 **TTC collection**，Fyne `ParseTTF` 會報 `collections not allowed` 並 panic——下載後務必 `file` 驗證是單一 TrueType。

### Step 4 — Status section（含 ZRAM）
- `sections_status.go`：接 `e.GetStatusSummary()` 全部欄位（§2.2 區塊 3）
- 5s ticker + Refresh 按鈕 + `refreshStatus()` 公用函式（其他 section 共用）
**gate**：xvfb 截圖顯示全部狀態值（Dev VM 無 zram → ZRAM Status 顯示 Inactive 即正確行為）。

### Step 5 — Swap section
- size 下拉 + 現值、Resize 確認對話框、goroutine + `fyne.Do`（§2.3 模板）、progress bar
- resize 前警告文案：swap 會短暫停用（含 zram 說明，措辭參照 web `swap.zramNote`）
**gate**：Dev VM 上以測試用小 swap file 實跑一次 resize（或至少確認對話框/progress 流程 + 錯誤路徑）。

### Step 6 — Memory / Presets / VRAM / Game Data
- 照 §2.2 區塊 5–8 接線；長任務全部走 §2.3 模板
**gate**：`go vet` 通過 + xvfb 截圖各區塊渲染正確（en + zh-TW 各一張）。

### Step 7 — 語言選單 + ui.json
- `lang.go`：`i18n.Available()` → Select；onChange → 存 `ui.json` → 重建 contentBox
- 啟動讀 `ui.json`
**gate**：切換 en↔zh-TW 全頁即時重繪；重啟後語言保留；**把一份 `fr.json` 放 `~/.cryoutils_ng/locales/` 重啟 → 選單出現第三語言**（drop-in 验收）。

### Step 8 — sudo 對話框
- §2.4 全部（首啟、略過 + banner、失敗重彈）
**gate**：非 root 啟動 → 對話框出現；錯誤密碼被拒；正確密碼後操作可用；略過後 banner 顯示。

### Step 9 — `cmd/desktop` 收斂
- `-ui` flag 完整（web 預設）；web 模式**回歸測試**（`-no-browser` 起 server → `curl /api/status` 正常、瀏覽器啟動邏輯不變）
- native 模式確認**不**開 port（`ss -ltn` 無 127.0.0.1 監聽）
**gate**：兩模式各跑一次冒煙。

### Step 10 — Headless 全面驗證
- 1280×800 與 2560×1440 兩種視窗尺寸 × en/zh-TW × （含 progress 顯示狀態）截圖
- 截圖放 `/tmp`（**不 commit**）
**gate**：視覺檢查無溢出/重疊/缺字；報告貼給用戶。

> **2026-08-25 實測筆記（headless 驗證方法論）**：
> - `w.Canvas().Capture()` 在 Xvfb+llvmpipe 下讀 front buffer 恆黑（swap 不翻 front）→ **截圖一律用 `scrot`**。
> - 最小 hello app 在此環境**不呈現**（純黑視窗；glxgears 正常）；真實 app 反而正常呈現——原因未明，不影響本專案，真機 Step 11 再確認。
> - 除錯工具：`CRYOUTILS_UI_SHOT=/path.png`（自動截圖後離開；注意它內部走 Capture，在此環境會是黑的，僅供其他平台用）、`CRYOUTILS_UI_DEBUG=1`（4 秒後印各 section layout 尺寸到 stdout）。
> - OCR 驗證管線：`scrot` → PIL 反相+放大 2x → `tesseract`（`-l eng` / `-l chi_tra+eng`）→ 檢查實際渲染文字。中文字非方框 = font gate 通過。
> - 教訓：`container.NewVScroll(...)` **不可**再包進 `NewVBox` 給 Border 當 center——VScroll 会被壓成 MinSize 高度（32px），scroll 直接當 Border center 即可。

### Step 11 — 真實 Steam Deck（用戶在場）
1. Dev VM build（CGO_ENABLED=1）→ 用戶在 deck 端 `scp root@192.168.1.15:... ~/.cryoutils_ng/cryoutils-ng-desktop`
2. Desktop Mode 啟動 `-ui native`；驗：1280×800、4K 外接、`.desktop` 圖標啟動
3. 實測：status 數值、zh-TW、resize（zram 恢復）、sudo 對話框、語言 drop-in
4. 確認原始 CryoUtilities 未被動過（`~/.cryo_utilities/` 完整）
5. glibc/GL 若報缺庫 → 回報缺失庫名（SteamOS 應有：mesa、xorg、fontconfig）
**gate**：用戶視覺 + 功能接受。

### Step 12 — 打包 / CI / 文件
- `install.sh`：二進位路徑不變；**README + `docs/manual-install.md` 補 desktop 二進位的系統相依**（glibc、libGL/mesa、X11、fontconfig — SteamOS 皆內含，其他 Linux 需自裝）
- `launcher.sh`：支援 `-ui` 傳遞（`$@` 已支援，確認即可）
- `.desktop`：entry 不變（同一二進位）
- `.github/workflows/release.yml`：**確認是否 build desktop 二進位**；是 → CI 需 `sudo apt-get install -y libgl1-mesa-dev xorg-dev` + `CGO_ENABLED=1`（Phase 0.5 曾移除 Fyne apt deps，要加回來）；CLI 維持 CGO=0
- 更新 `todo.md` / `AGENTS.md` / `README.md`（native UI 成為主要使用方式；web 標為 fallback + Decky 用途）
- 版本號：`v0.1.0` → **`v0.2.0`**（[待用戶確認]，§8）

### Step 13 — 刪除 `internal/`
- 確認 Step 11 gate 通過後：`git rm -r internal/`（**單獨 commit**，訊息記錄原因）
- root module `go mod tidy`（Fyne 保留 — 新 UI 在用）
- `go build ./...`（CGO=1）+ core/CLI（CGO=0）全綠

### Step 14 — 預設值切換（可選，用戶指示後）
- `-ui` 預設 `web` → `native`（一個字串 + 文件更新）

---

## 6. 最終接受條件（Definition of Done）

- [ ] dark theme；單頁直式無 tab；與 Web UI 六區塊同等（含 ZRAM 四欄）
- [ ] `-ui web|native`；web 模式零回歸；native 模式無 HTTP/token
- [ ] i18n：en/zh-TW built-in + `~/.cryoutils_ng/locales/` drop-in 免重編 + 切換即時重繪 + 持久化
- [ ] sudo 對話框（首啟/失敗/略過 + banner）；密碼不落磁碟
- [ ] 1280×800 與 3840×2160 渲染正確（真實 Deck）
- [ ] `cryoutils-ng` CLI 仍 `CGO_ENABLED=0` 靜態、行為不變
- [ ] 原始 CryoUtilities 安裝未被動
- [ ] `core/` 零改動（或任何改動有獨立理由與測試）
- [ ] `go vet` 全綠（core、i18n、ui/fyneui、cmd/*）
- [ ] internal/ 已刪除（Step 13）

---

## 7. 風險與 fallback

| 風險 | 處理 |
|------|------|
| v2.7.4 與預期 API 不符（`SetFontWithAlias`、`fyne.Do` 簽名等） | 照 v2.7.4 官方 docs 調整；屬 moderate 修正，不影響架構 |
| Xvfb + llvmpipe 渲染失敗 | Step 1 是 go/no-go gate：**停下回報**，由用戶決定是否放棄 native 路線（不自行換 Gio） |
| SteamOS glibc/GL 不相容 | Step 11 驗；glibc 太新 → 在 Arch container 建置；缺 GL 庫 → 回報（SteamOS 有 mesa/xorg） |
| Noto Sans TC 個別字缺 glyph（罕見罕用字） | 接受（罕見字顯示 fallback）；或補第二套字體（超出本期） |
| Fyne 2.5+ threading 模型細節 | 一律 `fyne.Do`；寫完跑 `go vet` + 實跑觀察 data race（可選 `-race` build 冒煙） |
| 二進位變大（25–40MB） | 接受（Releases 發佈，原 web-only 二進位也含 web dist） |

---

## 8. 未決事項（需用戶確認；其餘全部 DECIDED）

1. **版本號**：Phase 6.5 接受後 `v0.1.0` → `v0.2.0`（建議；semver minor，新增 major feature）。等用戶點頭才動。
2. **`-ui` 預設值何時切 `native`**：本文件建議 = Step 11 接受後（Step 14）。用戶若想更早/更晚，指示即可。

---

## 附：關鍵檔案索引

| 用途 | 路徑 |
|------|------|
| 本施工手冊 | `docs/fyne-ui-plan.md` |
| Phase 清單 | `todo.md`（Phase 6.5 節） |
| 專案規則 | `AGENTS.md` |
| 引擎（單一事實來源） | `core/`（`engine.go`、`swap.go`、`sudo.go`、`memory.go`、`gpu.go`、`gamedata.go`、`steamapi.go`、`library.go`、`progress.go`） |
| 動作→engine 方法對照 | `cmd/desktop/api.go` |
| 舊 Fyne UI（基準，Step 13 前保留） | `internal/` |
| web i18n 機制（參考，src/ 不改） | `web/src/i18n/index.tsx`、`web/src/i18n/locales.ts` |
| locale canonical（Step 2 後） | `i18n/locales/` |
| 新 UI | `ui/fyneui/`（Step 3 起） |
| 既有 UI 設計決策 | `AGENTS.md` §UI Design Decisions / §Desktop UI Launch |
