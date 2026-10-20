# CryoUtils NG — Todo

## Phase 0.5: CI/CD Infrastructure ✅
- [x] `.github/dependabot.yml` — gomod weekly, npm weekly, ignore npm major, 5 PR limit
- [x] `.github/workflows/dependabot-auto-merge.yml` — auto-merge patch/minor, major needs review
- [x] `.github/workflows/release.yml` — updated actions, removed Fyne apt deps, fixed binary path
- [x] GitHub UI: branch protection rulesets for `main` and `develop` (Require PR, Require status checks, Block force pushes)
- [x] GitHub UI: enable Dependabot security updates

## Phase 0: Dependency Modernization & Baseline ✅
- [x] Go 1.26.5 verified at `/usr/local/go/bin/go`
- [x] `core/` created as independent Go module (`go 1.26`, 5 deps, no Fyne, no CGO)
- [x] Root `go.mod` updated: go 1.26, updated deps, `replace cryoutils-ng/core => ./core`
- [x] Fyne retained in root module (internal/ still imports it)
- [x] `go build ./...` ✅ · `go vet ./...` ✅ · core `CGO_ENABLED=0` ✅
- [x] Root module `CGO_ENABLED=0` build ✅ (verified after Phase 1)

## Phase 1: Core Extraction ✅
- [x] `core/config.go` — all constants migrated from `internal/config.go`
- [x] `core/util.go` — utility functions migrated, Fyne removed, `removeFile` → Engine method
- [x] `core/sudo.go` — `renewAuth()` and `TestAuth()` methods on Engine
- [x] `core/engine.go` — `Engine` struct with `InfoLog`, `ErrorLog`, `Password`, `OnProgress`, `SteamAPIResponse`, `SwapFileLocation`; `UseRecommendedSettings()`, `UseStockSettings()`, `NewEngine()`
- [x] `core/swap.go` — all swap handlers as Engine methods
- [x] `core/memory.go` — all memory handlers as Engine methods (Set/Revert/Toggle/Status)
- [x] `core/gpu.go` — `GetVRAMValue()` as Engine method
- [x] `core/gamedata.go` — game data handlers as Engine methods
- [x] `core/library.go` — Library struct, `FindDataFolders()`, `ParseVDF()`
- [x] `core/steamapi.go` — `AppResponse`, `QuerySteamAPI()`, `GenerateGameMap()`
- [x] `core/util_test.go` — unit tests ported and passing
- [x] `cmd/cryoutilities/main.go` — updated to use `core.Engine`, added `status` command
- [x] `go build ./...` ✅ · `go vet ./...` ✅ · core `CGO_ENABLED=0` ✅ · CLI `CGO_ENABLED=0` ✅
- [x] `go test ./...` ✅ (all passing)
- [x] Delete `internal/` — deferred until Phase 4 (full UI rewrite removes all Fyne references)

## Phase 2: CLI Rework + `status` Command ✅
- [x] `status` command added to CLI (prints all tuning statuses)
- [x] All CLI subcommands use `core.Engine` methods
- [x] CLI smoke test: `status` ✅, `swappiness 60` ✅, `help` ✅

## Phase 3: Desktop Web Server ✅
- [x] `cmd/desktop/main.go` — localhost web server (127.0.0.1 + random token)
- [x] REST API endpoints: `GET /api/status`, `POST /api/swap/resize`, `POST /api/swap/swappiness`, `POST /api/memory/{name}`, `POST /api/recommended`, `POST /api/stock`, `POST /api/gamedata/sync`, `POST /api/gamedata/cleanup`, `GET /api/libraries`
- [x] SSE progress endpoint (`/api/progress`) with keepalive
- [x] `go:embed` of web build → single-file install
- [x] Token auth on all privileged endpoints
- [x] `core/progress.go` — broadcast progress channel
- [x] Exported `RenewAuth()`, `RemoveGameData()`, `DataToMove.GetRight()/GetLeft()`
- [x] `CGO_ENABLED=0` static build ✅
- [x] `go vet ./cmd/desktop/...` ✅

## Phase 4: Single-Page React UI ✅
- [x] `web/` — React + Vite + TypeScript single-page UI
- [x] Vertical single column, no tabs
- [x] Header inline sudo unlock
- [x] Responsive: 3840×2160 (4K) and 1280×800 (Steam Deck native)
- [x] Plain CSS, framework-agnostic components
- [x] Headless browser verification (Playwright) at both resolutions

