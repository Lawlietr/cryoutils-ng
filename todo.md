# CryoUtils NG — Todo

> **待提醒（已拍板、未施作）**：**UI-1 UI/UX 重新設計** — 範圍 = **只有 native**；IA = 單頁 + 左側 nav rail。
> 文件：`docs/ui-redesign-plan.md`（決策清單在檔頭）。細節見本檔 UI-1 節。
> **前置工作**：**P0.12**（CI 從未執行過，三個待拍板項在該節）。UI-1 的 **Step 0（基線截圖）不受阻，可立即進行**；
> Step 3 之後的每一步都應該有 CI 的遠端紅綠，所以 P0.12 宜在 Step 3 之前完成。

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
- [x] `web/src/types.ts` — N/A（`sudoLocked` prop 原本定義在 `App.tsx` 各 component 的 inline interface，`types.ts` 從未承載）
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
- [x] 真機清理：既有 log 的 5 處已 scrub 成 `[REDACTED-PASSWORD]`，log `chmod 600`；修好後的 binary 跑完驗收 probe，log 內含密碼 **0 處**
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
- [x] **Huge Pages「取消勾選」查證（此條結論下半段已被下方的「主因」推翻，保留原始觀察以便對照）**。log 實證（`~/.cryoutils_ng/cryoutils_ng.log`，19:39:56–19:43:14）她在 3.5 分鐘內點了 **7 次** Huge Pages，每一次都成功：`Enabling hugepages → Writing always → Writing always to /etc/tmpfiles.d/hugepages.conf`；`Disabling hugepages → Writing madvise → Removing /etc/tmpfiles.d/hugepages.conf`。奇數次 → 最後停在 off，所以重開 app 顯示未勾選是**正確讀值**
- [x] **驗收工具副作用已收尾**：probe 的 `Set()` 會建 conf、`Revert()` 會刪 conf。她按出來的三個（`compaction_proactiveness.conf`=0、`defrag.conf`=0、`page_lock_unfairness.conf`=1）保留；probe 多建的 `shmem_enabled.conf` 我已 `rm -f`。收尾後狀態 = 她的意圖：`thp=madvise / shmem=advise / defrag=0 / comp=0 / pl=1 / swappiness=60`，swapfile 16 GB 與 zram0 未動

### 主因（本日找到並已修）

- [x] **Fyne 的 `Check.SetChecked()` 會回呼 `OnChanged` → app 啟動時自己把參數翻掉**（這才是「每次重啟後取消勾選」的真正原因）。
  `sectionsMemory()` 在**建卡片時**就 `cb.SetChecked(狀態)`，而 Fyne v2.7.4 的 `Check.Refresh()` 會呼叫 `OnChanged(old, checked)`
  （`SetText` 內部直接 `Refresh`），等於每次同步狀態都觸發一次 `ToggleXxx()` —— 凡是已经在 recommended 值的參數都會被 app 自動翻掉，
  而 Huge Pages 在 SteamOS 上出廠就是 recommended（`always`），所以每次啟動都被自動關掉。
  - 真機實證：log `19:39:50` 啟動 → `19:39:56 Disabling hugepages... / Writing madvise / Removing hugepages.conf`，而她第一次點擊在 `19:40:24`（而且是 `Enabling`，代表當時已被 app 關掉）
  - 修法：`syncing` 旗標包住「把狀態推进 checkbox」的動作（`updating` 包住執行中），回呼內 `if syncing || updating { return }`
  - 測試：`ui/fyneui/memory_state_test.go` → `TestMemorySectionBuildNeverTogglesAnything`（只建卡片、不碰滑鼠，fake sudo 記錄必須為空）。
    把守衛拿掉會紅：`building the memory card ran privileged commands: -S -k tee …/hugepages`；加回來綠
  - 同類呼叫點全數稽核（swap/gamedata 的 `Select.SetSelected`、語言 select、swappiness `Entry`）：只有 memory checkbox 會做特權寫入，
    其餘回呼是空函式或已有「值相同即 return」守衛
  - 真機驗證：新版 native UI 執行期間 log 內 `Writing … for param` 0 次、`sudo` 0 次（Dev VM Xvfb 1280×800 zh-TW 截圖確認），
    Deck 上 `p09_check.sh` 新增 §5「app 開著不動 22 秒不得寫參數」→ PASS

