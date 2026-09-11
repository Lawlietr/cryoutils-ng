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
  - **SteamOS 端禁止開發**（改程式碼/建置/測試，一律在 Dev VM 做）；但**直接修改 repo 內容（如文件）並 commit/push 是允許的**（2026-09-11 用戶澄清）
- **二進位複製方向**：SteamOS 未啟用 sshd，Dev VM 無法 SSH 到 SteamOS。所有二進位複製都由 **SteamOS 主動執行 `scp` 從 Dev VM 拉取**（`scp root@192.168.1.15:/path/to/binary ~/.cryoutils_ng/`）。Dev VM 端編譯完成後只需報告結果，不嘗試推送。
- **Go 路徑**：`go` 不在 agent 的 PATH 中（`.bashrc` 對非互動式 shell 無效）。**所有 Go 命令必須使用完整路徑**：`/usr/local/go/bin/go`。例如：`/usr/local/go/bin/go build -o cryoutils-ng ./cmd/cryoutilities`。
- **Git 規範（MANDATORY）**：
  - **`.gitignore` 是單一來源**：任何不應該入版控的檔案，必須同時出現在 `.gitignore` 中。若發現檔案被追蹤但不應追蹤，**先更新 `.gitignore`，再執行 `git rm --cached <path>` 移除追蹤，然後才能 commit**。
  - **禁止追蹤的檔案類型**：
    - `node_modules/`（所有 Node.js 依賴）
    - 編譯產物：`cryoutils-ng`、`cryoutils-ng-desktop`、`*.exe`
    - 測試結果：`web/test-results/`
    - 日誌檔：`cryoutilities.log`、`cryoutils_ng.log`
  - **commit 前必做檢查**：執行 `git status --short` 與 `git diff --cached --stat`，確認沒有上述禁止類型。若發現違規檔案，停下並回報，不要直接 commit。
  - **已知歷史錯誤**：2025-07-17 曾因 `.gitignore` 缺少 `node_modules/` 規則，導致 3707 筆 `web/node_modules/` 檔案被錯誤追蹤。已修正 `.gitignore` 並從 git 移除，此條規則用於防止再次發生。

## Architecture & Structure
- **`core/`**: UI-independent Go engine (single source of truth for all tuning logic). `core.Engine` struct (loggers, sudo password, `OnProgress` callback) replaces the old global `CryoUtils`. No Fyne, no CGO (`CGO_ENABLED=0` static binary).
- **`cmd/cryoutilities`**: CLI. All original subcommands + new `status` command (scripting interface for the future Decky backend). Binary name: `cryoutils-ng`. Published via GitHub Releases.
- **`cmd/desktop`**: `-ui web|native`（預設 `web`，Phase 6.5 起）：web 模式 = localhost web server (127.0.0.1 + random token), REST API, SSE progress, `go:embed` of the web build；native 模式 = Fyne 原生 UI 直接 in-process 跑 `core.Engine`（無 HTTP/token/瀏覽器）。Binary name: `cryoutils-ng-desktop`（Phase 6.5 起需 `CGO_ENABLED=1`）。Not published via Releases — users build from source per README instructions.
- **`ui/fyneui`**: Fyne 原生 UI（Phase 6.5）— dark theme、單頁直式無 tab、與 web UI 六區塊同等、嵌入 Noto Sans TC、runtime drop-in 語言。根布局用 `container.NewBorder(top, nil, nil, nil, vScroll)`（**不可用 VBox 包 VScroll** — VBox 會把 Scroll 壓成 MinSize 常數高度）；5s ticker/Refresh 走 `refreshers` registry 做 in-place 值更新（不重建 tree，保留用戶選擇與 scroll 位置），只有語言切換才全量重建；sudo 密碼對話框在 `sudo.go` 的 `authManager`（`sudo_dialog.go` 零呼叫死代碼已於 Phase 6.6.1 刪除，2026-09-11）；`authManager` 有 `dialogOpen` flag + `pending` 佇列防對話框疊開。**Phase 6.6（4K 適配，dev 端完成 2026-09-11）**：FYNE_SCALE 自適應（啟動前自 re-exec，因 Fyne v2.7.4 無 `SetScale()` API）、內容欄寬度上限 1100px + 置中（`layout.go` `cappedCenterLayout`）、視窗 80% 螢幕÷scale + 置中（`windowSizeForScreen`）、Status 雙欄（`NewGridWithColumns`）、section Card 化（`widget.NewCard`，v2.7.4 為 3 參）。詳見 `docs/fyne-ui-plan.md` 與 `todo.md` Phase 6.6。
- **`i18n`**: locale 單一來源（Phase 6.5）— `//go:embed i18n/locales/*.json` + runtime `~/.cryoutils_ng/locales/` 掃描覆蓋（免重編）；`web/src/locales/` 為生成物（`npm run sync-loc`）。
- **`web/`**: React + Vite + TypeScript **single-page** UI (no tabs). Vertical single column; all settings + statuses on one view. Plain CSS, framework-agnostic components (reusable by Decky plugin later).
- **`internal/`**: legacy Fyne UI — **Phase 6.5 待刪**（新 Fyne UI 通過真實 Deck 接受後，單獨 commit 移除）；移除前它是 Fyne 版本升級的編譯基準（v2.3.1 可編）。詳見 `docs/fyne-ui-plan.md`。