## Phase 5: Packaging ✅
- [x] `install.sh` — installs both CLI + desktop binaries to `~/.cryoutils_ng/`, creates `.desktop` files
- [x] `uninstall.sh` — removes install artifacts, preserves data
- [x] `launcher.sh` — wrapper for Steam Big Picture mode (calls `cryoutils-ng-desktop`)
- [x] `.desktop` file — `CryoUtilsNG` entry in Steam Deck gamemode
- [x] Coexistence: never overwrite original CryoUtilities install paths
- [x] `.github/workflows/release.yml` — fixed to build web UI + both binaries + checksums
- [x] `README.md` — updated with desktop server instructions and two-binary install flow
- [x] `docs/manual-install.md` — updated for two binaries

## Phase 5.6: i18n 多國語言支援 + 版本號修正 ✅
- [x] 版本號 `v2.2.2` → `v0.1.0`（重寫專案，semver pre-release）
- [x] 建立 `web/src/locales/en.json` + `zh-TW.json` 翻譯檔（編譯後複製到 `dist/locales/`）
- [x] `web/src/i18n/locales.ts` — **單一來源** locale 代碼白名單（`LOCALES_CODES` + `Locale` 類型）
- [x] `web/src/i18n/index.tsx` — `import.meta.glob` 自動掃描 + Context provider + localStorage 持久化
- [x] `web/src/LanguageSelector.tsx` — 語言選單從 `LOCALES` 自動生成
- [x] `web/src/types.ts` — `MEMORY_PARAM_LABELS` → `MEMORY_PARAM_KEYS`（translation key 化）
- [x] `web/src/App.tsx` — 所有硬編碼字串替換為 `t.*` 翻譯呼叫
- [x] `web/src/main.tsx` — 包覆 `I18nProvider`
- [x] `web/src/App.css` — `.locale-select` 樣式
- [x] `tsc --noEmit` ✅ · `npm run build` ✅
- [ ] 後續添加更多語言（日語、簡體中文等）— 需同步新增 zram 翻譯鍵
- [ ] Playwright 截圖測試更新（zh-TW 截圖）

## Phase 5.5: Desktop UI — 無邊框視窗模式 (方案 D) ✅
**使用環境**:UI 主要在 Steam Deck **Desktop Mode**(Gaming Mode 非主要)。
- [x] `cmd/desktop/browser.go` — `findAppBrowsers()`(lookpath 候選:google-chrome / chromium / chromium-browser / microsoft-edge / brave-browser / brave / firefox)
- [x] Flatpak 偵測 fallback(`flatpak list --app` 匹配 Brave/Chromium/Chrome/Edge,用 `flatpak run`)
- [x] `openInAppWindow(url)` — `--new-window --app=<url>` 開無邊框獨立視窗
- [x] `steam://openurl` fallback(Steam 在 Deck 保證存在;原廠無瀏覽器時用其內建 CEF 瀏覽器)
- [x] `openInBrowser(url)` — `xdg-open` fallback
- [x] Flags:`-no-browser`(只印 URL)、`-browser <path>`(強制指定)
- [x] `main.go` 接線:app-window 成功才開,否則 fallback;log 記錄啟用方式
- [x] `browser_test.go` — 偵測順序 / fallback / override 單元測試(`t.Setenv("PATH", ...)`)
- [x] 驗證:`CGO_ENABLED=0 go vet ./cmd/desktop/...` ✅、build ✅、dev VM `-no-browser` 冒煙測試 ✅
- [x] README 更新(開窗方式 + flags + Desktop Mode 使用說明)
- [x] `launcher.sh` 更新(支援 `$@` 傳遞 flags)
- [ ] 真實 Deck 確認 `--app=`、flatpak、`steam://openurl` 參數傳遞與視窗行為(Phase 6)

**Deck 測試結果 (2026-08-11)**:
- ✅ Desktop server 啟動成功，印出 URL
- ✅ Web UI 可正常顯示數值（VRAM、swappiness 等）
- ❌ Flatpak Chrome `--app=` 啟動但有 `blink.mojom.Widget` 錯誤，顯示拒絕連線
- ❌ Flatpak Brave `--app=` 啟動但顯示拒絕連線
- ⚠️ zsh globbing 問題：URL 中的 `=` 需加單引號 `'url'`
- ⏳ `--app=` 在 SteamOS Flatpak 中的行為尚未確認（需 Valve CEF 測試）

**範圍外**:單實例鎖、Firefox 支援(走 xdg-open fallback)