### 已處理（顯示層與讀值）

- [x] **顯示真實值，不再只有 true/false**（原 C 案）：`core` 新增 `UnitState{Value, IsRecommended}` + `GetUnitStates()` 單趟讀
    （順帶消掉舊 `GetStatusSummary()` 每次刷新多 spawn 6 次 sudo）、讀不到時 `Value = unknown`（不再壓成 false）；
    `GetStatusSummary()` 帶 `<Param>Value` 五鍵 → CLI `status` 印出、native Status 卡改成「標籤: 值」（`大記憶體頁面: madvise`）、
    native Memory 五個 checkbox 標籤帶值且由真實狀態驅動（`addRefresher` + `syncing` 守衛）、web Status 卡以 `· <值>` 顯示、
    web Memory 每列顯示值，讀不到顯示新 i18n 鍵 `status.unknown`（en/zh-TW 已補，`npm run build` 後已 embed）
- [x] **SteamOS 五個參數的開機值已實測並寫進 `core/config.go` 與 `docs/unit-params-sudo-stdin.md` §8.3**：
    `always / advise / 20 / 5 / 1`（開機 2026-10-27、`/etc/tmpfiles.d` 無我方檔案、`/proc/cmdline` 無 `transparent_hugepage=`）。
    **`DefaultHugePages` 不改為 `always`**（曾實作並寫測試，隨後否決並還原）：`Default*` 是 `hugepages disable` / Stock 預設**寫进去的值**，
    改成 `always` 會讓那些命令回報成功卻什麼都不改，且「把 THP 關掉」在此 app 中直接消失。
    改由 `TestRecommendedValuesMatchWhatSteamOSShips` 同時鎖住兩側（recommended 必須等於出貨值、stock 必須維持上游值）
- [x] `TestRevertHugePagesWritesTheDefaultAndDropsTheConf`（原 `…EndsOnTheBootDefaultNoMatterTheOrder` 名稱不實，它並未測順序）：
    改成先種 `always`（Deck 的開機值）再驗 `Revert` 必須寫 default 且 `rm -f` conf；三種突變（不寫值／不刪 conf／`rm` 沒 `-f`）都會紅

### 待你拍板（未處理）

- [ ] **用戶明確「關掉」的選擇要不要跨重開機**：現在 `Set` 寫 conf、`Revert` 刪 conf，所以把 THP 關掉只在這次開機有效，重開機被 Valve 出廠值 `always` 蓋回（she 感到的「重啟後又不一樣」的另一半）。
  選項：**A** Revert 也寫 conf（`w <path> - - - - madvise`），關閉可持久，但等於把 Valve 的 THP 預設主動壓低；
  **B** 加一個明確的「套用/持久化」按鈕（與下面的亮綠燈同一區），把「這次有效」與「永久」分開；
  **C** 維持現狀，只在 UI 標明「此變更不跨重開機」。建議 **B**（語意最清楚，也與亮綠燈共用版面）
- [ ] **亮綠燈視覺**（用戶明示**暫不處理**，見下節）

### 補 TODO、用戶明示暫不處理

- [ ] **原版 CryoUtilities 的「亮綠燈」視覺**：已啟用的 tweaks 在原版有綠色發光/亮起的高亮，用戶要求補進我們的 UI。實作點與上一條的「顯示當前值」是同一個區域：五個 toggle 加狀態指示（綠＝recommended、灰＝default/unknown），並透過 refresher 跟隨 5 秒刷新。web UI 同处也要做（`web/src/App.tsx` 的 memory 區塊），兩邊共用 i18n 鍵
- [ ] **p09probe 應該先 snapshot `/etc/tmpfiles.d` 並在結束時還原**（現在它會建立/刪除 conf，對「使用者已持久化的設定」有破壞性；本次靠事後人工清除）

### 用戶已知情/接受的風險（不再重複提醒）

