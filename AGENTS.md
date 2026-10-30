# CryoUtils NG — AGENTS.md

## Project Overview
- **Purpose**: A Steam Deck utility to manage swap files, swappiness, and memory parameters — a rewrite of the original CryoUtilities (unmaintained for 1+ year).
- **Original maintainer status**: Original project by CryoByte33 is no longer maintained. This is a **rewrite with a new name** that must **coexist** with the original install (never overwrite `~/.cryo_utilities`, `cryo_utilities` binary, `cryoutilities_steam_data`, or its `.desktop` files).
- **Root Directory**: `/root/opencode-stuffs/steam-deck-utilities`
- **Dev environment**: Ubuntu 24.04 x86_64 (Proxmox **LXC container**, hostname `Code-Agent-01`, `a1` / `192.168.1.15`). Steam Deck target is also x86_64 → native compile, no cross-compile. **Toolchain 現況（2026-10-20）**：Go **1.26.10** 已裝在 `/usr/local/go`（repo `go 1.26` directive 對齊；舊記錄的 1.26.5 已不存在）；**Node 未裝**（要重刻 web UI 時才需要）。SteamOS needs **zero** dev toolchains — it only receives a prebuilt binary.
- **Dev VM 從 SteamOS 驅動**（2026-10-20 實測可行，免切 session）：`ssh -o BatchMode=yes -o IdentitiesOnly=yes -i ~/.ssh/id_centos7-proxy01 root@192.168.1.15`。**必頇帶 `IdentitiesOnly=yes`**，否則 ssh 先丟其他 key 被 server 拒絕，會誤判成免密不可用。
- **Dev VM 無法做「真 swap」端到端測試**：它是 LXC（核心 `7.0.0-3-pve`，`modprobe zram` 找不到 kernel modules；`/dev/loop-control` 不存在，無法用 loop+ext4 繞過），且 `/home` 在 **ZFS** 上（`swapon: ... it appears to have holes`）。swap 相關改動在 Dev VM 只用**命令插樁（fake PATH）單測**驗證，端到端**只能在 Deck 真機驗**。

## Working Rules (MANDATORY)
- **禁止自行 commit / push**：任何 `git commit`、`git push`、`git tag` 必須等你明確指示才執行。你會在 Dev VM 與 SteamOS 之間頻繁切換測試，擅自推送可能覆蓋你的本地修改。
- **工作前先確認環境**：每次任務開始時必須先用 `cat /etc/os-release | grep PRETTY_NAME` 確認當前在哪個環境，並回報給用戶。
  - Dev VM (`a1` / `root@192.168.1.15`，hostname `Code-Agent-01`)：Ubuntu 24.04.5，Go 1.26.10（`/usr/local/go/bin/go`）+ gcc（Fyne 的 CGO 需要）；**Node 目前未裝**（要重刻 web UI 時再裝）
  - SteamOS：rootfs 唯讀，無開發工具鏈，無 sshd，只接受預編譯二進位
  - **SteamOS 端禁止開發**（改程式碼/建置/測試，一律在 Dev VM 做）；但**直接修改 repo 內容（如文件）並 commit/push 是允許的**（2026-09-11 用戶澄清）
- **二進位複製方向**：SteamOS 未啟用 sshd 時，Dev VM 無法 SSH 到 SteamOS，所有二進位複製都由 **SteamOS 主動執行 `scp` 從 Dev VM 拉取**（`scp root@192.168.1.15:/path/to/binary ~/.cryoutils_ng/`）。Dev VM 端編譯完成後只需報告結果，不嘗試推送。
- **SteamOS sshd（2026-10-30 起可用）**：用戶已在真機開啟 sshd（`deck@192.168.1.232`）。**密碼存在 repo 根的 `.secrets`**（mode 600、已列入 `.gitignore` 的 `/.secrets`，**絕不 commit**；AGENTS.md 與任何文件都不寫密碼）。sshd 開著時**查證與佈署方向反過來**：Dev VM 直接 SSH 進 Deck 讀 log、跑 `status` A/B、跑驗收用 probe（`tools/p09probe/`，gitignored），二進位也可直接推到 `~/.cryoutils_ng/`（**佈署前必先 `cp -a` 備份成 `*.bak-p08` 之類的標籤**）。sshd 關掉時回到上一條的拉取方向。
  - **在 Deck 上改過的狀態要收尾**：驗收過程中翻轉的 kernel 參數、`/etc/tmpfiles.d/*.conf`、log 權限等，一併還原或標明，並在回報中列出。
