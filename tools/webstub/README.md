# tools/webstub — web UI 的假 API server（僅供 Dev VM 開發用）

## 這是什麼

一支用 Python 標準庫寫的假 server。它 serve `web/dist`（已建置好的 web UI），
然後用固定的 JSON 回答 `/api/*`。它取代了過去放在 `/tmp/stub.py` 的一次性腳本；
放在 repo 內才不會被 `/tmp` 清理弄丟。

真正的 `cryoutils-ng-desktop -ui web` 在 Dev VM 上看不出東西：Dev VM 是 LXC，
沒有 swap file、沒有 zram、沒有 `glxinfo`，而且 web 模式沒有 sudo 密碼入口。
跑真的二進位時，status 區每個欄位都是空或 `unknown`，正在檢視的部分剛好都被藏起來。
這個 stub 用 fixture 假裝成「只在真 Deck 上才會出現的狀態」。

## 怎麼跑

```bash
python3 tools/webstub/stub.py --no-sse           # 預設 fixture：deck-recommended
python3 tools/webstub/stub.py --list             # 列出所有 fixture 與說明
python3 tools/webstub/stub.py --fixture deck-factory
python3 tools/webstub/stub.py --verbose          # 印出每個請求
```

開瀏覽器連 `http://127.0.0.1:8099/`。注意它 serve 的是 `web/dist`（已建置的產物，
而且 repo 內追蹤的 `web/dist` 可能比 `web/src` 舊）。改過 `web/src` 之後要先
`cd web && npm run build`，否則看到的還是舊畫面。

`web/vite.config.ts` 沒有設 API proxy，所以 `npm run dev` 的 Vite port 上沒有 `/api/*`。
要在 Vite 的熱重載下看資料，就自己加 proxy 指到 8099，或直接走上面那條「先 build」的路。

## 配對 Playwright（雙解析度截圖，見 AGENTS.md 的驗收規則）

```bash
python3 tools/webstub/stub.py --no-sse --port 8099      # 一個終端
cd web && BASE_URL=http://127.0.0.1:8099 npx playwright test   # 另一個終端
```

**跑 spec 時 stub 一定要加 `--no-sse`。** `/api/progress` 平常保持連線（跟真 server
一樣），而截圖測試用 `waitForLoadState('networkidle')`，network idle 因此永不達到
（實測那兩個測試逾時失敗）。加 `--no-sse` 後 `/api/progress` 回 501（EventSource
對非 200 不重試），14 個測試全過，含 4K 與 1280×800 兩張截圖。

`web/playwright.config.ts` 的 `baseURL` 讀 `BASE_URL`，預設 `http://127.0.0.1:33031`
（那是真的 desktop server 的隨機 port）。CI 沒有跑 Playwright；config 為 CI 準備的
`npm run preview` 也不提供 `/api/status`，所以那條路徑即使跑起來，截圖的 status 區
還是空的（待辦見 `todo.md`）。

本機若 `npx playwright test` 抱怨 browser build 不存在，跑 `npx playwright install chromium`。
需要的 build 版號由 `node_modules/playwright-core/browsers.json` 決定，與 cache 裡
現有的 build 不一定對得上。

## 用 chrome 直接截圖（不跑 spec）

等網路 idle 的截圖模式一樣會被 open SSE 卡住，所以也要 `--no-sse`：

```bash
python3 tools/webstub/stub.py --no-sse --port 8099
```

另一終端：

```bash
"$CHROME_HEADLESS_SHELL" --headless --no-sandbox --window-size=1280,800 \
  --hide-scrollbars --virtual-time-budget=8000 --screenshot=out.png http://127.0.0.1:8099/
```

不加 `--virtual-time-budget` 時截圖會在 fetch 完成前就按下，status 區會是初始值。

## fixture

| 名稱 | 內容 | 那些值是實測的 |
|---|---|---|
| `deck-recommended` | 套用 Recommended 之後的 Deck | 五個 unit 參數的值 = `core/config.go` 的 recommended |
| `deck-factory` | 出廠的 Deck：只有 zram、沒有 swap file | 五個 unit 參數的值 = 真機開機實測（`docs/unit-params-sudo-stdin.md` §8.3：`always / advise / 20 / 5 / 1`） |
| `devvm-lxc` | Dev VM LXC 的真實 `/api/status` 擷取（2026-10-30） | 全部逐字保留，含 `SwapFile` 空但 `SwapSizeGB 1` 這種舊路徑的怪值 |
| `parameters-unreadable` | 五個參數都讀不到（Value = `unknown`） | 這是刻意造的，用來驗 `status.unknown` 的顯示路徑 |

每個 fixture 的 `_note` 欄位寫了它是實測還是示意值，`_` 開頭的欄位不會送到 client。
swap 大小、zram 大小、`Swappiness`、`VRAM` 若未標明實測，就是示意值，
**不可**當作 SteamOS 的真實數字引用。

`web/tests/screenshots/` 那兩張入版控的截圖是從 `deck-recommended` fixture 產生的：
它們用來看排版，不是真機的狀態參考。

## 安全邊界（必讀）

- **只 bind `127.0.0.1`**。真的 desktop server 也只 bind loopback，而且加每次啟動的隨機
  token（AGENTS.md）。stub 沒有 auth：綁到其他介面會讓 LAN 上任何人讀到整個 web build。
  `--addr` 給非 loopback 時會直接退出，除非明確設 `WEBSTUB_ALLOW_LAN=1`。
- **它不做任何寫入**。`POST /api/*` 一律回 `{"success": true}` 但什麼都不改，
  所以按 UI 上的按鈕只會看到「成功」然後畫面刷新回同一組 fixture。
  要驗證真的寫入，用真的二進位（native UI 或 CLI 以 root 跑），或在 Deck 真機驗。
- `/api/progress` 只送 SSE comment 保持連線，**不會**送 progress 事件。
  進度條的畫面要另外用真的 server 看。
- 預設不檢查 token（UI 仍會附 `?token=`，被忽略）。要測真的 401 路徑就加 `--token foo`。

## 它不覆蓋什麼

`cmd/desktop/api.go` 的真實行為（token 檢查、SSE 事件內容、錯誤處理）在這裡沒有等價物。
任何關於 core 或 API 正確性的結論，都不能從這個 stub 得到。
