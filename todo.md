# CryoUtils NG — Todo

## Phase 0.5: CI/CD Infrastructure ✅
- [x] `.github/dependabot.yml` — gomod weekly, npm weekly, ignore npm major, 5 PR limit
- [x] `.github/workflows/dependabot-auto-merge.yml` — auto-merge patch/minor, major needs review
- [x] `.github/workflows/release.yml` — updated actions, removed Fyne apt deps, fixed binary path
- [ ] GitHub UI: branch protection rulesets for `main` and `develop` (Require PR, Require status checks, Block force pushes)
- [ ] GitHub UI: enable Dependabot security updates

## Phase 0: Dependency Modernization & Baseline ✅
- [x] Go 1.26.5 verified at `/usr/local/go/bin/go`
- [x] `core/` created as independent Go module (`go 1.26`, 5 deps, no Fyne, no CGO)
- [x] Root `go.mod` updated: go 1.26, updated deps, `replace cryoutils-ng/core => ./core`
- [x] Fyne retained in root module (internal/ still imports it)
- [x] `go build ./...` ✅ · `go vet ./...` ✅ · core `CGO_ENABLED=0` ✅
- [ ] Root module `CGO_ENABLED=0` build — blocked until Phase 1 removes Fyne from internal/

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
- [ ] Delete `internal/` — deferred until Phase 3 (desktop server still needs some internal refs) or Phase 4 (full UI rewrite)

## Phase 2: CLI Rework + `status` Command ✅
- [x] `status` command added to CLI (prints all tuning statuses)
- [x] All CLI subcommands use `core.Engine` methods
- [x] CLI smoke test: `status` ✅, `swappiness 60` ✅, `help` ✅

## Phase 3: Desktop Web Server
- [ ] `cmd/desktop/main.go` — localhost web server (127.0.0.1 + random token)
- [ ] REST API endpoints: `GET /api/status`, `POST /api/swap`, `POST /api/swappiness`, etc.
- [ ] SSE progress endpoint for long-running operations
- [ ] `go:embed` of web build → single-file install

## Phase 4: Single-Page React UI
- [ ] `web/` — React + Vite + TypeScript single-page UI
- [ ] Vertical single column, no tabs
- [ ] Header inline sudo unlock
- [ ] Responsive: 3840×2160 (4K) and 1280×800 (Steam Deck native)
- [ ] Plain CSS, framework-agnostic components
- [ ] Headless browser verification (Playwright) at both resolutions

## Phase 5: Packaging
- [ ] `install.sh` — installs binary to `~/.cryoutils_ng/`, creates `.desktop` file
- [ ] `uninstall.sh` — removes install artifacts, preserves data
- [ ] `launcher.sh` — wrapper for Steam Big Picture mode
- [ ] `.desktop` file — `CryoUtilsNG` entry in Steam Deck gamemode
- [ ] Coexistence: never overwrite original CryoUtilities install paths

## Phase 6: Verification
- [ ] Real Steam Deck acceptance testing
- [ ] Verify original CryoUtilities untouched
- [ ] All CLI commands verified on real hardware
- [ ] Web UI visual acceptance on real Deck

## Phase 7 (future): Decky Loader Plugin
- [ ] React frontend reused from Phase 4
- [ ] Python shim calling CLI binary (`main.py`, `plugin.json`)
- [ ] Distribution zip with `backend/src → backend/out → bin/` CI convention
- [ ] Safety gating for swap-resize on Decky

---
**Naming (confirmed):** CryoUtils NG · `cryoutils-ng` · `~/.cryoutils_ng/` · `cryoutils_ng_steam_data`
**License:** GPLv3 (derivative of CryoUtilities by CryoByte33)