- **Go 路徑**：`go` 不在 agent 的 PATH 中（`.bashrc` 對非互動式 shell 無效）。**所有 Go 命令必須使用完整路徑**：`/usr/local/go/bin/go`。例如：`/usr/local/go/bin/go build -o cryoutils-ng ./cmd/cryoutilities`。
- **特權命令單一入口（MANDATORY）**：所有需要 root 的外部命令一律走 `core.Engine.runSudo()` / `runSudoWithInput()`（`core/sudo.go`），**禁止**直接 `exec.Command("sudo", ...)`。理由不是風格：沒有 `-S` 的 sudo 會從 tty 問密碼，桌面 UI 因此變成「終端問一次 + 對話框再問一次」（2026-10-20 真機 feedback），而且 `.Output()` 丟掉 stderr 讓失敗無法查證。`runSudo` 在沒有密碼時直接回 `ErrNoSudoPassword`、**不會啟動 sudo**，所以也永不碰 tty。
  - **附加三條鐵則（P0.9，2026-10-30；少任何一條都會把系統參數寫壞、或把密碼寫進 log）**：
    1. **sudo 一律 `-S -k`**。sudo 只在「真的需要驗證」時讀 stdin；timestamp 還有效時它完全不讀，**整段 payload（第一行是密碼）會被當資料餵給命令** → `tee /sys/kernel/mm/transparent_hugepage/enabled` 把密碼寫進 sysfs → kernel EINVAL → `tee` exit 1 → 五個記憶體參數全部「不可用」。迴歸測試：`core/sudo_test.go` `TestRunSudoWithInputNeverLeaksPasswordIntoTheCommand`（用重現 sudo stdin 契約的 fake sudo）。
    2. **stdout / stderr 分流**。sudo 的密碼提示寫在 stderr，兩條流合併後讀值會變成 `[sudo] password for deck: 60`，永遠比對不上預期值。錯誤只帶 stderr（`tee` 的 stdout 是把寫入內容 echo 回來，不該進 log）。測試：`TestRunSudoReturnsStdoutOnly`。
    3. **任何要進 log / 進 UI 的字串先過 `e.redact()`**。`tee` 會 echo 其輸入，而該輸入第一行曾是 sudo 密碼 → Deck 的 log 一度存有**明文 sudo 密碼 5 處**（已在真機 scrub、log 權限改 0600；兩支 main 的 `os.OpenFile` mode 從 0666 改 0600）。
  - **讀狀態不要要 root**：unit 矩陣六個參數（`/proc/sys/vm/*`、`/sys/kernel/mm/transparent_hugepage/*`）在 Linux 上都是 `0644 root:root`，讀取一律直接 `os.ReadFile`，只有 `EPERM` 才退回 sudo（`core/util.go` `getUnitStatus`）。把讀取塞進 sudo 的後果：沒有密碼時（web 模式無密碼入口）全部顯示 false/unknown，且每 5 秒刷新會為每個參數寫一行失敗 log。
  - **procfs/sysfs 的檔案不能 unlink**（即使是 root）：`rm /proc/sys/vm/nr_hugepages` 必然 EPERM。要改數值只能寫入，不能刪除。
- **UI 即時更新規則**：native UI 的 status 卡片每 5 秒整塊重建；其他卡片（含下拉選單、輸入框）**必須**透過 `uiCtx.addRefresher()` 註冊就地更新，否則數值會 freeze 在啟動當時（swap 下拉的教訓，見 Known Issues）。
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
- **Platform**: optimized for SteamOS. Swap resize disables swap briefly — risky; handle with warnings. **只允許關自己的 swap file（`swapoff <path>`），絕不可用 `swapoff -a`**：後者會让 systemd 摧毀 zram 裝置（見 P0.7）。Decky/swap-resize safety gating deferred to Phase 7.
- **VRAM read** depends on `glxinfo` (mesa-utils) — original app already validated this on SteamOS.
- **Security**: desktop server binds 127.0.0.1 only + per-launch random token (prevents other local processes driving privileged ops).

