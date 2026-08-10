# CryoUtils NG — AGENTS.md

## Project Overview
- **Purpose**: A Steam Deck utility to manage swap files, swappiness, and memory parameters — a rewrite of the original CryoUtilities (unmaintained for 1+ year).
- **Original maintainer status**: Original project by CryoByte33 is no longer maintained. This is a **rewrite with a new name** that must **coexist** with the original install (never overwrite `~/.cryo_utilities`, `cryo_utilities` binary, `cryoutilities_steam_data`, or its `.desktop` files).
- **Root Directory**: `/root/opencode-stuffs/steam-deck-utilities`
- **Dev environment**: Ubuntu 24.04 x86_64 (Proxmox VM). Steam Deck target is also x86_64 → native compile, no cross-compile. Go 1.26.5 + Node v22 (already present). SteamOS needs **zero** dev toolchains — it only receives a prebuilt binary.

## Architecture & Structure
- **`core/`**: UI-independent Go engine (single source of truth for all tuning logic). `core.Engine` struct (loggers, sudo password, `OnProgress` callback) replaces the old global `CryoUtils`. No Fyne, no CGO (`CGO_ENABLED=0` static binary).
- **`cmd/cryoutils-ng`**: CLI. All original subcommands + new `status` command (scripting interface for the future Decky backend).
- **`cmd/desktop`**: localhost web server (127.0.0.1 + random token), REST API, SSE progress, `go:embed` of the web build → single-file install.
- **`web/`**: React + Vite + TypeScript **single-page** UI (no tabs). Vertical single column; all settings + statuses on one view. Plain CSS, framework-agnostic components (reusable by Decky plugin later).
- **`internal/`**: legacy Fyne UI — to be deleted at end of Phase 1.

