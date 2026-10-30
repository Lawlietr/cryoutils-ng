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

## P0.8: 真機 feedback 第二批（2026-10-20，已修並已真機結案）

- [x] **啟動 native UI 時被問兩次 sudo 密碼**（終端一次 + UI 對話框一次）— 根因是 `core/util.go` 四處舊的 `exec.Command("sudo", ...)` 沒有 `-S`，sudo 從 tty 問密碼；`getUnitStatus()` 的 `sudo cat` 在 status 卡片建立時就觸發。已全部改走 `runSudo()`；沒有密碼時回 `ErrNoSudoPassword`，完全不啟動 sudo。
- [x] **native UI swap 下拉的 "Current Size" 不會跟隨真實大小**（要重開 app 才更新）— swap 卡片只建立一次；改為 `uiCtx.addRefresher()` + `Select.SetOptions` + `nextSwapSelection()`，5 秒 ticker 與操作完成後都會更新；swappiness placeholder 跟隨（僅在使用者沒輸入時）。
- [x] `setUnitValue()` 原本吞掉所有錯誤、還把 `sudo tee` 輸出倒進 `os.Stdout`（寫入失敗看起來像成功）→ 改走 `runSudoWithInput()` 並回傳 error。
- [x] 密碼解鎖後立即 `refreshStatus()`（原本要等 5 秒才把需要 root 的讀值補上）。
- [x] **真機驗收通過（2026-10-20，用戶確認）** — binary：CLI `85642ef7f01c10d6e79a7482767a454d`、desktop `a2bcd4bf0330ef8bc3f2fd2ed66a41bc`
  - 啟動 native UI：終端不再出現 `[sudo] deck 的密碼：`，只有 UI 對話框問一次 ✅
  - 改完 swap size 後，下拉的 "Current Size" 立即更新，不必重開 app ✅；從清單明確選的項目也保留 ✅
  - 佐證（`~/.cryoutils_ng/cryoutils_ng.log`）：`grep -c sudo` = **0**（不再留 sudo 失敗行，啟動問密碼也不產生噪音）；每 5 秒出現 `Available Swap Sizes: [1 - Default 2 4 6 8 12 16 - Current Size 20 24 32]` → refresher 確實在跑；swapfile 17179869184（16 GB）與 zram 並存
  - Dev VM 端 pty A/B：控制組裸 `sudo -k cat` 出現 `[sudo] deck 的密碼：`；新 CLI 以非 root 跑 `status` 完全無提示（132 ms）
- [ ] （新 issue，未修）log 噪音與膨脹：5 秒 status poll 每次寫約 5 行，43 分鐘 721 行，log 無 rotation；refresher 又重複抓一次 system state（見 AGENTS.md Known Issues）
- [ ] （新 issue，未修）未解鎖時 unit 讀值顯示 `false` 而非 unknown；見 AGENTS.md Known Issues
- [ ] （新 issue，未修）web UI 的 swap 大小清單寫死，未走 core 的 `GetAvailableSwapSizes()`

### 真機驗收（2026-10-20，Deck）
- [x] 方法 A binary 真機通過：改 swap size 不再讓 zram 失效（四次 resize 2→8→16→12 GB，zram0 Priority 100 全程不動）
- [ ] 換頁安全檢查的拒絕路徑尚未在真機觸發（需刻意把 swap used 拉高）

---

## P0.9: unit 參數不可偵測 / 不可用 + sudo 密碼外洩到 log（**2026-10-30 修正，已在 Deck 真機驗證**）

> 真機 feedback：SteamOS 上 UI 的 **Huge Pages、Shared memory in THP、Compaction Proactiveness、HugePage Defrag、Page Lock Unfairness** 全部「不可偵測與不可用」——狀態永遠是 off/unknown，點了只彈錯誤。
> 根因鏈與實證：`docs/unit-params-sudo-stdin.md`。

### 根因（兩個獨立缺陷疊在一起，都在 P0.8 重構後浮現）