## Known Issues & Fixes
- **Native UI sudo 多次詢問**（2026-09-11 真機發現，**已修正並已真機驗證通過（2026-10-20）**）：非 root 執行時，用戶回報啟動輸入一次後操作又被問一次。成因：(a) 啟動對話框延遲 200ms，用戶先點操作時 `runTask → ensure` 再開第二個對話框 → 疊開雙問；(b) `ui/fyneui/sudo_dialog.go` 為零呼叫死代碼，容易改錯路徑。修正：`authManager` 單對話框 flag + pending 佇列 + 刪死代碼；**啟動問密碼保留**（部分狀態讀取需 root）。另 Fyne v2.7.4 無 `SetScale()` API，4K scale 走 `FYNE_SCALE` env + 自 re-exec（Phase 6.6.2，dev 端完成）。
- **`.gitignore` 遺漏 `node_modules/`**（2025-07-17，已修正）：`.gitignore` 缺少 `node_modules/` 與 `web/test-results/` 規則，導致 3707 筆 `web/node_modules/` + 1 筆測試結果檔案被錯誤追蹤並 commit。修正方式：在 `.gitignore` 加入 `node_modules/` 與 `web/test-results/`，執行 `git rm -r --cached web/node_modules web/test-results` 移除追蹤。此問題已記錄於 Working Rules 的 Git 規範中。
- **啟動時問兩次 sudo 密碼**（native UI，2026-10-20 真機回報，**已修正並已真機驗證通過 2026-10-20**）：`core/util.go` 殘留四處舊的 `exec.Command("sudo", ...)`（無 `-S`），sudo 直接從 tty 問密碼；啟動時 status 卡片 → `getUnitStatus()` 的 `sudo cat /proc/sys/...` 就會觸發終端提示，接著 UI 對話框走 `RenewAuth()`（刻意 `sudo -k`、不看 cache）再問一次。修法：這四處（`getUnitStatus`、`setUnitValue`、`replaceFile` 的 `sudo mv`、`removeFile` 的 `sudo rm`）改走 `runSudo()`/`runSudoWithInput()`，且沒有密碼時回 `ErrNoSudoPassword` 不啟動 sudo。真機 A/B（pty）：控制組 bare sudo 出現 `[sudo] deck 的密碼：`，新 CLI 以 deck 非 root 跑 `status` 完全無提示（132 ms）。**啟動問一次（UI 對話框）仍是鎖定的設計**，不是 bug。
- **native UI swap 下拉不跟隨真實大小**（2026-10-20 真機回報，**已修正並已真機驗證通過**）：status 卡片每 5 秒重建，swap 卡片只建立一次，`GetAvailableSwapSizesStr()` 的結果與 `SetSelected` 從啟動就 freeze，改完 size 後 "N - Current Size" 仍是舊值，要重開 app。修法：`uiCtx.refreshers` 註冊表（5 秒 ticker 與操作完成後都跑）+ `Select.SetOptions` + `nextSwapSelection()`（保留使用者明確挑選的項目，偽項目 "N - Current Size" 跟隨真實大小）；swappiness placeholder 一併跟隨。
- **status 刷新造成的 log 噪音與無限膨脹**（**未修**）：native UI 每 5 秒刷新，status 讀取會寫進 log（`Found a swap file with a size of …`、`Found a swappiness of …`、`Found a VRAM of …`），swap 卡片的 refresher 又觸發一次 `GetAvailableSwapSizesStr()`（多兩行）。2026-10-20 真機實測：43 分鐘寫了 721 行（約 17 行/分、保守估計 1–2 MB/日），且 log 只 append、沒有 rotation。兩件事要處理：(1) 這些 poll 級別的 INFO 應該降噪（只在值變化時寫、或加 `-verbose` 開關）；(2) `GetAvailableSwapSizes()` 不該在 refresher 裡重複抓 system state（可用 `uiCtx.lastStatus` 的 `SwapSizeGB` + 較低頻率的空間檢查）。
- **unit 參數（hugepages / shared memory in THP / compaction / defrag / page_lock_unfairness）不可偵測也不可用**（P0.9，2026-10-30 真機回報，**已修正並已在 Deck 真機驗證**）：兩個獨立缺陷疊在一起。(a) **寫入壞**：`runSudoWithInput` 用 `sudo -S`（**沒有 `-k`**），而 `RenewAuth`/`TestAuth` 剛認證過、timestamp 還有效 → sudo 完全不讀 stdin → **密碼那一行連同值一起被 `tee` 寫進 sysfs/procfs**，kernel 回 EINVAL、`tee` exit 1 → 每次 toggle 都回報錯誤（UI 彈錯誤對話框）。Deck log 原話：`tee /sys/kernel/mm/transparent_hugepage/enabled: exit status 1: <密碼>\nalwaystee: ... 不適用的引數`。(b) **讀取壞**：`getUnitStatus()` 走 sudo，沒有密碼（web 模式永遠沒密碼；native 尚未解鎖時也沒有）時直接失敗 → `GetStatusSummary()` 全報 false，UI 顯示成「關閉/unknown」。另有兩個小項：`rm` 沒帶 `-f`（revert 一個不存在的 tmpfiles.d conf 會被當失败）、`SetHugePages` 試圖 unlink `/proc/sys/vm/nr_hugepages`（procfs 不能 unlink，root 也 EPERM）。**修正**：`sudo -S -k` + stdout/stderr 分流 + `e.redact()`；讀值改走 `os.ReadFile`（EPERM 才 fallback sudo）；`rm -f`；拿掉不可能成功的 unlink。**副作用（也是修好的一部分）**：log 不再出現明文密碼（真機驗證 0 處）、5 秒刷新不再為六個參數寫失敗行。
- **log 曾外洩 sudo 密碼（已處理）**：`cryoutils_ng.log` 的 mode 原是 **0666**（世界可讀）而且內容出現過明文密碼 5 處 → 真機已 scrub（替換成 `[REDACTED-PASSWORD]`）+ `chmod 600`，兩支 main 的 `os.OpenFile` 改 0600。**建議：更新 Deck 的 sudo 密碼**（它已在那台機器的可讀 log 裡存在了 10 天）。
- **web 模式仍無法寫入任何參數**（已知缺口，非 bug）：P0.5 移除了 web 密碼入口 → web 模式 `e.Password` 恒空 → 讀取现在正常（不再需要 root）、**寫入一律 `ErrNoSudoPassword`**。要改東西必須用 native UI（有密碼對話框）或 CLI 以 root 跑。
- **VRAM 讀值隨 session 變化**（**未修，新觀察**）：2026-10-30 從 ssh session 跑 `status` 得到 `VRAM: 100MB`，而 2026-10-20 真機 log 是 `Found a VRAM of 1024`。`GetVRAMValue()` 依賴 `glxinfo`，沒有 X session / 拿到 llvmpipe 時就會報錯值。需在桌面環境驗證並考慮加「讀不到時顯示 unknown」。
- **未解鎖時 unit 讀值顯示 `false`**（瑕疵已大幅縮小）：`getUnitStatus()` 在沒有密碼時回 `"nil"`，CLI/UI 把它呈現成 `false`（例如 `HugePages: false`），實際意義是「不知道」。P0.9 之後六個 unit 參數**不經 root 就能讀**，正常情況下不會再走到那條分支；只有真的讀不到（不存在的參數、不可讀的客製 kernel）才會出現。要改成 unknown 仍需同時動 CLI 輸出、web/native 的 `formatEnabled` 與 i18n 鍵。
- **web UI 的 swap 大小清單是寫死的**（**未修，不一致**）：`web/src/App.tsx` 用 `['2','4','6','8','12','16','20','24','32']`，沒走 core 的 `GetAvailableSwapSizes()`（會依 `/home` 剩餘空間過濾、並標示 Current Size）。native 是正確來源；兩邊清單因此可能不同。
- **Sudo cached timestamp bypass** (`core/sudo.go`): `sudo -S` 有 cached timestamp 時會忽略 stdin 密碼導致 `TestAuth` 永遠通過。修正：所有 `sudo` 指令加 `-k` flag 強制忘記 cache，`RenewAuth()` 回傳 `error` 並檢查空密碼。
- **CLI sudo 執行時 RenewAuth 失敗** (`core/sudo.go`): CLI 以 `sudo` 執行時 `e.Password == ""`，`RenewAuth()` 因檢查空密碼而失敗，但實際操作會成功（已是 root）。**已修正（P0.5）**：`RenewAuth()` 新增 `os.Geteuid() == 0` 檢查，已為 root 時直接回傳 `nil`。
- **Web UI 密碼層多餘** (`web/src/App.tsx`): Steam Deck 環境預設信任同機使用者，token 已提供 API 安全。**已修正（P0.5）**：移除密碼輸入 UI、`sudoLocked` 狀態與全部 `disabled={sudoLocked}` 條件，並同步移除 `web/src/api.ts` 的 `auth()`、`cmd/desktop/api.go` 的 `/api/auth` 端點、相關 CSS 與 locale 鍵。
- **Swap file location bug** (`core/swap.go`): `/proc/swaps` 遇到 `/dev/zram0` 時應 `continue` 跳過，而非 `return error`。修正後正確讀取 `/home/swapfile`。
- **Log directory auto-creation** (`cmd/cryoutilities/main.go` + `cmd/desktop/main.go`): 啟動時自動 `os.MkdirAll(core.InstallDirectory, 0755)` 建立目錄，避免 `sudo` 下 `os.UserHomeDir()` 回傳 `/root` 時找不到 log 檔。
- **Status command stdout** (`cmd/cryoutilities/main.go`): `printStatus()` 新增 `fmt.Printf` 輸出到 stdout，log 檔仍保留 `InfoLog` 輸出。
- **Web 模式 sudo 入口缺口**（已知缺口）：P0.5 移除密碼輸入 UI，`POST /api/auth` 端點也已刪除；web 模式 `e.Password` 恒空 → **非 root web 模式** `RenewAuth()` 失敗（需權限的操作會報錯）；web 以 root 執行時無影響（`Geteuid` 跳過）。Fyne 原生 UI（Phase 6.5）有自己的首啟/失敗密碼對話框，不受影響；是否補回 web 入口是獨立議題。
- **ZRAM 支援** (`core/swap.go`): SteamOS 3.6+ 使用 zram swap（`/dev/zram0`，由 zram-generator 管理：`/usr/lib/systemd/zram-generator.conf` → `zram-size = ram/2`、`zstd`、`swap-priority = 100`）與 swap file 並存。狀態讀取：`GetZramStatus()` 讀 `/proc/swaps`，`GetZramSizeBytes()` 讀 `/sys/block/zram0/disksize`，`getTotalSwapGB()` 從 `/proc/meminfo` 的 `SwapTotal` 計算總 swap（zram + swap file）。舊的「resize 後主動 `swapon /dev/zram0`」已被移除 —— 它根本不可能成功，見下一條 P0.7。
- **swap resize 摧毀 zram（P0.7，已修正並已真機驗證通過 2026-10-20）** (`core/swap.go` `DisableSwap()`): `sudo swapoff -a` 在 SteamOS 上不只是關掉 zram 的 swap，會觸發 systemd 把 zram 整套拆掉 —— `dev-zram0.swap` Deactivated → `systemd-zram-setup@zram0.service` 停止 → `ExecStop=zram-generator --reset-device zram0` 把 `disksize` 寫 0（kernel：`zram0: detected capacity change from 15159296 to 0`）。之後的 `swapon /dev/zram0` 是對容量 0 的裝置操作（EINVAL）必然失敗，失敗又被 log 吞掉；第二次 resize 起 `GetZramStatus()` 已是 false → 連嘗試都不嘗試（bug 自我固化）。teardown 是**非同步**（swapoff 後約 21 秒），所以「事後補 swapon」從設計上就會 race。**不會**自己回來（舊 P0 「等 systemd 掃描約 18 分鐘」假設錯誤）。**已採方法 A**：`swapoff <location>` 只關自己的 swap file，完全不碰 zram；zram 若仍不見只 log 警告。另有換頁安全檢查（`Used` vs `MemAvailable` + 其他 swap 剩餘容量 + 512 MB margin）。完整根因實證與驗收條件見 **`docs/zram-swapoff-reset.md`**；工作項見 `todo.md` P0.7。**SteamOS 端復原方法**（deck 在 `wheel` 群組，polkit 允許，不需密碼）：`systemctl restart systemd-zram-setup@zram0.service`。
- **resize 路徑曾有兩份實作**（已修正 2026-10-20）：`cmd/desktop/api.go` 的 `handleSwapResize` 自己重寫了 `DisableSwap → ResizeSwapFile → SetSwapPermissions → InitNewSwapFile` 四步，**繞過 `ChangeSwapSize()`** —— 這是 core 層的 zram/安全邏輯對 web UI 永遠無效的原因。現在三條路徑（CLI / web / native）都只呼叫 `ChangeSwapSize()`；進度訊息由 core 的 `emitProgress` → `OnProgress` 送出（desktop 端轉到 `core.ProgressChannel`，**非阻塞** send，避免 SSE 無人消費時在 swap 已關的狀態下卡死）。教訓：**任何跨 UI 的操作序列只允許一份實作**。
- **CLI 未解析 swap file 位置**（已修正 2026-10-20）：只有 `GetSwapFileSize()` 會設 `e.SwapFileLocation`；CLI `swap <n>` 從不查狀態 → `dd of=` 是空字串。新增 `resolveSwapFileLocation()`，每個 swap 步驟各自解析（含 fallback 到 `DefaultSwapFileLocation`）。
- **log 自我抹除**（已修正 2026-10-20）：CLI 與 desktop 都在啟動時 `os.Remove(core.LogFilePath)` 且共用同一個 log 檔 → 前一次失敗的 `Failed to re-enable zram` 被下一次啟動覆蓋，查證只能去翻 journal。現改為 `O_APPEND`；同時 CLI 失敗也會把錯誤印到 **stderr**（安全檢查的拒絕訊息不能只存在 log 裡）。

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
- **P0: ZRAM 支援**（最高優先級）— zram 狀態顯示與 CLI 輸出 ✅；舊的 resize 後 zram 恢復設計錯誤，由 **P0.7** 取代並已修復 ✅
- **P0.7**: swap resize 摧毀 zram — 根因已定位（`swapoff -a` 觸發 zram-generator `--reset-device`）、**方法 A 已實作並已 Deck 真機驗收通過（2026-10-20：四次 resize 2→8→16→12 GB，zram0 Priority 100 全程不動）**；單測 12 個全過（見 `docs/zram-swapoff-reset.md`）
- **P0.8**: 真機 feedback 第二批 — 雙重 sudo 密碼提示（裸 `sudo` 會碰 tty）與 native UI swap 下拉 freeze — 已修正並真機結案（2026-10-20）
- **P0.9**: unit 參數（hugepages / shared memory in THP / compaction proactiveness / hugepage defrag / page lock unfairness）不可偵測與不可用 — 根因：`runSudoWithInput` 用 `sudo -S` 但**缺 `-k`**，sudo 在 timestamp 快取態下不讀 stdin，**密碼那一行被 `tee` 當值寫進 sysfs**（kernel EINVAL → `tee` exit 1 → 每次 toggle 都報錯）；另有狀態讀取不必要地走 root（web 模式無密碼入口 → 五個值永遠 false）。**已修正並已 Deck 真機結案（2026-10-30）**；順帶修掉**明文 sudo 密碼外洩到 log**（log mode 0666 → 0600 + `e.redact()`）。實證與規格見 **`docs/unit-params-sudo-stdin.md`**，工作項見 `todo.md` P0.9
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