## Phase 6: 統合真機驗證（user — requires real Steam Deck）

> **重排（2026-08）**：原本排在 Phase 4/5 後；延後到 **Phase 6.5 完成後**一次做統合真機驗證（native UI 是主要路徑；web `--app=` 已知真機失敗，由 native 取代）。

- [x] Real Steam Deck acceptance testing (2026-08-11, CLI + Web UI basic smoke)
- [x] Verify original CryoUtilities untouched (confirmed at `~/.cryo_utilities/`)
- [x] CLI `status` command verified on real hardware (`SwapSizeGB: 16` ✅)
- [x] 真實 Deck 驗證 zram 狀態顯示與 resize 後 zram 不再失效（2026-10-20，方法 A binary）：四次 resize 後 `/dev/zram0` 仍在、Priority 仍 100、`zramctl` zstd/7.2G、log 零次提及 zram、journal 無新 `--reset-device`。（同一項在舊 binary `425bc04` 上首次真機驗是**失敗**的，根因與修法見 `docs/zram-swapoff-reset.md`）
- [ ] All CLI commands verified on real hardware（`swappiness`、`recommended`、`stock`、`hugepages` 等，含 P0.5 後 `sudo` 執行 CLI 的 `Geteuid` 跳過路徑） → 部分已完成（2026-10-20）：`status` 與 `swap 2/8/16/12` 已真機驗；其餘待驗
- [ ] Web UI 視覺驗收（fallback 路徑，與 native 一次驗）
- [ ] `steam://openurl` fallback 確認（次要；web 模式無瀏覽器可用時才走）
- [ ] Flatpak zsh globbing 問題記錄到 README（URL 需加引號）
- ~~Web UI `--app=` 視窗驗證~~ → **移除**：`--app=` 真機失敗已確認（見 Phase 5.5 Deck 測試結果），由 Fyne 原生 UI 取代

## Phase 6.5: Fyne 原生 UI（新）

> 完成後接 Phase 6 統合真機驗證。

> **施工手冊：`docs/fyne-ui-plan.md`**（全部決策已定案，標 [DECIDED] 的不要再評估）。
> 方向（2026-08-21 定案）：全新撰寫 **Fyne v2.7.4** 原生 UI（**不沿用 `internal/` 舊 UI**）；web UI 保留（Decky + fallback）；`cmd/desktop -ui web|native`（預設 web）。

### 6.5.0 PoC Gate
- [x] Dev VM：`apt-get install libgl1-mesa-dri scrot mesa-utils`；Xvfb 下確認 llvmpipe/softpipe（2026-08-25：glxinfo 確認 llvmpipe LLVM 20.1.2、GLX direct rendering）
- [x] Root module Fyne v2.3.1 → **v2.7.4**；`CGO_ENABLED=1 go build ./internal/` 仍可過（2026-08-25 重驗 ✅）
- [x] Fyne hello xvfb 渲染 gate（**go/no-go**）（2026-08-25）：最小 hello 在 Xvfb+llvmpipe 下**不呈現**（已知環境怪癖，見 plan §5 註記），但**真實 app 呈現正常**且 OCR 驗證 en/zh-TW 全頁文字 → native 路線判定 **go**
- [ ] （可選）v2.8.0 試測；有問題維持 v2.7.4 並記錄

### 6.5.1 i18n 單一來源
- [x] `git mv web/src/locales i18n/locales`（canonical）
- [x] `i18n/` package：`//go:embed` + `Available/Load/T` + runtime `~/.cryoutils_ng/locales/` 掃描覆蓋
- [x] i18n 單元測試（fallback / override / 缺鍵 / 壞 JSON 跳過）
- [x] web `package.json`：`sync-loc` + build pipeline；`.gitignore` 加 `web/src/locales/`（已加，本 phase 生效）
- [x] `native.*` 新鍵補 en + zh-TW；web `src/` 零改動