- [x] Deck 只有用戶一人使用（2026-10-30 明示）→ log 曾外洩 sudo 密碼這件事她已知情，**更新密碼的建議她未采纳**。AGENTS.md 對應條目已標註，避免我每輪重複提醒

---

## P0.11: web UI 驗證鏈（2026-11-11 開出）

把 web UI 的無頭驗證真正跑起來時發現的三個缺口。已修的標 ✅，未做的標 `[ ]`。

### 已修（2026-11-11，Dev VM）
- [x] **`web/package.json` 沒宣告 Playwright，但 `package-lock.json` 有**：`npm ci` 會直接把
    `@playwright/test`、`playwright`、`playwright-core` **移除**（`npm ci --dry-run` 實測），
    也就是任何乾淨環境裝完相依套件之後 `npx playwright test` 就不存在——
    本機能跑只是因為 `node_modules` 是早前留下的。已補 devDeps `@playwright/test` 與 `playwright`
    （`^1.62.1`，與 lock 已解析的版號一致），lock 已同步，`npm ci --dry-run` 不再移除
- [x] **`web/tests/screenshot.spec.ts` 的斷言與 UI 文案脫節**：找 `label` 裡的 `Swap Size`，
    但實際是 `swap.swapSizeGB` = `Swap File Size (GB)`（`Swap Size` 只剩 status 卡的 `<span>`）
    → 該測試在兩個 project 都紅。已改為 `Swap File Size`
- [x] **驗證鏈可用**：`python3 tools/webstub/stub.py --no-sse` + `BASE_URL=http://127.0.0.1:8099 npx playwright test`
    → 14 passed（兩個 project × 7 測試），4K 與 1280×800 截圖重新產生

### 待辦
- [ ] **CI 完全沒有跑 web UI 測試**：`.github/workflows/release.yml` 只有 `cd web && npm ci && npm run build`，
    `playwright.config.ts` 為 CI 準備的 `webServer`（`npm run preview`）從未被觸發。要做：
    1. 加 UI 驗證步驟：`npm ci` → `npx playwright install chromium`（需要的 browser build 版號由
       `node_modules/playwright-core/browsers.json` 決定，cache 裡的舊 build 不算）→ 跑 `npx playwright test`
    2. `npm run preview` 不提供 `/api/status`，status 區會是空的或 `unknown`，
       **P0.10 的顯示在 CI 看不到**。兩條修法：**(a)** spec 用 `page.route('**/api/status', …)`
       直接吃 `tools/webstub/fixtures/*.json`（不需要任何 server）；**(b)** job 裡起
       `python3 tools/webstub/stub.py --no-sse`，走與本機相同的路
    3. 建議 **(a)**：fixture 是單一來源，且 CI 不用維持一支 Python server
- [ ] **spec 的等待條件要寫對**：`waitForLoadState('networkidle')` 在 API 保持連線（open SSE）時
    永不達到。走 (a) 之後 spec 不依賴 server；若仍要接 stub，**必須加 `--no-sse`**（實測不加時該測試逾時失敗）
- [ ] `web/tests/screenshots/` 那兩張 PNG 入版控、且由 fixture（假資料）產生：要嘛保留並標明
    provenance（已在 `tools/webstub/README.md` 標明），要嘛改成 CI artifact 上傳、版控不存二進位

---

## P0.12: CI 從未執行過 + desktop 的 CGO 建置命令已失效（2026-11-11 開出）

一輪 UI-1 的前置調查裡發現的。**先前在 AGENTS.md／對話中曾誤判為「CI 每次 push 都紅」，實測是 CI 一次都沒跑過**，
以下全部是本輪實測數字。

### 事實 1：`Main` workflow 從未有過任何一次執行

- `gh api repos/Lawlietr/cryoutils-ng/actions/runs` → `total_count: 0`
- `gh api .../actions/workflows/331011566/runs`（`Main` = `.github/workflows/release.yml`）→ `total_count: 0`
- `gh api .../actions/permissions` → `{"enabled":true,"allowed_actions":"all"}` → **不是被停用**
- 觸發條件永不成立：`release.yml:3-7` 是 `on: push: branches: [main, develop]`，但 remote 上只有
  `dev`（`74b36a6`，全部重寫工作）與 `main`（`db020c2`，**2024-01-31**，fork 時從上游帶來的原 CryoUtilities 狀態）。
  **`develop` 分支不存在**（`contents/.github/workflows/release.yml?ref=develop` → 404）。
  所有 push 都進 `dev`，它不在觸發清單裡。