**(a) 寫入壞 —— sudo 的密碼行被當成資料寫進系統參數**
- `runSudoWithInput` 用 `sudo -S`（**沒帶 `-k`**）。sudo 只在「真的需要驗證」時讀 stdin；`RenewAuth`/`TestAuth` 剛在每個操作前驗證過 → timestamp 還有效 → sudo **完全不讀 stdin** → 整段 payload（第一行是明文 sudo 密碼）被 `tee` 當資料寫進 sysfs/procfs。
- Deck log 實證（2026-10-20 17:56，desktop md5 `a2bcd4bf…`）：`tee /sys/kernel/mm/transparent_hugepage/enabled: exit status 1: <密碼>` 舆 `alwaystee: …: 不適用的引數`（EINVAL）；五個參數每一次 toggle 都是同一個形狀。
- 結果：kernel 拒寫 → `tee` exit 1 → `setUnitValue` 回錯 → `runTask` 彈錯誤對話框 = **不可用**；toggle 沒成功、5 秒刷新讀到原值 = 看起來像**不可偵測**。
- Dev VM 先用非 root 帳號（`lawliet`）+ 真實 sudo 重現：快取態下目標檔案 = `<密碼>\n<value>\n`；改用 `sudo -S -k` 後 = `<value>`。

**(b) 讀取壞 —— 狀態讀取不需要 root 卻走了 root**
- `getUnitStatus()` 走 `runSudo("cat", …)`，沒有密碼時（web 模式永遠沒密碼；native 未解鎖時也沒有）回 `ErrNoSudoPassword` → 五個 `get*Status()` 全部 false。
- 真機 A/B（2026-10-30，同一台 Deck、同一指令 `cryoutils-ng status`，deck 身份、完全不給密碼）：
  - 舊 binary：五項全 false，log 出現 5 行 `Unable to get current …`
  - 新 binary：`HugePages: true / ShMem: true / CompactionProactiveness: false / Defrag: false / PageLockUnfairness: false`，**log 零錯誤行**；與 deck 直接 `cat` 的 kernel ground truth（`[always] madvise never` / `[advise]` / `1` / `20` / `5`）逐項吻合

**(c) 兩個小項**：`rm` 沒帶 `-f`（revert 一個不存在的 tmpfiles.d conf 會被當成失敗）；`SetHugePages` 試圖 unlink `/proc/sys/vm/nr_hugepages`（procfs 檔案不能 unlink，root 也 EPERM；真機 log：`拒絕不符權限的操作`）。

### 資安（同一批代碼的副作用）
- [x] **明文 sudo 密碼曾被寫進 log**：`tee` 把輸入 echo 回 stdout，而該輸入第一行就是密碼；失敗時 `commandError` 又把它帶進 log。真機 `~/.cryoutils_ng/cryoutils_ng.log` 實測 **5 處**含密碼，而 log mode 原是 **0666（世界可讀）**
- [x] 修正：`-k`（密碼永遠不會變成 payload）、stdout/stderr 分流（錯誤只帶 stderr）、`e.redact()` 抹掉字串中的密碼、兩支 main 的 `os.OpenFile` mode **0666 → 0600**
- [x] 真機清理：既有 log 的 5 處已 scrub 成 `[REDACTED-PASSWORD]`，log `chmod 600`；修好後的 binary 跑完驗收 probe，log 内含密碼 **0 處**
- [ ] **建議用戶更新 Deck 的 sudo 密碼**（它已在該機器的可讀 log 裡存在了 10 天）

### 施作（Dev VM，2026-10-30）
- [x] `core/sudo.go`：`runSudoWithInput` → **`sudo -S -k`**；stdout/stderr 分開 buffer；回傳值只用 stdout；錯誤帶 command line + **stderr** tail；新增 `e.redact()`；新增 `useSudoAlways` 測試錨點（讓 sudo 包裝路徑在 root 跑測試時也能被覆蓋）
- [x] `core/util.go`：`getUnitStatus` 先用 `unitFileReader`（= `os.ReadFile`）直接讀，**只有 `EPERM` 才退回 sudo**；ENOENT（參數在此 kernel 不存在）不浪費一次特權嘗試；抽出 `parseUnitValue`（括號列表解析）；未知參數名明確報錯；`removeFile` → `rm -f`；`writeFile` 的暫存檔改用 `os.Remove`（它住在自己的 install 目錄，不該 `sudo rm`）
- [x] `core/memory.go`：`SetHugePages` 拿掉對 `/proc/sys/vm/nr_hugepages` 的 unlink 嘗試（附註解；若要釋放已保留的 huge pages 應改成寫 0 —— **未採，屬行為變更，需用戶拍板**）
- [x] `cmd/cryoutilities/main.go`、`cmd/desktop/main.go`：log mode 0666 → 0600