### 6.5.2 Fyne UI（dark、單頁無 tab、與 web 六區塊同等）
- [x] `ui/fyneui/` 骨架：app / dark theme / window / Noto Sans TC 嵌入（**font gate**）（2026-08-25：TTC collection 換成 google/fonts 變數字型 instancer 產出的靜態 Regular/Bold TTF + OFL.txt；OCR 驗證中文字正常非方框）
- [x] Status section（含 ZRAM Size / ZRAM Status / Total Swap）+ 5s 自動刷新 + Refresh 鈕
- [x] Swap section（size 選單 / 確認 / progress / zram 警告文案）
- [x] Memory / Presets / VRAM / Game Data sections
- [x] 語言選單 + `ui.json` 持久化 + 切換即時全頁重繪（2026-08-25 OCR 驗證 en↔zh-TW；drop-in `~/.cryoutils_ng/locales/` 覆蓋路徑待測）
- [ ] sudo 密碼對話框（首啟 / 失敗重彈 / 略過 + banner；只存記憶體）——root 下自動跳過，headless 無法測，留待真機
- [x] 執行緒模型：UI 更新一律 `fyne.Do`；長任務 goroutine（plan §2.3 模板）

### 6.5.3 整合與驗證
- [x] `cmd/desktop` `-ui web|native`（預設 web；native 不啟 HTTP/token/瀏覽器）
- [ ] web 模式零回歸（`/api/status`、瀏覽器啟動鏈）
- [ ] Headless 截圖（1280×800 ✅ 2026-08-25 en/zh-TW via Xvfb+scrot+OCR；2560×1440 待做）
- [ ] 真實 Steam Deck：1280×800 / 4K / .desktop / 全功能 / 原始 CryoUtilities 未動
- [ ] 打包：install.sh / manual-install.md（CGO 運行期相依）/ release.yml（若 build desktop → apt GL/X11 deps + `CGO_ENABLED=1`）
- [x] `go vet` 全綠；core/CLI 維持 `CGO_ENABLED=0`（2026-08-25：desktop `CGO_ENABLED=1` vet 乾淨；core 測試全過）

### 6.5.4 收尾
- [ ] 刪除 `internal/`（**單獨 commit**；Fyne 依賴保留）
- [x] 文件更新：AGENTS.md / README.md / todo.md
- [ ] 版本號 `v0.1.0` → `v0.2.0`（待用戶確認）
- [ ] （可選，接受後）`-ui` 預設值切 `native`

## Phase 6.6: Native UI 4K 適配與 sudo 流程修正（2026-09-11 真機 feedback）

> 由來：真實 Deck 4K（3840×2160）測試 — 字太小、元件被拉滿 3840px 寬、sudo 密碼被問多次。
> 施工順序（用戶確認 2026-09-11）：6.6.1 → 6.6.2 → 6.6.3 → 6.6.4。

### 6.6.1 sudo 雙問修正（native）
- [x] 保留啟動密碼對話框（**用戶決策 2026-09-11**：部分狀態讀取走 `sudo cat`（`core/util.go`），非 root 時很多狀態看不到 → **不做** pure lazy 問密碼）
- [x] 防對話框疊開：`authManager` 加 `dialogOpen` flag + `pending` 佇列（啟動對話框延遲 200ms，若用戶先點操作 → 佇列，對話框解決後 `flushPending` 重入 `ensure`）（2026-09-11 dev 完成）
- [x] 刪死代碼 `ui/fyneui/sudo_dialog.go`（`askPassword`/`withAuth` 零呼叫點；實際路徑是 `sudo.go` 的 `authManager`）（2026-09-11 已刪）
- [x] 密碼維持只存記憶體（`e.Password`，不落磁碟）— 不變
- [ ] 驗證（真機）：啟動輸入一次後，後續所有操作不再問；快速連點操作不疊開兩個對話框

### 6.6.2 FYNE_SCALE 自適應
- [x] Fyne v2.7.4 **無運行時 `SetScale()` API**（已查證 `Settings` interface 只有 `Scale()` getter）→ 啟動前自 re-exec：`cmd/desktop/scale.go` `maybeReexecWithScale()`，native 分支啟動時（Fyne app 建立前）`FYNE_SCALE` 與 `CRYOUTILS_FYNE_SCALE` 皆未設 → 讀 `xrandr` 解析度 → scale = clamp(寬/1280 四捨五入到 0.25 倍數, 1.0, 2.0) → 帶 env var `syscall.Exec` 重啟自己（變數已設即跳過，防迴圈）（2026-09-11 dev 完成）
- [x] 映射：1280×800→1.0 · 1920×1080→1.5 · 2560×1440→2.0 · 3840×2160→2.0（上限）；**scale==1.0 時直接跳過 re-exec**（原生解析度免無謂 process restart）（2026-09-11 dev）
- [x] 手動覆蓋：`FYNE_SCALE` env var 優先（與 Fyne 既有行為相容）；**re-exec 同時設 `FYNE_SCALE`（Fyne 實際讀取的變數）與 `CRYOUTILS_FYNE_SCALE`（防迴圈 guard + UI 邏輯尺寸計算用）**（2026-09-11 dev）
- 真機已驗證：`FYNE_SCALE=1.5 ./cryoutils-ng-desktop -ui native` 可用（用戶測試 2026-09-11）
- headless 驗證（2026-09-11 a1 Xvfb）：1280×800 無 re-exec（scale 1.0）· 3840×2160 re-exec `FYNE_SCALE=2` 生效

