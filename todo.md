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
- [ ] 真實 Deck 驗證 zram 狀態顯示與 swap resize 後 zram 恢復（P0 代碼已合併未真機驗過，優先度高）
- [ ] All CLI commands verified on real hardware（`swappiness`、`recommended`、`stock`、`hugepages` 等，含 P0.5 後 `sudo` 執行 CLI 的 `Geteuid` 跳過路徑）
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
- [x] `core/swap.go` — `GetSwapFileSize()` 行為保持不变（繼續只報告 swap file 大小）
- [x] CLI `status` 命令新增 zram 狀態輸出
- [x] `AGENTS.md` — Known Issues 更新（記錄 zram 行為與修正）

---

## Phase 7 (future): Decky Loader Plugin
- [ ] React frontend reused from Phase 4
- [ ] Python shim calling CLI binary (`main.py`, `plugin.json`)
- [ ] Distribution zip with `backend/src → backend/out → bin/` CI convention
- [ ] Safety gating for swap-resize on Decky

---
**Naming (confirmed):** CryoUtils NG · `cryoutils-ng` · `~/.cryoutils_ng/` · `cryoutils_ng_steam_data`
**License:** GPLv3 (derivative of CryoUtilities by CryoByte33)
