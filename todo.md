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
- [x] 建立 `web/public/locales/en.json` + `zh-TW.json` 翻譯檔
- [x] `web/src/i18n/locales.ts` — **單一來源** locale 定義（`LOCALES` const + `Locale` 類型自動推導）
- [x] `web/src/i18n/index.tsx` — i18n hook（Context + JSON 載入 + localStorage 持久化）
- [x] `web/src/LanguageSelector.tsx` — 語言選單從 `LOCALES` 自動生成
- [x] `web/src/types.ts` — `MEMORY_PARAM_LABELS` → `MEMORY_PARAM_KEYS`（translation key 化）
- [x] `web/src/App.tsx` — 所有硬編碼字串替換為 `t.*` 翻譯呼叫
- [x] `web/src/main.tsx` — 包覆 `I18nProvider`
- [x] `web/src/App.css` — `.locale-select` 樣式
- [x] `tsc --noEmit` ✅ · `npm run build` ✅
- [ ] 後續添加更多語言（日語、簡體中文等）
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

## Phase 6: Verification (user — requires real Steam Deck)
- [x] Real Steam Deck acceptance testing (2026-08-11, CLI + Web UI basic smoke)
- [x] Verify original CryoUtilities untouched (confirmed at `~/.cryo_utilities/`)
- [x] CLI `status` command verified on real hardware (`SwapSizeGB: 16` ✅)
- [ ] Web UI visual acceptance on real Deck (swap size fixed, need to verify `--app=` window)
- [ ] All CLI commands verified on real hardware (`swappiness`, `recommended`, `stock`, `hugepages`, etc.)
- [ ] `steam://openurl` fallback 確認（Flatpak `--app=` 失敗時的關鍵 fallback）
- [ ] Flatpak zsh globbing 問題記錄到 README（URL 需加引號）

## Phase 7 (future): Decky Loader Plugin
- [ ] React frontend reused from Phase 4
- [ ] Python shim calling CLI binary (`main.py`, `plugin.json`)
- [ ] Distribution zip with `backend/src → backend/out → bin/` CI convention
- [ ] Safety gating for swap-resize on Decky

---
**Naming (confirmed):** CryoUtils NG · `cryoutils-ng` · `~/.cryoutils_ng/` · `cryoutils_ng_steam_data`
**License:** GPLv3 (derivative of CryoUtilities by CryoByte33)