### 6.6.3 4K 排版修正
- [x] 內容欄寬度上限 + 置中：`ui/fyneui/layout.go` `cappedCenterLayout`（`contentMaxWidth=1100`），scroll 內容外包 `container.New(cappedCenterLayout{}, content)`，寬度 = min(視窗寬, 1100) 水平置中（修 select/按鈕在 4K 最大化被拉成 ~3750px 長條）（2026-09-11 dev）
- [x] 視窗尺寸跟隨螢幕：`windowSizeForScreen()` 取代固定 `Resize(1280×800)` — 螢幕 >1280×800 時 = **80% 螢幕（physical px）÷ scale**（scale 生效後 logical px 不超螢幕），並 `w.CenterOnScreen()` 置中；Deck 原生 1280×800 維持 full screen（2026-09-11 dev）
- [x] Status 區塊改雙欄：`container.NewGridWithColumns(2, ...)`（**Fyne v2.7.4 用 `NewGridWithColumns`**，非 `NewGridWrap` — 後者是固定 cell size 會改尺寸）7 行 key: value → 4 行（2026-09-11 dev）
- [x] 各 section 包 `widget.NewCard(title, subtitle, content)`（v2.7.4 為 3 參）（標題層級 + 分組感）（2026-09-11 dev）
- [ ] （可選，**違反單欄設計決策，需用戶另行拍板**）視窗寬 >2000px 時兩欄 GridWrap

### 6.6.4 驗證
- [x] a1 Xvfb headless 截圖：1280×800（scale 1.0，full screen，無 re-exec）+ 3840×2160（scale 2.0 re-exec，視窗 3070×1726 於 (384,216) = 80% 置中），Deck 原生解析度沒被改壞、中文渲染正常（2026-09-11 dev）
- [ ] scp 二進位上真 Deck，用戶驗收（4K + 1280×800 雙解析度）

## Phase 5.7: Sudo 安全修正 ✅
- [x] `core/sudo.go` — `RenewAuth()` 加 `-k` flag 強制忘記 cached timestamp，回傳 `error`，檢查空密碼
- [x] `core/sudo.go` — `TestAuth()` 加 `-k` flag 防止 cached timestamp 繞過驗證
- [x] `core/swap.go` — `ChangeSwapSize` 的 `RenewAuth()` 檢查 error
- [x] `cmd/desktop/api.go` — `handleSwapResize` 的 `RenewAuth()` 檢查 error + emit 錯誤訊息
- [x] `AGENTS.md` — Known Issues 更新
- [x] `go vet ./...` ✅ · build ✅ · 錯誤密碼正確拒絕 ✅

## P0.5: Sudo 認證重構 ✅（2026-08，代碼層完成；真機驗證併入 Phase 6 統合）

**背景**: 經過分析發現兩個問題需要修正：
1. Web UI 密碼層在 Steam Deck 環境下多餘（遊戲機預設信任同機使用者）
2. CLI 以 `sudo` 執行時，`RenewAuth()` 因 `e.Password == ""` 而失敗，但實際操作會成功（已是 root）

**決策**:
- Web UI 移除密碼輸入 UI（保留 token 作為 API 安全層）
- `RenewAuth()` 新增 `os.Geteuid() == 0` 檢查：已以 root 執行時直接跳過

