# CryoUtils NG — Todo

> Status legend: `[ ]` pending · `[x]` done · `[~]` in progress
> **Project name**: CryoUtils NG (binary: `cryoutils-ng`, install dir: `~/.cryoutils_ng`, data root: `cryoutils_ng_steam_data`)

## Phase 0.5 — CI/CD Infrastructure (GitHub)
- [x] 建立 `.github/dependabot.yml`，設定 Dependabot 自動 PR：
  - `gomod` ecosystem → `/` 目錄，weekly 檢查
  - `npm` ecosystem → `web/` 目錄，weekly 檢查
  - commit message prefix: `deps` / `deps(web)`
  - 限制 open PR 數量（各 5 則），避免 PR 洪水
  - npm 忽略 major 升級（需手動審視）
- [x] GitHub repository 設定：branch protection + auto-merge (patch/minor)
  - **注意**：auto-merge 需在 GitHub UI 手動開啟（Settings → Pull requests → Auto-merge → Allow auto-merge）
  - branch protection（Require CI pass / Require branches to be up to date）也需在 UI 手動設定

## Phase 0 — Dependencies & Baseline
- [ ] Install Go 1.26.5 on dev box (Ubuntu 24.04 x86_64, Node v22 present)
- [ ] Create `core/` with fresh `go.mod` (`go 1.26`), keeping only:
  - `golang.org/x/sys@v0.47.0`
  - `github.com/moby/sys/mountinfo@v0.7.2`
  - `github.com/otiai10/copy@v1.14.1`
  - `github.com/cristalhq/acmd@v0.12.0`
  - `github.com/andygrunwald/vdf@v1.1.0` (unchanged, already latest)
- [ ] Drop `fyne.io/fyne/v2` + full indirect tree (glfw, gopherjs, gl-js, oksvg, rasterx, textlayout, freetype, systray, x/mobile, x/image, x/net, x/text, fsnotify)
- [ ] Confirm `CGO_ENABLED=0` static build works (no GLFW/OpenGL runtime deps)

## Phase 1 — Core Extraction (`core/`)
- [ ] Migrate `internal/config.go` constants wholesale (recommended/default values, `UnitMatrix`, size lists)
- [ ] Create `core.Engine` struct replacing global `CryoUtils`:
  - `InfoLog`, `ErrorLog *log.Logger`
  - `Password string` (sudo password injected by UI backend)
  - `OnProgress func(msg string, pct float64)` callback for long tasks
  - `SteamAPI map[int]string`, `SwapFileLocation string`
- [ ] Port `handler_swap.go` → Engine methods (swap file location/size, swappiness, resize flow: disableSwap→dd→chmod→mkswap→swapon, getAvailableSwapSizes)
- [ ] Port `handler_memory.go` → Engine methods (hugepages/shmem/compaction/defrag/page_lock Set+Revert+Toggle+Status)
- [ ] Port `handler_gpu.go` → read-only VRAM (glxinfo), note: actual VRAM change is via BIOS
- [ ] Port `handler_game_data.go`, `handler_library.go`, `handler_steam_api.go` (sync/cleanup, VDF parse, Steam app list)
- [ ] Move sudo helpers into `core/sudo.go`: password → `sudo -S echo` cache mechanism; `writeFile/setUnitValue/getUnitStatus/removeFile` renew auth first
- [ ] Replace Fyne progress widgets (`SwapResizeProgressBar`, `MoveDataProgressBar`) with `OnProgress` callback
- [ ] Port `internal/util_test.go` → `core` tests; `go test ./core/...`
- [ ] Behavior parity: recommended/stock values, swap resize flow, memory toggles identical to original

## Phase 2 — CLI Rework (`cmd/cryoutils-ng`)
- [ ] Point imports at `core`; keep all subcommands: `swap`, `swappiness`, `hugepages`, `compaction_proactiveness`, `defrag`, `page_lock_unfairness`, `shmem`, `recommended`, `stock`
- [ ] Replace `gui` subcommand with `desktop` (launches web server)
- [ ] Add `status` subcommand: print current settings (swap size/location, swappiness, memory params, VRAM) — scripting interface for future Decky backend

