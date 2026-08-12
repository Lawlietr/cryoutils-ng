# CryoUtils NG — AGENTS.md

## Project Overview
- **Purpose**: A Steam Deck utility to manage swap files, swappiness, and memory parameters — a rewrite of the original CryoUtilities (unmaintained for 1+ year).
- **Original maintainer status**: Original project by CryoByte33 is no longer maintained. This is a **rewrite with a new name** that must **coexist** with the original install (never overwrite `~/.cryo_utilities`, `cryo_utilities` binary, `cryoutilities_steam_data`, or its `.desktop` files).
- **Root Directory**: `/root/opencode-stuffs/steam-deck-utilities`
- **Dev environment**: Ubuntu 24.04 x86_64 (Proxmox VM). Steam Deck target is also x86_64 → native compile, no cross-compile. Go 1.26.5 + Node v22 (already present). SteamOS needs **zero** dev toolchains — it only receives a prebuilt binary.

## Working Rules (MANDATORY)
- **禁止自行 commit / push**：任何 `git commit`、`git push`、`git tag` 必須等你明確指示才執行。你會在 Dev VM 與 SteamOS 之間頻繁切換測試，擅自推送可能覆蓋你的本地修改。
- **工作前先確認環境**：每次任務開始時必須先用 `cat /etc/os-release | grep PRETTY_NAME` 確認當前在哪個環境，並回報給用戶。
  - Dev VM (`a1` / `root@192.168.1.15`)：Ubuntu 24.04，有 Go + Node + 編譯工具
  - SteamOS：rootfs 唯讀，無開發工具鏈，無 sshd，只接受預編譯二進位
- **二進位複製方向**：SteamOS 未啟用 sshd，Dev VM 無法 SSH 到 SteamOS。所有二進位複製都由 **SteamOS 主動執行 `scp` 從 Dev VM 拉取**（`scp root@192.168.1.15:/path/to/binary ~/.cryoutils_ng/`）。Dev VM 端編譯完成後只需報告結果，不嘗試推送。
- **Go 路徑**：`go` 不在 agent 的 PATH 中（`.bashrc` 對非互動式 shell 無效）。**所有 Go 命令必須使用完整路徑**：`/usr/local/go/bin/go`。例如：`/usr/local/go/bin/go build -o cryoutils-ng ./cmd/cryoutilities`。

## Architecture & Structure
- **`core/`**: UI-independent Go engine (single source of truth for all tuning logic). `core.Engine` struct (loggers, sudo password, `OnProgress` callback) replaces the old global `CryoUtils`. No Fyne, no CGO (`CGO_ENABLED=0` static binary).
- **`cmd/cryoutilities`**: CLI. All original subcommands + new `status` command (scripting interface for the future Decky backend). Binary name: `cryoutils-ng`. Published via GitHub Releases.
- **`cmd/desktop`**: localhost web server (127.0.0.1 + random token), REST API, SSE progress, `go:embed` of the web build. Binary name: `cryoutils-ng-desktop`. Not published via Releases — users build from source per README instructions.
- **`web/`**: React + Vite + TypeScript **single-page** UI (no tabs). Vertical single column; all settings + statuses on one view. Plain CSS, framework-agnostic components (reusable by Decky plugin later).
- **`internal/`**: legacy Fyne UI — **ready for deletion** (Phase 4 complete); retained only to avoid breaking root module `go build ./...` until packaging is done.