- [x] `core/sudo.go` — `RenewAuth()` 新增 `os.Geteuid() == 0` 檢查，已為 root 時直接回傳 `nil`
- [x] `core/sudo.go` — `TestAuth()` 保留（供 Fyne 原生 UI / 未來可能的管理員模式使用）
- [x] `web/src/App.tsx` — 移除 Header 密碼輸入框、`sudoLocked`/`password` 狀態、`handleAuth`
- [x] `web/src/App.tsx` — 移除全部 11 處 `disabled={sudoLocked ...}` 條件（按鈕僅受 `busy`/選值控制）
- [x] `web/src/types.ts` — N/A（`sudoLocked` prop 原本定義在 `App.tsx` 各 component 的 inline interface，`types.ts` 從未承载）
- [x] `cmd/desktop/api.go` — 移除 `handleAuth` 端點（`POST /api/auth` + handler 全刪）
- [x] `web/src/api.ts` — 移除 `auth()` 函數
- [x] 附帶清理：`web/src/locales/{en,zh-TW}.json` 移除 `header.*` sudo 鍵；`App.css` 移除 `.sudo-group`/`.btn-sudo`/`.sudo-locked/.sudo-unlocked`
- [x] 驗證：`core` build+vet+test ✅、`web` `npm run build` ✅、`cmd/desktop` `CGO_ENABLED=0 go vet` ✅
- [x] `AGENTS.md` — Known Issues + Permissions + UI Design Decisions 更新
- [ ] CLI `swap` 命令在 SteamOS 上驗證（`sudo ~/.cryoutils_ng/cryoutils-ng swap 16`）→ 併入 Phase 6 統合真機驗證

## P0: ZRAM 支援（最高優先級）

**背景**: SteamOS 3.6 引入 zram swap（priority 100, ~7.2 GB），與 swap file（priority -2, 16 GB）並存。原始 CryoUtilities 開發時 zram 尚未存在，因此程式碼完全未處理 zram。

**已確認的 Bug**:
- `ChangeSwapSize()` 執行 `swapoff -a` 後僅重新啟用 swap file，zram 需等待 systemd 裝置掃描（約 18 分鐘）才恢復
- 程式碼中零 zram 相關處理（`grep zram` 無匹配）

- [x] `core/swap.go` — `ChangeSwapSize()` 完成後主動重新啟用 zram（`swapon /dev/zram0`）
- [x] 在 `swapoff -a` 前偵測 zram 是否已啟用，作為恢復的判斷依據
- [x] `core/swap.go` — 新增 `GetZramStatus()` 方法：偵測 `/dev/zram0` 是否在 `/proc/swaps` 中，回傳啟用狀態
- [x] `core/swap.go` — 新增 `GetZramSizeBytes()` 方法：讀取 `/sys/block/zram0/disksize`
- [x] `core/engine.go` — 新增 `getTotalSwapGB()`：從 `/proc/meminfo` 的 `SwapTotal` 計算（包含 zram + swap file）
- [x] `core/engine.go` — `GetStatusSummary()` 新增 `ZramSizeGB`, `ZramActive`, `TotalSwapGB` 欄位
- [x] `web/` — Swap 區塊 UI 重構：
  - 顯示 **Swap File Size**（用戶可調整的參數）
  - 顯示 **ZRAM Size**（直接讀取 zram disksize）
  - 顯示 **ZRAM Status**（啟用 / 停用）
  - 顯示 **Total Swap**（zram + swap file 總和，從 `/proc/meminfo` 讀取）
- [x] `web/src/types.ts` — 新增 `ZramSizeGB`, `ZramActive`, `TotalSwapGB` 欄位到狀態結構
- [x] `cmd/desktop/api.go` — `GET /api/status` 自動回傳新增欄位（透過 `GetStatusSummary()`）
- [x] `core/swap.go` — `GetSwapFileSize()` 行為保持不變（繼續只報告 swap file 大小）
- [x] CLI `status` 命令新增 zram 狀態輸出
- [x] `AGENTS.md` — Known Issues 更新（記錄 zram 行為與修正）

> ⚠️ **本 Phase 的 zram 恢復設計是錯的**，實際會把 zram 摧毀 — 見下方 **P0.7**。

---

## P0.7: swap resize 摧毀 zram（**方法 A 已實作並已 Deck 真機驗收通過，2026-10-20**）

> **調查報告與根因實證：`docs/zram-swapoff-reset.md`**（2026-10-20 真機 journal 實證，含時序與單位名稱）。
> 症狀：只要改 swap size，zram 就消失；`zramctl` 空、`/sys/block/zram0/disksize` = 0、總 swap 23 GiB → 16 GB。

**根因一句話**：`core/swap.go:109` 的 `sudo swapoff -a` 會讓 systemd 把 zram **整套拆掉** —— `dev-zram0.swap` 轉 Deactivated →
`systemd-zram-setup@zram0.service` 停止 → 其 `ExecStop=/usr/lib/systemd/system-generators/zram-generator --reset-device zram0`
把 `disksize` 寫 0。之後 `core/swap.go:231` 的 `swapon /dev/zram0` 是對**容量 0 的裝置**操作，必然 EINVAL 失敗，且失敗被 `core/swap.go:232` 吞掉。
**不會**自己回來（舊 P0 「等 systemd 掃描約 18 分鐘」的假設錯誤）；且第二次 resize 起 `GetZramStatus()` 已是 false → 連嘗試都不嘗試。