## UI Design Decisions (locked)
- **Single page, vertical single column** — no tabs (original's tabs deemed redundant).
- **No password unlock** — Steam Deck 環境預設信任同機使用者；token 已提供 API 安全層。
- **Responsive**: `clamp()` fluid fonts + `max-width` + compact breakpoint. Must render correctly at **3840×2160 (4K)** and **1280×800 (Steam Deck native)**.
- **VRAM section is read-only** (actual VRAM change is via BIOS).
- Long tasks (swap resize) show a top-of-page progress bar via SSE, never lost on scroll.
- Headless browser (Playwright) verification at both resolutions on dev box; final visual acceptance on real Deck.
- **Native Fyne UI（Phase 6.5）**：第二種 UI 路徑（`-ui native`）；**dark theme**、同樣單頁直式版面、嵌入 Noto Sans TC、**runtime drop-in 語言切換**（`~/.cryoutils_ng/locales/*.json` 免重編）。詳見 `docs/fyne-ui-plan.md`。
- **sudo（native）**：首啟 / 失敗密碼對話框；密碼只存 `e.Password`（記憶體），**不落磁碟**。**保留啟動問密碼**（用戶決策 2026-09-11：部分狀態讀取走 `sudo cat`，非 root 看不到，不做 pure lazy）；Phase 6.6.1 已加 `authManager` 單對話框 flag + pending 佇列防疊開（啟動對話框 200ms 延遲 vs 操作對話框會導致雙問），dev 端完成 2026-09-11，真機驗證待用戶驗收。
- **scale（native，Phase 6.6，dev 完成 2026-09-11）**：Fyne v2.7.4 無運行時 `SetScale()`，scale 於 app 初始化從 `FYNE_SCALE`/`Xft.dpi` 讀定 → 自適應用「啟動前自 re-exec」（`cmd/desktop/scale.go`）：`FYNE_SCALE` 與 `CRYOUTILS_FYNE_SCALE` 皆未設時讀 xrandr 解析度算 scale（clamp(寬/1280, 1.0, 2.0)，0.25 倍數）後帶 **`FYNE_SCALE`（Fyne 實際讀取的變數）+ `CRYOUTILS_FYNE_SCALE`（防迴圈 guard + UI 邏輯尺寸用）** 重啟；scale==1.0 跳過 re-exec；`FYNE_SCALE` 手動覆蓋優先。xrandr 解析用 `current WxH` regex（**非行首 `*`** — xrandr 的 `*` 在行尾）。視窗尺寸：`windowSizeForScreen()` = 80% 螢幕（physical px）÷ scale + `CenterOnScreen()`；Deck 原生 ≤1280×800 維持 full screen。
- **sudo（web，已知缺口）**：P0.5 已完成：密碼輸入 UI 與 `POST /api/auth` 端點皆已移除 → web 模式 `e.Password` 恒空：以 root 執行時無影響（`RenewAuth()` 走 `Geteuid` 跳過）；非 root web 模式需權限的操作會失敗。主要路徑是 native UI（有密碼對話框）；是否補回 web 輸入是獨立議題。

## Desktop UI Launch (Decision — 方案 D)
- **使用環境**:UI 主要在 Steam Deck **Desktop Mode** 使用(Gaming Mode 非主要;但 fallback 鏈對兩種 mode 都適用)。
- **決策**:維持 web server + 瀏覽器架構,用 Chromium 系瀏覽器 `--new-window --app=<URL>` 開**無邊框獨立視窗**(外觀等同原生 app),不開普通分頁。
- **Fallback 鏈**:lookpath 候選(google-chrome / chromium / microsoft-edge / brave-browser 等)→ flatpak(`flatpak list --app` 匹配 Brave/Chromium/Chrome/Edge,用 `flatpak run`)→ `steam://openurl/<url>`(Steam 在 Deck 上保證存在,用其內建 CEF 瀏覽器;原廠無瀏覽器時的關鍵 fallback)→ `xdg-open`。
- **Flags**:`-no-browser`(只印 URL,不開視窗)、`-browser <path>`(強制指定瀏覽器)。
- **不採用(已評估)**:
  - WebView 內嵌(Wails/webview/webkit2gtk):Stock SteamOS 無 webkit2gtk(Valve issue #1851)且 rootfs 唯讀,依賴安裝不可行;需 CGO。
  - Fyne/Gio 原生 UI:dev VM 無 GPU 無法驗證;丟失 React UI 與 Phase 7 Decky 重用。(Phase 6.5 後已重估:Fyne 原生 UI 正式採用,見下)。
  - Electron/CEF:二進位 100MB+ 過重。
- **保持**:CLI `CGO_ENABLED=0` 靜態單檔、token 安全、React UI 供 Decky 重用。
- **範圍外**:單實例鎖、Firefox 支援(走 xdg-open fallback)。
- **2026-08 補充(Phase 6.5)**:真實 Deck 驗證 `--app=` 失敗(拒絕連線,見 todo.md Deck 測試結果)→ 新增 **Fyne 原生 UI** 為主要使用路徑(`-ui native`,施工手冊 `docs/fyne-ui-plan.md`);web 模式保留供 Decky(Phase 7)+ fallback。`-ui` 預設值待 native 通過 Deck 接受後才切 `native`。

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
- **Build desktop（Phase 6.5 起,Fyne 需 CGO）**: `cd web && npm ci && npm run build` then `cp -r web/dist cmd/desktop/web/dist` then `cd ../cmd/desktop && CGO_ENABLED=1 go build -o cryoutils-ng-desktop .`
- **Build web UI**: `cd web && npm ci && npm run build`
- **Test**: `cd core && CGO_ENABLED=0 go test ./...` · `cd core && CGO_ENABLED=0 go vet ./...` · `cd cmd/desktop && CGO_ENABLED=1 go vet ./...`
- **Note**: Phase 6.5 起 root module `go build ./...` / `go vet ./...` 需 `CGO_ENABLED=1`（Fyne：`ui/fyneui/` + legacy `internal/`）；`core/` 與 CLI 維持 `CGO_ENABLED=0` 建置/測試。Headless 驗證：`Xvfb` + `scrot`（見 `docs/fyne-ui-plan.md` §5）。
- **Run server (dev)**: `go run ./cmd/desktop` → opens `http://127.0.0.1:<port>/?token=...`
- **CLI (target)**: `sudo ~/.cryoutils_ng/cryoutils-ng <command> [parameter]`
- **Permissions**: tweaks need sudo; `core/sudo.go` handles auth:
  - CLI 以 `sudo` 執行時（`os.Geteuid() == 0`）：直接跳過 `RenewAuth()`
  - Web UI 執行時：密碼輸入已移除（P0.5 完成，含 `/api/auth` 端點），`e.Password` 恒空；以 root 執行時正常（`Geteuid` 跳過），非 root web 模式需權限操作會失敗（已知缺口，見 Known Issues）
  - Fyne 原生 UI 執行時（deck 使用者）：首啟 / 失敗對話框取得密碼 → `e.Password`（只存記憶體，不落磁碟）；`RenewAuth()` 失敗時重彈

## Dependencies (modernized)
- **Kept + updated**: `golang.org/x/sys@v0.47.0`, `mountinfo@v0.7.2`, `otiai10/copy@v1.14.1`, `acmd@v0.12.0`, `vdf@v1.1.0` (already latest)
- **Removed**: `fyne.io/fyne/v2` and its entire stale indirect tree (glfw, gopherjs, gl-js, oksvg, rasterx, textlayout, freetype, systray, x/mobile, x/image, x/net, x/text, fsnotify) — most pinned 2021–2022
- **Go directive**: `go 1.26` in `core/go.mod`
- **Root module（Phase 6.5）**: `fyne.io/fyne/v2` v2.3.1 → **v2.7.4**（新 Fyne 原生 UI；v2.8.0 為 PoC 後可選評估）。`internal/` 刪除後 Fyne 仍留在 root module（`ui/fyneui/` 在用）。

## Important Constraints & Quirks
- **Swap Files**: only supports swap files, not swap partitions.
- **Coexistence**: new project must not overwrite original install. Both apps tune the same `/proc/sys` params and `/etc/tmpfiles.d/*.conf` — install-level coexistence is fine; system tuning state is "last one wins" (expected, not a conflict).
- **Platform**: optimized for SteamOS. Swap resize disables swap briefly — risky; handle with warnings. Decky/swap-resize safety gating deferred to Phase 7.
- **VRAM read** depends on `glxinfo` (mesa-utils) — original app already validated this on SteamOS.
- **Security**: desktop server binds 127.0.0.1 only + per-launch random token (prevents other local processes driving privileged ops).

## Known Issues & Fixes
- **Native UI sudo 多次詢問**（2026-09-11 真機發現，**已修正（2026-09-11，dev 端）**，真機驗收待用戶）：非 root 執行時，用戶回報啟動輸入一次後操作又被問一次。成因：(a) 啟動對話框延遲 200ms，用戶先點操作時 `runTask → ensure` 再開第二個對話框 → 疊開雙問；(b) `ui/fyneui/sudo_dialog.go` 為零呼叫死代碼，容易改錯路徑。修正：`authManager` 單對話框 flag + pending 佇列 + 刪死代碼；**啟動問密碼保留**（部分狀態讀取需 root）。另 Fyne v2.7.4 無 `SetScale()` API，4K scale 走 `FYNE_SCALE` env + 自 re-exec（Phase 6.6.2，dev 端完成）。
- **`.gitignore` 遺漏 `node_modules/`**（2025-07-17，已修正）：`.gitignore` 缺少 `node_modules/` 與 `web/test-results/` 規則，導致 3707 筆 `web/node_modules/` + 1 筆測試結果檔案被錯誤追蹤並 commit。修正方式：在 `.gitignore` 加入 `node_modules/` 與 `web/test-results/`，執行 `git rm -r --cached web/node_modules web/test-results` 移除追蹤。此問題已記錄於 Working Rules 的 Git 規範中。
- **Sudo cached timestamp bypass** (`core/sudo.go`): `sudo -S` 有 cached timestamp 時會忽略 stdin 密碼導致 `TestAuth` 永遠通過。修正：所有 `sudo` 指令加 `-k` flag 強制忘記 cache，`RenewAuth()` 回傳 `error` 並檢查空密碼。
- **CLI sudo 執行時 RenewAuth 失敗** (`core/sudo.go`): CLI 以 `sudo` 執行時 `e.Password == ""`，`RenewAuth()` 因檢查空密碼而失敗，但實際操作會成功（已是 root）。**已修正（P0.5）**：`RenewAuth()` 新增 `os.Geteuid() == 0` 檢查，已為 root 時直接回傳 `nil`。
- **Web UI 密碼層多餘** (`web/src/App.tsx`): Steam Deck 環境預設信任同機使用者，token 已提供 API 安全。**已修正（P0.5）**：移除密碼輸入 UI、`sudoLocked` 狀態與全部 `disabled={sudoLocked}` 條件，並同步移除 `web/src/api.ts` 的 `auth()`、`cmd/desktop/api.go` 的 `/api/auth` 端點、相關 CSS 與 locale 鍵。
- **Swap file location bug** (`core/swap.go`): `/proc/swaps` 遇到 `/dev/zram0` 時應 `continue` 跳過，而非 `return error`。修正後正確讀取 `/home/swapfile`。
- **Log directory auto-creation** (`cmd/cryoutilities/main.go` + `cmd/desktop/main.go`): 啟動時自動 `os.MkdirAll(core.InstallDirectory, 0755)` 建立目錄，避免 `sudo` 下 `os.UserHomeDir()` 回傳 `/root` 時找不到 log 檔。
- **Status command stdout** (`cmd/cryoutilities/main.go`): `printStatus()` 新增 `fmt.Printf` 輸出到 stdout，log 檔仍保留 `InfoLog` 輸出。
- **Web 模式 sudo 入口缺口**（已知缺口）：P0.5 移除密碼輸入 UI，`POST /api/auth` 端點也已刪除；web 模式 `e.Password` 恒空 → **非 root web 模式** `RenewAuth()` 失敗（需權限的操作會報錯）；web 以 root 執行時無影響（`Geteuid` 跳過）。Fyne 原生 UI（Phase 6.5）有自己的首啟/失敗密碼對話框，不受影響；是否補回 web 入口是獨立議題。
- **ZRAM 支援** (`core/swap.go`): SteamOS 3.6+ 使用 zram swap（`/dev/zram0`）與 swap file 並存。`ChangeSwapSize()` 執行 `swapoff -a` 後會主動重新啟用 zram（`swapon /dev/zram0`），避免等待 systemd 裝置掃描（約 18 分鐘）。`GetZramStatus()` 讀取 `/proc/swaps`，`GetZramSizeBytes()` 讀取 `/sys/block/zram0/disksize`，`getTotalSwapGB()` 從 `/proc/meminfo` 的 `SwapTotal` 計算總 swap（zram + swap file）。

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
- **Phase 6**: 統合真機驗證（重排到 Phase 6.5 完成後執行；含 P0 zram、CLI 全命令、web fallback、native 驗收，見 `todo.md`）
- **Phase 6.5**: Fyne 原生 UI — 全新 `ui/fyneui/`（dark、單頁、i18n drop-in、sudo 對話框）+ `-ui web|native`；施工手冊 `docs/fyne-ui-plan.md`。**dev 端完成（2026-08-25）**：Fyne v2.7.4、i18n 單一來源、六區塊 + in-place refresh、headless OCR 驗證通過（1280×800 × en/zh-TW，2560×1440 待測）；剩真機驗收（Step 11）、打包（Step 12）、`internal/` 刪除（Step 13，單獨 commit）
- **Phase 6.6**: Native UI 4K 適配與 sudo 流程修正（2026-09-11 真機 feedback）— 依序：(1) sudo 雙問修正（保留啟動問 + `authManager` 防疊開 + 刪 `sudo_dialog.go` 死代碼）(2) FYNE_SCALE 自適應（自 re-exec）(3) 4K 排版（內容欄寬度上限置中、視窗跟隨螢幕、Status 雙欄、Card 化）(4) headless + 真機雙解析度驗證。**dev 端 + headless 雙解析度驗證完成（2026-09-11）**，剩真機驗收（用戶 scp binary 上 Deck，4K + 1280×800）。見 `todo.md` Phase 6.6
- **Phase 7 (future)**: Decky Loader plugin — React frontend reused + Python shim calling CLI binary (`main.py`, `plugin.json`, distribution zip; `backend/src → backend/out → bin/` CI convention)

## License & Usage Rights
- **License**: GNU General Public License v3.0 (GPLv3) — inherited from the original CryoUtilities project.
- **Derivative Work**: this is a derivative work; must retain original copyright notices/license declarations, be distributed under GPLv3 (or compatible), and provide complete source code to recipients.
- **What You Cannot Do**: incorporate into proprietary/closed-source software; remove or alter the GPLv3 license terms.
- **UI Rewrite Intent**: core logic (handlers/config/utilities) preserved and rewritten in Go; UI rewritten independently (single-page web, 加上新單頁 Fyne 原生 UI（Phase 6.5）；兩者皆與原 5-tab Fyne UI 設計不同)。

## Attribution & Documentation
- **Original Project**: CryoUtilities by CryoByte33 (unmaintained 1+ year) — this is a **rewrite** that honors and continues the original work.
- **README Policy**: 推上 GitHub 前必須重寫 `README.md`：
  - **移除**原作者個人資訊（YouTube、Patreon、Discord、個人網站）
  - **保留**對 CryoByte33 / CryoUtilities 的致敬與衍生作品聲明
  - **更新**所有描述符合當前專案實際狀態