## UI Design Decisions (locked)
- **Single page, vertical single column** — no tabs (original's tabs deemed redundant).
- **No password unlock** — Steam Deck 環境預設信任同機使用者；token 已提供 API 安全層。
- **Responsive**: `clamp()` fluid fonts + `max-width` + compact breakpoint. Must render correctly at **3840×2160 (4K)** and **1280×800 (Steam Deck native)**.
- **VRAM section is read-only** (actual VRAM change is via BIOS).
- Long tasks (swap resize) show a top-of-page progress bar via SSE, never lost on scroll.
- Headless browser (Playwright) verification at both resolutions on dev box; final visual acceptance on real Deck.

## Desktop UI Launch (Decision — 方案 D)
- **使用環境**:UI 主要在 Steam Deck **Desktop Mode** 使用(Gaming Mode 非主要;但 fallback 鏈對兩種 mode 都適用)。
- **決策**:維持 web server + 瀏覽器架構,用 Chromium 系瀏覽器 `--new-window --app=<URL>` 開**無邊框獨立視窗**(外觀等同原生 app),不開普通分頁。
- **Fallback 鏈**:lookpath 候選(google-chrome / chromium / microsoft-edge / brave-browser 等)→ flatpak(`flatpak list --app` 匹配 Brave/Chromium/Chrome/Edge,用 `flatpak run`)→ `steam://openurl/<url>`(Steam 在 Deck 上保證存在,用其內建 CEF 瀏覽器;原廠無瀏覽器時的關鍵 fallback)→ `xdg-open`。
- **Flags**:`-no-browser`(只印 URL,不開視窗)、`-browser <path>`(強制指定瀏覽器)。
- **不採用(已評估)**:
  - WebView 內嵌(Wails/webview/webkit2gtk):Stock SteamOS 無 webkit2gtk(Valve issue #1851)且 rootfs 唯讀,依賴安裝不可行;需 CGO。
  - Fyne/Gio 原生 UI:dev VM 無 GPU 無法驗證;丟失 React UI 與 Phase 7 Decky 重用。
  - Electron/CEF:二進位 100MB+ 過重。
- **保持**:`CGO_ENABLED=0` 靜態單檔、token 安全、React UI 供 Decky 重用。
- **範圍外**:單實例鎖、Firefox 支援(走 xdg-open fallback)。

## Naming (confirmed)
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
- **Build CLI**: `cd cmd/cryoutilities && CGO_ENABLED=0 go build -o cryoutils-ng .`
- **Build desktop server**: `cd web && npm ci && npm run build` then `cp -r web/dist cmd/desktop/web/dist` then `cd ../cmd/desktop && CGO_ENABLED=0 go build -o cryoutils-ng-desktop .`
- **Build web UI**: `cd web && npm ci && npm run build`
- **Test**: `cd core && CGO_ENABLED=0 go test ./...` · `cd core && CGO_ENABLED=0 go vet ./...` · `cd cmd/desktop && CGO_ENABLED=0 go vet ./...`
- **Note**: `go build ./...` and `go vet ./...` at repo root fail due to Fyne GL dependency in `internal/` (no GPU in dev env); build each module separately.
- **Run server (dev)**: `go run ./cmd/desktop` → opens `http://127.0.0.1:<port>/?token=...`
- **CLI (target)**: `sudo ~/.cryoutils_ng/cryoutils-ng <command> [parameter]`
- **Permissions**: tweaks need sudo; `core/sudo.go` handles auth:
  - CLI 以 `sudo` 執行時（`os.Geteuid() == 0`）：直接跳過 `RenewAuth()`
  - Web UI 執行時（deck 使用者）：`RenewAuth()` 使用儲存的密碼 → `sudo -S -k -- echo`（`-k` forces cache invalidation）
  - Web UI 已移除密碼輸入，所有操作直接執行（token 提供 API 安全層）

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

## Known Issues & Fixes
- **Sudo cached timestamp bypass** (`core/sudo.go`): `sudo -S` 有 cached timestamp 時會忽略 stdin 密碼導致 `TestAuth` 永遠通過。修正：所有 `sudo` 指令加 `-k` flag 強制忘記 cache，`RenewAuth()` 回傳 `error` 並檢查空密碼。
- **CLI sudo 執行時 RenewAuth 失敗** (`core/sudo.go`): CLI 以 `sudo` 執行時 `e.Password == ""`，`RenewAuth()` 因檢查空密碼而失敗，但實際操作會成功（已是 root）。修正：`RenewAuth()` 新增 `os.Geteuid() == 0` 檢查，已為 root 時直接跳過。
- **Web UI 密碼層多餘** (`web/src/App.tsx`): Steam Deck 環境預設信任同機使用者，token 已提供 API 安全。修正：移除密碼輸入 UI 與 `sudoLocked` 狀態。
- **Swap file location bug** (`core/swap.go`): `/proc/swaps` 遇到 `/dev/zram0` 時應 `continue` 跳過，而非 `return error`。修正後正確讀取 `/home/swapfile`。
- **Log directory auto-creation** (`cmd/cryoutilities/main.go` + `cmd/desktop/main.go`): 啟動時自動 `os.MkdirAll(core.InstallDirectory, 0755)` 建立目錄，避免 `sudo` 下 `os.UserHomeDir()` 回傳 `/root` 時找不到 log 檔。
- **Status command stdout** (`cmd/cryoutilities/main.go`): `printStatus()` 新增 `fmt.Printf` 輸出到 stdout，log 檔仍保留 `InfoLog` 輸出。

## Testing & Verification
- Unit tests in `core/` (ported from `internal/util_test.go`).
- Status verified via CLI `status` command or `GET /api/status`.
- Headless screenshots at 4K + 1280×800; CLI smoke (`status`/`recommended`/`stock`).
- Final: real Steam Deck acceptance, verify original CryoUtilities untouched.

## Roadmap
- **Phase 0.5**: CI/CD infrastructure (Dependabot for gomod + npm, auto-merge patch/minor) — see `todo.md`
- **Phase 0**: dependency modernization (Go 1.26.5, fresh `core/go.mod`, drop Fyne) — see `todo.md`
- **Phase 1**: core extraction ✅
- **Phase 2**: CLI rework + `status` command ✅
- **Phase 3**: desktop web server (REST + SSE + token) ✅
- **Phase 4**: single-page React UI ✅
- **Phase 5**: packaging (install.sh, .desktop, launcher.sh, uninstall.sh) ✅
- **Phase 5.5**: desktop UI launch — chromeless app window (方案 D, see above) ✅
- **Phase 5.6**: i18n 多國語言支援 + 版本號 `v2.2.2` → `v0.1.0` ✅ — 自訂輕量 i18n hook（Context + JSON），單一來源 `locales.ts`，零程式碼變更新增語言
- **Phase 6**: verification (user — requires real Steam Deck)
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