### 方法 A 施作項（Dev VM，2026-10-20 完成）
- [x] `DisableSwap()`：`swapoff -a` → 只關 swap file 本身（`swapoff <location>`）
- [x] 移除 `ChangeSwapSize()` 的 zram 重啟區塊與 `zramActive` 前置偵測（保留 `zramBefore` **只用於警告**：resize 後若 zram 不見了就 log 出 `systemctl restart systemd-zram-setup@zram0.service`，不自行動系統）
- [x] 新增換頁安全檢查 `checkSwapMigrationSafety()`：`/proc/swaps` 的 `Used` vs `MemAvailable` **加上其他 swap 的剩餘容量**，並保留 `SwapMigrationMarginKB`（512 MB）；不足時在 swapoff **之前**明確報錯中止（真機有 `earlyoom`）
- [x] 新增 `core/sudo.go` `runSudo()`：root 時直接執行（不包 sudo）、非 root 時 `sudo -S` + stdin 密碼（cached timestamp 時 stdin 寫入失敗不計錯）；**一律把命令列 + stderr 帶進 error**（舊代碼 `.Output()` + 固定誤訊息是本次 bug 藏起來的原因）
- [x] 可觀測性（方法 C）：log 改 `O_APPEND`，去掉 `cmd/cryoutilities/main.go` 與 `cmd/desktop/main.go` 的 `os.Remove(LogFilePath)`；CLI 失敗同時印到 **stderr**（不然安全檢查的拒絕訊息只能在 log 裡看到）
- [x] `core/` 單測 12 個（`core/swap_test.go`）：`/proc/swaps` 解析、zram 行辨識（舊 bug 迴歸）、`MemAvailable` 解析、遷移空間數學（含 Deck 實測數字）、**命令插樁（fake PATH）**驗證 `swapoff` 只被叫一次且只帶 location、命令序列不出現 `-a`/`zram`、dd 不會拿到空 `of=`、失敗會 best-effort 重開 swap、`status` 的 `SwapFile` 不為空
- [ ] （方法 B 安全網）**刻意不做**：只 log 警告、不執行 `systemctl restart …`。真機已證明 zram 不再消失（2026-10-20）→ **降為可選、低優先**；除非之後又發現別的路徑會拆 zram
- [x] 驗收：`docs/zram-swapoff-reset.md` §7 逐項真機驗 —— **2026-10-20 通過**（新 binary 四次 resize 2→8→16→12 GB，`/dev/zram0` 仍在且 Priority 100、Size 7579644 kB；詳見該文件 §10）
  - ⚠ 唯一未觸發：換頁安全檢查的**拒絕路徑**（需先把 swap used 拉高到超過 MemAvailable 才能驗證），待後續專門測一次

### 實作時新發現的三個缺陷（已一併修）
- [x] **`cmd/desktop/api.go` 把 resize 五步驟重寫了一遍，完全繞過 `ChangeSwapSize()`** → web 路徑從來沒有 zram 處理、也不會吃到新安全檢查（兩份實作就是這個 bug 類的來源）。現已改為只呼叫 `e.ChangeSwapSize()`，進度訊息由 core 透過 `OnProgress` → `core.ProgressChannel` 送出（非阻塞 send，避免 SSE 無人讀取時 swap 已關卻卡住）
- [x] **CLI `swap <n>` 從不解析 `e.SwapFileLocation`**（只有 `GetSwapFileSize()` 會設）→ 單獨跑 CLI 時 `dd of=` 是空路徑。新增 `resolveSwapFileLocation()`，每個 swap 步驟都自己解析
- [x] **`status` 的 `SwapFile:` 印空**（真機冒煙新 binary 時發現）：`GetStatusSummary()` 的 map 複合字面量按原始碼順序求值，`"SwapFile": e.SwapFileLocation` 在會解析位置的 `"SwapSizeGB"` 之前就讀了 → 改成開頭先 `resolveSwapFileLocation()`，並加迴歸測試