- 本 repo 是 fork：`parent: CryoByte33/steam-deck-utilities`，2026-08-10 建立，`default_branch: main`。
- `gh pr list --state all` → **零筆 PR**，所以 `dependabot-auto-merge.yml` 同樣從未執行。

### 事實 2：CI 的 desktop 建置命令已經失效，而且永遠不會報錯

- `release.yml:48`：`CGO_ENABLED=0 go build -ldflags="-s -w" -o cryoutils-ng-desktop ./cmd/desktop`
- Dev VM 重現同一條命令的失敗訊息：
  `imports cryoutils-ng/ui/fyneui → fyne.io/fyne/v2/internal/painter/gl → go-gl/gl/v2.1/gl:
  build constraints exclude all Go files`（Fyne 的 GL driver 帶 `//go:build cgo` 條件）
- 時間線：那條命令由 `e0c6235`（2026-08-11，Phase 4+5）寫下，**當時正確**（`cmd/desktop` 還只跑 web server）。
  `ed88424`（2026-08-23，Fyne native UI）讓 `cmd/desktop/main.go:34` 開始 `import "cryoutils-ng/ui/fyneui"`
  → 命令當場失效。AGENTS.md 同期已記錄「desktop 需 `CGO_ENABLED=1`」，**只有 CI 沒跟上**。

### 事實 3：本地建置一直正常，所以真機驗收從未被這件事誤導

- `CGO_ENABLED=1 /usr/local/go/bin/go build ./cmd/desktop` → 成功，**49,531,848 bytes**
- `ldd` 實測含 `libGL.so.1`、`libX11.so.6`、`libGLX.so.0`、`libXau.so.6` → **動態連結**，
  不是靜態單檔。若要把它當成 prebuilt 二进制發佈，必須同時寫明 runtime 相依。
- 她 scp 上 Deck 的每一支二進位都是這樣來的 → 本地 gate 是有效的，只是沒有遠端 gate。

### 事實 4：CI 沒有任何驗證步驟

`grep -rn 'go test|go vet|playwright' .github/workflows/` → **NONE**。
兩個 workflow 只有建置與上傳。本機三支 `ui/fyneui` 測試與 `core` 測試目前是綠的（實測 `ok 0.732s`、`ok 0.132s`），
但沒有任何東西保證它們未來綠。

### 事實 5：Dependabot 的目標分支也是錯的

- `.github/dependabot.yml` 的兩個 entry（`/` gomod、`/web` npm）**都沒有 `target-branch`**
  → PR 會落在 default branch `main`，也就是 2024-01-31 的上游舊狀態，不是 `dev`。
- 目前零 PR，所以尚未造成實際損害；但一旦 Dependabot 找到更新，會自動往一個已放棄的分支開 PR，
  而且 `dependabot-auto-merge.yml` 的 `on: pull_request`（無分支過濾）會對它執行 `gh pr merge --auto`。

### 待你拍板（三項；這是本節唯一需要你的部分）

- [ ] **(A) 分支模型**：把 `dev` 加進 `on: push` 的分支清單？還是把 `dev` 改名 `develop`／開始把 `dev` 併回 `main`？
    AGENTS.md 寫的是 gitflow，但 remote 上只有 `dev` 與 `main`，`develop` 不存在。
    這條決定事實 1、事實 5 與 Dependabot 目標分支三處。
- [ ] **(B) 版號與 release 名稱**：CI 一跑起來就會發佈。`release.yml:10` 是 `release_name: 2.0-${{ github.run_number }}`，
    `main-release` 還會 `gh release delete latest` 再重建 `latest`。本地 tags 裡有上游帶過來的
    `2.0-8`／`2.0-9`／`2.0-87`／`latest`／`develop`。這與 Phase 5.6 拍板的 `v0.1.0`／`v0.2.0` 方案直接衝突
    （見本檔 line 165 那條「版本號 `v0.1.0` → `v0.2.0` 待用戶確認」）。
    **建議：先改名再讓 CI 上線**，否則第一次觸發就多一個 public release。