### 測試（Dev VM；`core` 全綠）
- [x] 新增 `core/unit_test.go`：`TestGetUnitStatusReadsWithoutRoot`（零 sudo 呼叫 + 列表/純數值解析）、`TestGetUnitStatusFallsBackToSudoWhenDenied`（模擬 EPERM → 仍走 sudo）、`TestParseUnitValue`、`TestRemoveFileToleratesAMissingPath`、`TestSetUnitValueReportsFailures`
- [x] `core/sudo_test.go`：fake sudo 改為**重現 sudo 的 stdin 契約**（要驗證時才吃掉第一行密碼；有 `-k` 就一定驗證）→ 新增 `TestRunSudoWithInputNeverLeaksPasswordIntoTheCommand`、`TestRunSudoReturnsStdoutOnly`、`TestRunSudoFailureCarriesStderrButNotThePassword`；`TestRunSudoWithoutPasswordNeverSpawnsSudo` 改用 `useSudoAlways`（不再因 root 身份 skip）
- [x] **三個回歸實驗**（把舊行為改回去，新測試必須紅）：去掉 `-k` → `parameter file = "<密碼>\nalways\n", want "always\n"`；合併 stdout/stderr → `output = "VALUE\nPROMPT_AND_NOISE\n", want only stdout`；讀值改回走 sudo → `getUnitStatus = "", sudo password not provided`
- [x] `CGO_ENABLED=0 go vet`（core、`tools/p09probe`）✅ · `CGO_ENABLED=1 go vet ./cmd/desktop/...` ✅ · core `go test -count=1 ./...` 全過
- [x] 新二進位 md5：CLI `1751459e162dfb1ed15cf1c11686f057`（6.2 MB）· desktop `cc95f58e23032b23239132a6b5001400`（49.5 MB）· 驗收 probe `640cce9a27f7c4af97a0dbb06e7c23b0`

### 真機驗收（2026-10-30，Steam Deck `deck@192.168.1.232`，sshd 首次使用）
- [x] 佈署前備份：`cryoutils-ng.bak-p08`、`cryoutils-ng-desktop.bak-p08`；推送後兩支 md5 與 Dev VM 一致
- [x] **讀取 A/B**（deck 身份、無密碼）：舊 = 五項 false + 5 行 `Unable to get`；新 = 五項與 kernel ground truth 一致、零錯誤行
- [x] **寫入 round-trip（`tools/p09probe`，以 deck 身份經真 sudo）**：五個參數各自 `Toggle()` 兩次 → 值確實改變並回到原值、`err=<nil>`；`Set()` 後 `/etc/tmpfiles.d/<param>.conf` 確實建立（tmpfiles.d 持久層正常）、`Revert()` 後值回 default、conf 被移除。**probe failures counted: 0**
- [x] 收尾：probe 留下的 `hugepages.conf` / `shmem_enabled.conf` 已 `sudo rm -f`，`/etc/tmpfiles.d` 回空；kernel 值全數回原（`always` / `advise` / `20` / `1` / `5`）
- [x] 副作用檢查：`/dev/zram0` 仍在（zstd、7.2G、[SWAP]）、swapfile 16 GB 未動、log 內密碼 0 處、log mode 600
- [ ] **用戶驗收**：native UI 實際點擊五個 toggle（checkbox 即時反映、不再彈錯誤），與 Phase 6.6 的 4K/1280×800 驗收一同做