### Dev VM 環境與測試限制（2026-10-20）
- [x] Dev VM（`a1` = `192.168.1.15` / `Code-Agent-01`，Ubuntu 24.04.5）原**無 Go / 無 Node**；已裝 **Go 1.26.10** 到 `/usr/local/go`（對齊 repo `go 1.26` directive；官方 sha256 已驗證）。`pkg-config` 缺 `xkbcommon` 不影響 Fyne v2.7.4 建置（`cmd/desktop` vet + `ui/fyneui` build 皆過）
- [x] 變更前基準：`core` vet  clean、`go test ./...` ok → 變更後相同，且新增測試全 PASS（無 skip）
- [x] 兩個 binary 已建：`cmd/cryoutilities/cryoutils-ng`（`CGO_ENABLED=0`，6.2 MB）、`cmd/desktop/cryoutils-ng-desktop`（`CGO_ENABLED=1`，49.5 MB）；兩者皆在 `.gitignore`（第 84/85 行），不入版控
- [ ] **Dev VM 無法做「真 swap」端到端**：它是 LXC（核心 `7.0.0-3-pve`，`modprobe zram` 無 kernel modules；`/dev/loop-control` 不存在）、且 `/home` 在 ZFS 上（`swapon: it appears to have holes`）。所以端到端只能**在 Deck 真機驗**；Dev VM 端以命令插樁測試鎖定行為

### 真機已做的補救（2026-10-20，SteamOS 端）
- [x] `systemctl restart systemd-zram-setup@zram0.service` 已把 zram 完整還原（disksize 7761559552、zstd、Priority 100、`SwapTotal: 24356856 kB`）

---

## P0.8: 真機 feedback 第二批（2026-10-20，已修，待真機驗收）

- [x] **啟動 native UI 時被問兩次 sudo 密碼**（終端一次 + UI 對話框一次）— 根因是 `core/util.go` 四處舊的 `exec.Command("sudo", ...)` 沒有 `-S`，sudo 從 tty 問密碼；`getUnitStatus()` 的 `sudo cat` 在 status 卡片建立時就觸發。已全部改走 `runSudo()`；沒有密碼時回 `ErrNoSudoPassword`，完全不啟動 sudo。
- [x] **native UI swap 下拉的 "Current Size" 不會跟隨真實大小**（要重開 app 才更新）— swap 卡片只建立一次；改為 `uiCtx.addRefresher()` + `Select.SetOptions` + `nextSwapSelection()`，5 秒 ticker 與操作完成後都會更新；swappiness placeholder 跟隨（僅在使用者沒輸入時）。
- [x] `setUnitValue()` 原本吞掉所有錯誤、還把 `sudo tee` 輸出倒進 `os.Stdout`（寫入失敗看起來像成功）→ 改走 `runSudoWithInput()` 並回傳 error。
- [x] 密碼解鎖後立即 `refreshStatus()`（原本要等 5 秒才把需要 root 的讀值補上）。
- [ ] **真機驗收**（使用者）：新 binary md5 — CLI `85642ef7f01c10d6e79a7482767a454d`、desktop `a2bcd4bf0330ef8bc3f2fd2ed66a41bc`
  - 啟動 native UI：終端**不再**出現 `[sudo] password for deck:`，只有 UI 對話框問一次
  - 改完 swap size 後，下拉的 "Current Size" 應立即變成新值（不必重開 app）
  - 從清單明確選一個大小（例如 8），5 秒後與 resize 後都應保持你選的項目（只要它還在選項裡）
- [ ] （新 issue，未修）未解鎖時 unit 讀值顯示 `false` 而非 unknown；見 AGENTS.md Known Issues
- [ ] （新 issue，未修）web UI 的 swap 大小清單寫死，未走 core 的 `GetAvailableSwapSizes()`

### 真機驗收（2026-10-20，Deck）
- [x] 方法 A binary 真機通過：改 swap size 不再讓 zram 失效（四次 resize 2→8→16→12 GB，zram0 Priority 100 全程不動）
- [ ] 換頁安全檢查的拒絕路徑尚未在真機觸發（需刻意把 swap used 拉高）

---

## Phase 7 (future): Decky Loader Plugin
- [ ] React frontend reused from Phase 4
- [ ] Python shim calling CLI binary (`main.py`, `plugin.json`)
- [ ] Distribution zip with `backend/src → backend/out → bin/` CI convention
- [ ] Safety gating for swap-resize on Decky

---
**Naming (confirmed):** CryoUtils NG · `cryoutils-ng` · `~/.cryoutils_ng/` · `cryoutils_ng_steam_data`
**License:** GPLv3 (derivative of CryoUtilities by CryoByte33)