- [ ] **(C) desktop 要不要經 CI 產出**：
    - 要 → `CGO_ENABLED=1` + apt `libgl1-mesa-dev xorg-dev`（build 端），且產物是動態連結（見事實 3）→
      要把 runtime 相依寫進 `docs/manual-install.md` 與 README。runner 是否從 `ubuntu-22.04` 升到 `ubuntu-24.04`
      （與 Dev VM 對齊，避免 glibc/GL 版本差異）也一起決定。
    - 不要 → CI 只 build CLI，desktop 維持本機 build —— 這才是 AGENTS.md 現在寫的做法（「Not published via
      Releases — users build from source」）。**建議先走這條**，風險最低、改動最小，之後要 prebuilt 再開一輪。

### 施作（Dev VM，不需要 Deck；(A)(B)(C) 有答案後半小時內可完成）

- [ ] 依 (A) 修 `release.yml:3-7` 與 `dependabot.yml` 的 `target-branch`
- [ ] 依 (B) 改 `release_name`，必要時先手動建一個正式 tag（`v0.2.0`）
- [ ] 依 (C) 拆建置：CLI 維持 `CGO_ENABLED=0` 靜態單檔；desktop 移除或改 `CGO_ENABLED=1` + apt 相依
- [ ] 加 fail-fast 測試步驟（放在 build 之前）：
    `cd core && CGO_ENABLED=0 go test ./...`、root module `CGO_ENABLED=1 go vet ./...`、
    `CGO_ENABLED=1 go test ./ui/fyneui/...`（三支本機 headless 已綠；若 CI 需要顯示，用 `xvfb-run` + apt `xvfb`）
- [ ] 接 P0.11 的 web UI 驗證（走建議的 (a)：spec 用 `page.route('**/api/status', …)` 讀
    `tools/webstub/fixtures/*.json`，CI 不起 Python server；`npx playwright install --with-deps chromium`）。
    Dev VM 這側已可用：Node `v24.22.2`、`web/node_modules` 54 MB、`~/.cache/ms-playwright/chromium-1234` 都在
- [ ] 第一次上線後確認 GitHub 上出現一次完整綠的 run，且 Release 頁沒有意外產物

### 環境注記（2026-11-11）

- **AGENTS.md 的「Node 未裝」已過時**：Node `v24.22.2` 在 `/root/.local/share/pi-node/current/bin/`
  （pi 自帶的 runtime，`node`/`npm`/`npx` 都可用）。web 的依賴與 Playwright 也已就位。
  `df -h /` 可用 **22 GB**（遠高於 3 GB 閾值）。這條要在 AGENTS.md 更正。

### 與 UI-1 的先後關係

- **不阻塞 UI-1 Step 0/1/2**：push 到 `dev` 不會觸發任何東西，本機 gate（三支測試 + `go vet` + Xvfb 截圖）已可用。
- **阻塞建議在 Step 3 之前**：Step 3 起改動面擴大（拆 Status 卡、拿掉 `NewGridWithColumns(2)`、VRAM 卡消失），
  沒有遠端紅綠就無法在幾天後回答「這一改是不是回歸了舊缺陷」。

---

## UI-1: UI/UX 重新設計（2026-11-11 開出，**未施作**）

用戶回饋：UI 比原始項目彆腳、設置擠在一起、大小排序不整齊、不美觀。本輪只做**量測與設計規格**，
**一項改動都沒做**。診斷數據、設計 token、版面規格、驗收斷言全部在 **`docs/ui-redesign-plan.md`**。

### 已完成的調查（2026-11-11，Dev VM）

- [x] web 兩條截圖（P0.11 那組）與 native Xvfb 截圖都逐項量過。native 新舊 PNG 逐位元組相同
      （81,125 bytes）→ 舊截圖代表現況