### 本輪留下的新待辦
- [ ] （新 issue，未修）**VRAM 讀值不穩定**：ssh session 下 `status` 報 `VRAM: 100MB`，真機 log 曾是 `Found a VRAM of 1024` —— `GetVRAMValue()` 依賴 `glxinfo`，無 X session / 拿到 llvmpipe 時就會錯。要嘛偵測 session，要嘛讀不到時顯示 unknown
- [ ] （新 issue，未修）**web 模式寫入一律失敗**（`e.Password` 恒空，P0.5 已移除密碼入口）：讀取現在正常，寫入仍限 native UI 或 root CLI。是否補回 web 密碼入口是獨立決策
- [ ] （未變）log 噪音與無 rotation（P0.8 那項）：六個 unit 參數的失敗行已消失（每輪少約 5 行），但 swap/swappiness/VRAM 與 `Available Swap Sizes` 仍每 5 秒寫

### 環境注記（2026-10-30）
- [x] SteamOS 真機已開 sshd（`deck@192.168.1.232`）→ Dev VM 可直接查證/佈署（見 AGENTS.md Working Rules）。憑證存 repo 根 `.secrets`（mode 600、已加入 `.gitignore`，**絕不 commit**）
- [x] 真機 kernel `7.2.7-valve2-2-neptune-72-g31504582a073`；六個 unit 路徑全部存在且 `0644 root:root`（所以「讀值不需 root」成立）；`/` 掛載為 `rw`（所以 `writeUnitFile` 的 `sudo mv` 能成功，tmpfiles.d 可用 —— 但 SteamOS 更新會蓋掉 rootfs 內容，持久化的 conf 可能需要 `atomic-update-keep`，待驗證）
- [x] Dev VM 磁碟：`/` 可用 18G（44% 使用）

---

## P0.10: 真機 feedback 第三批（2026-10-30，驗收當天）

> 用戶在 Deck 上跑完 P0.9 驗收後的三則回報：`p09_check.sh` 全 PASS；`p09probe` 那條命令跑不出東西；Memory 五項已可按，但 **Huge Pages 每次重啟後都會取消勾選**（用戶認為不正常）。另要求補入原版的「亮綠燈」視覺（用戶明示**暫不處理**，只補 TODO）。

### 已處理

- [x] **`p09probe` 跑不出東西**（我的工具問題，不是 app）：舊版只從 stdin 第一行讀密碼，且**在讀之前不印任何東西**，所以她那條 `read -rs PW && printf ... | p09probe` 看起來像整個卡住、沒有輸出。修正 `tools/p09probe/main.go`：進場先印 banner；stdin 是 tty 時改向 `/dev/tty` 提示並用 `stty -echo` 關掉回顯（自己問密碼）；有管線時仍讀 stdin（我腳本化的用法不變）。真機兩種路徑都驗過：pipe → `failures counted: 0`；pty → 出現 `sudo password (input hidden): `。新 md5 `19d63b4ba15c6c77a05d94b4755be67b`（已佈署到 Deck）
- [x] **Huge Pages「取消勾選」查證：不是持久化 bug**。log 實證（`~/.cryoutils_ng/cryoutils_ng.log`，19:39:56–19:43:14）她在 3.5 分鐘內點了 **7 次** Huge Pages，每一次都成功：`Enabling hugepages → Writing always → Writing always to /etc/tmpfiles.d/hugepages.conf`；`Disabling hugepages → Writing madvise → Removing /etc/tmpfiles.d/hugepages.conf`。奇數次 → 最後停在 off，所以重開 app 顯示未勾選是**正確讀值**
- [x] **驗收工具副作用已收尾**：probe 的 `Set()` 會建 conf、`Revert()` 會刪 conf。她按出來的三個（`compaction_proactiveness.conf`=0、`defrag.conf`=0、`page_lock_unfairness.conf`=1）保留；probe 多建的 `shmem_enabled.conf` 我已 `rm -f`。收尾後狀態 = 她的意圖：`thp=madvise / shmem=advise / defrag=0 / comp=0 / pl=1 / swappiness=60`，swapfile 16 GB 與 zram0 未動

### 真正的三個缺口（要動代碼/要拍板，未處理）