## UI Design Decisions (locked)
- **Single page, vertical single column** — no tabs (original's tabs deemed redundant).
- **Header inline sudo unlock** — page always visible (read-only statuses); password field in header unlocks privileged actions.
- **Responsive**: `clamp()` fluid fonts + `max-width` + compact breakpoint. Must render correctly at **3840×2160 (4K)** and **1280×800 (Steam Deck native)**.
- **VRAM section is read-only** (actual VRAM change is via BIOS).
- Long tasks (swap resize) show a top-of-page progress bar via SSE, never lost on scroll.
- Headless browser (Playwright) verification at both resolutions on dev box; final visual acceptance on real Deck.

## Naming (BLOCKING — PENDING USER INPUT)
Project name confirmed: **CryoUtils NG**.
- Go module: `cryoutils-ng`
- Binary name: `cryoutils-ng`
- Install dir: `~/.cryoutils_ng/`
- External data root: `cryoutils_ng_steam_data`
- Log file: `cryoutils_ng.log`
- CLI command: `cryoutils-ng`
- Desktop name: `CryoUtilsNG`
- Fyne App ID: `io.cryoutils-ng`

## Disk Space Policy (MANDATORY)
- **安裝任何開發套件前,必先執行 `df -h /` 檢查磁碟空間**。適用於任何環境:新 LXC、clone 後的環境、本機、未來所有機器。
- 套件粗估佔用(記住,用於判斷是否足夠):
  - Go toolchain + module/build cache: ~700MB
  - Node.js: ~100MB
  - Playwright + Chromium(含系統相依): ~1GB
  - `web/` node_modules: ~200–400MB
- **若可用空間 < 3GB:停下,回報實際可用數字,提醒用戶清空間或擴容,取得指示後才繼續**。
- 此規則不因環境變更而跳過(即使換機器/LXC/VM)。
- 佈建完成後回報最終 `df -h` 結果。

## Key Commands
- **Build core + CLI**: `go build ./...` (from repo root)
- **Build web UI**: `cd web && npm ci && npm run build`
- **Test**: `go test ./core/...` · `go vet ./...`
- **Run server (dev)**: `go run ./cmd/desktop` → opens `http://127.0.0.1:<port>/?token=...`
- **CLI (target)**: `sudo ~/.cryoutils_ng/cryoutils-ng <command> [parameter]`
- **Permissions**: tweaks need sudo; `core/sudo.go` handles password → `sudo -S echo` timestamp cache (same mechanism as original `renewSudoAuth`).

## Dependencies (modernized)
- **Kept + updated**: `golang.org/x/sys@v0.47.0`, `mountinfo@v0.7.2`, `otiai10/copy@v1.14.1`, `acmd@v0.12.0`, `vdf@v1.1.0` (already latest)
- **Removed**: `fyne.io/fyne/v2` and its entire stale indirect tree (glfw, gopherjs, gl-js, oksvg, rasterx, textlayout, freetype, systray, x/mobile, x/image, x/net, x/text, fsnotify) — most pinned 2021–2022
- **Go directive**: `go 1.26` in `core/go.mod`

## Important Constraints & Quirks
- **Swap Files**: only supports swap files, not swap partitions.
- **Coexistence**: new project must not overwrite original install. Both apps tune the same `/proc/sys` params and `/etc/tmpfiles.d/*.conf` — install-level coexistence is fine; system tuning state is "last one wins" (expected, not a conflict).
- **Platform**: optimized for SteamOS. Swap resize disables swap briefly — risky; handle with warnings. Decky/swap-resize safety gating deferred to Phase 7.
- **VRAM read** depends on `glxinfo` (mesa-utils) — original app already validated this on SteamOS.
- **Security**: desktop server binds 127.0.0.1 only + per-launch random token (prevents other local processes driving privileged ops).

## Testing & Verification
- Unit tests in `core/` (ported from `internal/util_test.go`).
- Status verified via CLI `status` command or `GET /api/status`.
- Headless screenshots at 4K + 1280×800; CLI smoke (`status`/`recommended`/`stock`).
- Final: real Steam Deck acceptance, verify original CryoUtilities untouched.

## Roadmap
- **Phase 0.5**: CI/CD infrastructure (Dependabot for gomod + npm, auto-merge patch/minor) — see `todo.md`
- **Phase 0**: dependency modernization (Go 1.26.5, fresh `core/go.mod`, drop Fyne) — see `todo.md`
- **Phase 1**: core extraction
- **Phase 2**: CLI rework + `status` command
- **Phase 3**: desktop web server (REST + SSE + token)
- **Phase 4**: single-page React UI
- **Phase 5**: packaging (install.sh, .desktop, launcher.sh, uninstall.sh)
- **Phase 6**: verification
- **Phase 7 (future)**: Decky Loader plugin — React frontend reused + Python shim calling CLI binary (`main.py`, `plugin.json`, distribution zip; `backend/src → backend/out → bin/` CI convention)

## License & Usage Rights
- **License**: GNU General Public License v3.0 (GPLv3) — inherited from the original CryoUtilities project.
- **Derivative Work**: this is a derivative work; must retain original copyright notices/license declarations, be distributed under GPLv3 (or compatible), and provide complete source code to recipients.
- **What You Cannot Do**: incorporate into proprietary/closed-source software; remove or alter the GPLv3 license terms.
- **UI Rewrite Intent**: core logic (handlers/config/utilities) preserved and rewritten in Go; UI rewritten independently (single-page web, different design from original Fyne tabs).

## Attribution & Documentation
- **Original Project**: CryoUtilities by CryoByte33 (unmaintained 1+ year) — this is a **rewrite** that honors and continues the original work.
- **README Policy**: 推上 GitHub 前必須重寫 `README.md`：
  - **移除**原作者個人資訊（YouTube、Patreon、Discord、個人網站）
  - **保留**對 CryoByte33 / CryoUtilities 的致敬與衍生作品聲明
  - **更新**所有描述符合當前專案實際狀態