## Phase 3 — Web Server (`cmd/desktop`)
- [ ] Bind `127.0.0.1:<free port>` only; random token per launch, opened via `xdg-open http://127.0.0.1:<port>/?token=...`
- [ ] REST API:
  - `POST /api/auth` — validate + cache sudo password
  - `GET /api/status` — current settings + recommended/default + available swap sizes + drives + game data
  - `POST /api/swap/resize` `{size}`, `POST /api/swap/swappiness` `{value}`
  - `POST /api/memory/{toggle}`, `POST /api/recommended`, `POST /api/stock`
  - `POST /api/gamedata/sync`, `POST /api/gamedata/cleanup`
  - `GET /api/progress` — SSE stream for long tasks
- [ ] Token check on all privileged endpoints (protect against other local processes)
- [ ] `go:embed` web build into binary → single-file install

## Phase 4 — Single-Page React UI (`web/`)
- [ ] Vite + React + TypeScript scaffold (plain CSS, no heavy UI lib, no @decky/ui)
- [ ] **Single page, vertical single column** layout:
  - Header: app name + current-status summary strip + inline sudo unlock (password field; actions disabled until unlocked)
  - Action row: `Apply Recommended` / `Revert to Stock`
  - Swap: current size/location + resize selector + swappiness selector, each with current value
  - Memory: 5 cards (hugepages/shmem/compaction/defrag/page_lock) with inline current value + recommended-match badge + toggle
  - Storage: game data sync + cleanup controls
  - VRAM: read-only current value + note (change via BIOS)
  - Footer: version / disclaimer
- [ ] Long tasks: top-of-page progress bar fed by SSE (never lost on scroll)
- [ ] Responsive: `clamp()` fluid fonts, `max-width` content, compact breakpoint for 1280×800
- [ ] Components kept framework-agnostic for Phase 2 Decky reuse

## Phase 5 — Packaging (new name)
- [ ] `install.sh`: build Go binaries + `npm ci && npm run build`, install to `~/.cryoutils_ng/`
- [ ] New `.desktop` / `launcher.sh` for `CryoUtilsNG`
- [ ] New log file (`cryoutils_ng.log`), new external data root (`cryoutils_ng_steam_data`)
- [ ] `uninstall.sh`
- [ ] SteamOS side: zero dev toolchains, single prebuilt binary only

## Phase 6 — Verification
- [ ] `go build ./...`, `go vet ./...`, `go test ./core/...`
- [ ] `npm run build` in `web/`
- [ ] Headless browser (Playwright) screenshots at 3840×2160 and 1280×800
- [ ] CLI smoke test: `status` / `recommended` / `stock`
- [ ] Final acceptance on real Steam Deck (coexist with original CryoUtilities, no clobber)

## Phase 8 — Documentation Rewrite (README)
- [x] 重寫 `README.md`：
  - **移除**：CryoByte33 的個人資訊（YouTube 頻道、Patreon、Discord 連結、個人網站）
  - **保留**：對原始項目 CryoByte33/CryoUtilities 的致敬與說明（衍生作品、GPLv3 授權）
  - **重寫**：根據實際專案狀態（rewrite、新架構、新命名空間）更新所有描述
  - **更新**：安裝方式、功能列表、FAQ 連結指向新文件結構
  - **符合**：當前專案事實（Phase 0–7 路線圖、`core/` 架構、`web/` React UI）

## Phase 7 (future) — Decky Loader Plugin
- [ ] React frontend reused from Phase 4
- [ ] Python shim (`main.py`) calling `cryoutils-ng` CLI binary via subprocess
- [ ] `plugin.json` / `package.json` / distribution zip; `backend/src → backend/out → bin/` CI convention
- [ ] Game-mode sudo + swap-resize safety gating (strong warning / desktop-mode-only)