- [ ] **SteamOS 的 THP 開機預設是 `always`，不是 `madvise`**（本機 2026-10-27 開機、`/etc/tmpfiles.d` 全空情況下實測 `[always] madvise never`）。後果有兩個：
  - app 啟動時 Huge Pages 本來就是**勾選**狀態，她第一次按下去其實是 **關閉**（log 19:39:56 第一行就是 `Disabling hugepages...`）。「勾選 = 我要啟用」的 checkbox 直覺在這裡是反的
  - `DefaultHugePages = madvise`（沿用上游 CryoUtilities）與 kernel boot 預設 `always` 不一致，而 **Revert 的語意是「寫 runtime 值 + 刪除 conf」**，所以「關閉」不會跨重開機：重開機後 THP 又回到 `always` → checkbox 又變回勾選。她感覺到的「每次重啟都不一樣」就是這個振盪（另外四個參數的 Default 值剛好等於 kernel 預設，所以只有 hugepages 有這個問題）
- [ ] **Memory 五個 checkbox 沒有接 `addRefresher`**（`ui/fyneui/sections_memory.go`）：5 秒 ticker 只更新 status 卡片，checkbox 只在「建立當時」與「自己點完之後」設值。真實值被外部改動、或 task 失敗時，UI 會與 kernel 不同步。修法 = P0.8 swap 下拉同一個 `uiCtx.addRefresher()` 模式，但只在 `updating == false` 時同步，避免把使用者正在進行的操作蓋掉。驗收方式：headless Fyne（`test` package）對 checkbox 打兩下，斷言 `ToggleXxx` 各被呼叫一次、`Checked` 等於真實狀態、task 進行中的重複點擊不會第二次觸發
- [ ] **Revert 是否應該持久化（需用戶拍板，三選一）**：
  - **A**：Revert 也寫 conf（`w <path> - - - - madvise`），「關閉」就能跨重開機 —— 但等於是把 Valve 的 THP 預設 `always` 主動壓成 `madvise`
  - **B**：`DefaultHugePages` 改成 `always`（對齊 SteamOS 實際預設），這一項在 Deck 上幾乎永遠是開，checkbox 實際變成「只能關」
  - **C**：不動 default，只修 checkbox refresher，並且在 UI 上顯示**當前值**（`always` / `madvise`）而不只 true/false，把「這個勾代表什麼」講清楚
  - 我的建議：先做無爭議的 checkbox refresher + 顯示當前值，A/B 交由你決定（C 的顯示層其實和下面那條「亮綠燈」是同一块地方，可以併入同一個改造）

### 補 TODO、用戶明示暫不處理

- [ ] **原版 CryoUtilities 的「亮綠燈」視覺**：已啟用的 tweaks 在原版有綠色發光/亮起的高亮，用戶要求補進我們的 UI。實作點與上一條的「顯示當前值」是同一個區域：五個 toggle 加狀態指示（綠＝recommended、灰＝default/unknown），並透過 refresher 跟隨 5 秒刷新。web UI 同处也要做（`web/src/App.tsx` 的 memory 區塊），兩邊共用 i18n 鍵
- [ ] **p09probe 應該先 snapshot `/etc/tmpfiles.d` 並在結束時還原**（現在它會建立/刪除 conf，對「使用者已持久化的設定」有破壞性；本次靠事後人工清除）

### 用戶已知情/接受的風險（不再重複提醒）

- [x] Deck 只有用戶一人使用（2026-10-30 明示）→ log 曾外洩 sudo 密碼這件事她已知情，**更新密碼的建議她未采纳**。AGENTS.md 對應條目已標註，避免我每輪重複提醒

---

## Phase 7 (future): Decky Loader Plugin
- [ ] React frontend reused from Phase 4
- [ ] Python shim calling CLI binary (`main.py`, `plugin.json`)
- [ ] Distribution zip with `backend/src → backend/out → bin/` CI convention
- [ ] Safety gating for swap-resize on Decky

---
**Naming (confirmed):** CryoUtils NG · `cryoutils-ng` · `~/.cryoutils_ng/` · `cryoutils_ng_steam_data`
**License:** GPLv3 (derivative of CryoUtilities by CryoByte33)