- [x] **web 沒有 4K 適配**：`#root { max-width: 720px }`（`App.css:34`）一條走到底，實測內容寬
      688 px @1280（54%）與 688 px @3840（**18%**）→ 4K 兩側 3,156 px 純背景
- [x] **「不整齊」的根因是沒有設計系統**：實測 15 種間距值（只算 padding/margin/gap）、
      11 種 `font-size`（含 3 種互不相容的 `clamp()` 公式）；`:root` 只有顏色與 radius，
      沒有 spacing／type scale
- [x] **Memory 五行沒有值欄**：`.memory-item { justify-content: space-between }`（`App.css:189`）
      → 中間空白 **395~539 px**（五行都不同）；label 起點對齊在 315/316，但右側群組貼齊 963、
      起點只有 864 與 903 兩種 → 值字串左邊界隨字串長度飄移（無 value 欄）
- [x] **Swap 兩行按鈕不同線**：控制件 465..574 vs 432..509（起點差 33 px）、按鈕 592..702 vs
      526..591（起點差 66 px、寬 110 vs 65）；`.swap-row` flex + `label{min-width:100px}`，
      控制件無寬度規則
- [x] **Status 卡 12 筆塞 3 欄 ~204 px**，而 label 是 uppercase + letter-spacing 的 20+ 字元；
      VRAM 是一張不顯示數值的空卡
- [x] **native 的 4K 骨架是對的**（`contentMaxWidth = 1100` 在 1280/3840 都生效且置中、
      `FYNE_SCALE=2` re-exec 生效）→ **不要推翻它**，要改的是內部對齊
- [x] **native 的實際缺陷**：Status 兩欄中間空白 **478 px** 且左右欄語意交錯；Select 寬 **1,085 px**；
      Card 在暗色主題下無可見邊界（六區塊連成一條長欄）；checkbox 未勾選對比過低；
      1280×800 一屏只露出 3/6 區塊
- [x] **三項懷疑已排除**（不要再查，見文件 §2.8）：native 標題字級不一致（實測 23.5/23/23.6 px 相同）、
      web 卡間距不一致（實測 20~21 px 一致）、native 1100px 上限沒生效（生效）

### 已拍板（2026-11-11，用戶；也寫進文件檔頭的決策清單）

- [x] **IA = 單頁 + 左側 nav rail**（≥1200 logical px 出現，200 px 寬，5 個項目）。
      響應式 tab 與摺疊區塊**否決**（理由見文件 §6）
- [x] **Presets 上移**：新順序 Status → Presets → Swap → Memory → Game Data
- [x] **VRAM 合併進 Status 卡 A**，獨立卡消失
- [x] **施作範圍 = 只有 native**（`ui/fyneui/` + `i18n/locales/`）。**web 那一輪尚未排程**：
      §2.2（web 4K 內容只占 18%）與 §2.6（web 無設計 token）的缺陷本輪不會被修掉

### 施作順序（native only；每步都要過 gate）

- [ ] **Step 0 — 基線（先做，不需你在場）**：把 `docs/ui-redesign-plan.md` 附錄 A 的截圖／量測片段收進
      `tools/`（建議 `tools/ui1shots/`），產生 native 的 **`-before` 基線**：三解析度
      （1280×800 / 2560×1440 / 3840×2160）×（zh-TW／en）。**現況盤上沒有任何 native 截圖**
      （實測 `find` 只有 `web/tests/screenshots/` 那兩張 web PNG 與兩個 icon），
      沒有基線就無法證明 Step 1 之後的效果，也無法證明沒回歸。同步確認三支測試綠（本輪實測 `ok 0.732s`）。
- [ ] Step 1 — theme wrapper：新增一個包住 `theme.DefaultTheme()` 的 `fyne.Theme` 實作。
      v2.7.4 的 interface 只有 `Color`／`Font`／`Icon`／`Size` 四個方法，**沒有 `NewTheme` 建構子**。
      把 §4 token 映到 `SizeNameText`、`SizeNameHeadingText`、`SizeNameSubHeadingText`、
      `SizeNameCaptionText`（名稱是 `helperText`）、`SizeNamePadding`、`SizeNameInnerPadding`、
      `SizeNameSeparatorThickness`
- [ ] Step 2 — 區塊邊界：Fyne 的 `Card` renderer 用 `NewShadowingRenderer(…, CardLevel)` →
      **只有陰影沒有邊框**，theme 也沒有 Card 邊框色名。要自畫 `canvas.Rectangle`
      （`StrokeColor` = `theme.ColorNameSeparator`、`StrokeWidth = 1`），或改自訂 card-like 容器
- [ ] Step 3 — Status 拆卡 A／卡 B + 固定 label 寬的定義列表（拿掉 `NewGridWithColumns(2)` 的
      語意交錯、中間 478 px）；VRAM 併入卡 A（`sections_vram.go` 的卡消失）
- [ ] Step 4 — Swap／Game Data 改 `widget.NewForm`，Select 上限 320 px（現況 1,085 px），
      按鈕全部落到同一動作欄
- [ ] Step 5 — Memory 三欄：名稱／固定寬值欄／開關（現況值是塞在 label 括號裡，沒有值欄）；
      checkbox 未勾對比與 44 px hit target
- [ ] Step 6 — 區塊順序改 Presets 上移；nav rail：`container.NewBorder` 第三參數、
      `vScroll.ScrollToOffset(fyne.NewPos(0, sectionObj.Position().Y))`（v2.7.4 沒有 `ScrollTo(float32)`）、
      **不得**註冊進 `addRefresher`；i18n 新增 5 鍵（en／zh-TW 同步）
- [ ] Step 7 — 幾何斷言腳本進 `tools/`（附錄 A 已有可用片段）：native Xvfb 三解析度
      （1280×800 / 2560×1440 / 3840×2160）×（zh-TW／en），存 `-before`/`-after`
- [ ] Step 8 — 真機驗收（用戶在場）：4K 看 nav rail 出現且可點、Deck 1280×800 看它不出現；
      驗收中翻轉的 kernel 參數／tmpfiles.d／log 狀態要收尾並列進回報

每步的 gate：`ui/fyneui` 三支測試（`memory_state_test.go`、`swap_select_test.go`、`refresh_test.go`）
+ `CGO_ENABLED=1 /usr/local/go/bin/go vet ./...` + Xvfb 截圖對比。

### 硬性紅線（施作時不能踩）

- [ ] **改 native 版面不得破壞 P0.10 的 `syncing` 守衛**：`ui/fyneui/memory_state_test.go` 必須綠
      （「把狀態同步進 widget」不等於「接受使用者輸入」）
- [ ] 換 Select 實作要重跑 `ui/fyneui/swap_select_test.go`（P0.8：下拉要跟隨真實大小）與
      `ui/fyneui/refresh_test.go`
- [ ] **本輪不跑 `npm run sync-loc`**：它會把 `../i18n/locales/*.json` 覆蓋進 `web/src/locales/`，
      與「本輪不動 `web/src/`」抵觸。nav rail 的 5 個新鍵只進 `i18n/locales/en.json` 與 `zh-TW.json`；
      web 沒有 nav rail、不引用那些鍵，所以少鍵不會壞。要同步時另外單獨一輪做
- [ ] 不動 `core/` 行為與既有 locale 鍵語意（新增 nav rail 標籤鍵除外）；**本輪不動 `web/src/`**；
      web 端不補 sudo 密碼輸入（P0.5 缺口是獨立議題）
- [ ] 垂直節奏不用改（實測本來就是一致的，見 §2.8）

---
## Phase 7 (future): Decky Loader Plugin
- [ ] React frontend reused from Phase 4
- [ ] Python shim calling CLI binary (`main.py`, `plugin.json`)
- [ ] Distribution zip with `backend/src → backend/out → bin/` CI convention
- [ ] Safety gating for swap-resize on Decky

---
**Naming (confirmed):** CryoUtils NG · `cryoutils-ng` · `~/.cryoutils_ng/` · `cryoutils_ng_steam_data`
**License:** GPLv3 (derivative of CryoUtilities by CryoByte33)
