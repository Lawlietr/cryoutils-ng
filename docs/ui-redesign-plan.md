# UI-1 — UI/UX 重新設計企劃（量測診斷 + 設計系統規格）

> 日期：2026-11-11（Dev VM `a1` / `Code-Agent-01`，Ubuntu 24.04.5）
> 性質：**設計文件，不是施工紀錄**。本文件描述的改動**一項都尚未實作**。
> 受影響檔案：`web/src/App.css`、`web/src/App.tsx`、`ui/fyneui/{app.go,layout.go,sections_*.go,fonts.go}`。
> 不影響：`core/`、REST/SSE API、i18n locale 鍵的語意、CLI。
> 觸發來源：用戶回饋「UI 比原始項目彆腳、設置擠在一起、大小排序不整齊、不美觀」。

### 決策清單（勿再評估）

| # | 決策 | 來源 |
|---|---|---|
| 1 | 資訊架構 = **單頁 + 左側 nav rail**（tab 與摺疊都否決） | [DECIDED] 2026-11-11 用戶 |
| 2 | **Presets 上移**：新順序 Status → Presets → Swap → Memory → Game Data | [DECIDED] 同上 |
| 3 | **VRAM 合併進 Status 卡 A**，不再獨立成卡 | [DECIDED] 同上 |
| 4 | **施作範圍 = 只有 native**（`ui/fyneui/`）；web 那一輪尚未排程 | [DECIDED] 同上 |
| 5 | 不換 GUI 框架，沿用 Fyne v2.7.4 | [DECIDED] |
| 6 | 垂直節奏不動（實測本來一致） | [DECIDED] |

施作順序與 gate 見 `todo.md` UI-1。

---

## 1. 結論（先讀這段）

1. 回饋是可驗證的，但**根因不是缺框架，是缺兩樣東西**：（a）沒有設計系統（無間距階梯、無字級階梯）；（b）表單沒有欄位對齊（沒有 label 欄、沒有 value 欄）。
2. **web 與 native 的 4K 處境完全不同**：native 的 4K 骨架（`FYNE_SCALE` re-exec + 內容寬度上限 + 置中）**實測是對的**；web 端**完全沒有 4K 適配**，同一條 `max-width: 720px` 用在 1280 與 3840，結果 4K 時內容只佔螢幕寬 18%。
3. 「擠在一起」與「中間空一大片」是同一件事的兩面：**水平對齊沒有欄位可依循**，所以元素各自被 `space-between`／`flex` 推到不同位置，視覺上既空又亂。
4. 垂直節奏沒問題（實測 web 卡間距 20~21px 一致、native 列距 40~41px 一致）。**不要改垂直節奏**。
5. 本企劃不換 GUI 框架。Fyne 的具體瑕疵（§2.7）都能在 Fyne v2.7.4 內修掉（修法見 §5.6、§5.7）。

---

## 2. 量測資料（全部實測，附 provenance）

### 2.1 資料來源

| 來源 | 產生方式 | 代表現況？ |
|---|---|---|
| `web/tests/screenshots/steam-deck-1280x800.png`（1280×1618 fullPage） | `tools/webstub` fixture + Playwright，2026-11-11 產生（P0.11） | 是：產生日比 `App.css`（2026-10-30）新 |
| `web/tests/screenshots/4k-3840x2160.png` | 同上 | 是 |
| native Xvfb 截圖 1280×800、3840×2160 | 本輪重新編譯 `cmd/desktop`（`CGO_ENABLED=1`）後 `-ui native` 截圖 | 是：新舊 PNG **逐位元組相同**（81,125 bytes） |

fixture 內只有標明實測的值可用（見 `tools/webstub/README.md`）。本文件只引用**版面幾何**，不引用 fixture 的數值。

### 2.2 web：容器寬度

| 視窗 | 內容欄實寬 | 佔螢幕 | 兩側空白合計 |
|---|---|---|---|
| 1280 | **688 px** | 54% | 592 px |
| 3840 | **688 px** | **18%** | **3,156 px** |

內容寬在兩個視窗**完全相同** → 4K 是一張名片貼在黑畫布中間。來源：`App.css:34` `#root { max-width: 720px }`，無 breakpoint、無 scale 層。

另外：1280 寬度下頁面總高 **1,618 px**，Deck 螢幕高 800 → 要滾 2 屏才能看完六個區塊。

### 2.3 web：Memory 五行（「不整齊」的直接證據）

`.memory-item { justify-content: space-between }`（`App.css:189`）。實測方式：文字 block 聚合，缺口 > 25 px 斷開：

| 行 | label（起點 → 結尾） | 右側群組（值 + 開關） | 中間空白 |
|---|---|---|---|
| Huge Pages | 316 → 401 | 864 → 963 | 463 px |
| ShMem | 315 → 370 | 864 → 963 | 494 px |
| Compaction Proactiveness | 315 → 508 | 903 → 963 | 395 px |
| Defrag | 316 → 364 | 903 → 963 | 539 px |
| Page Lock Unfairness | 316 → 470 | 903 → 963 | 433 px |

- **label 起點是對齊的**（全部落在 315/316）→ 這一項不用改。
- **右側群組全部貼齊 x=963**（卡內右內邊界），但起點只有 864 與 903 兩種 → 值字串的左邊界隨字串長度飄移。
- **中間空白 395~539 px**，五行各不同 → 這是「中間空一大片、右邊擠」的直接成因。
- 原因：沒有 value 欄。值字串長度不同（`always`/`madvise`/`advise`/`20`/`5`/`1`）+ `space-between` 把開關推到最右 → 值就貼著開關隨字串飄移

### 2.4 web：Swap 兩行

`.swap-row` 是 flex，`label { min-width: 100px }`，但 select／input／button 都沒有寬度規則（`App.css:156-181`）。兩行的元素以填色定位（`--bg-input` 找控制件、`--accent-dim` 找按鈕）：

| 行 | 控制件 | 動作按鈕 |
|---|---|---|
| Swap File Size + Resize Swap | x 465..574（寬 109 px） | x 592..702（寬 110 px） |
| Swappiness + Apply | x 432..509（寬 77 px） | x 526..591（寬 65 px） |

→ 控制件起點差 **33 px**、按鈕起點差 **66 px**、寬度也各不相同。兩行的 label 起點都對齊在 315/316，
但其右側沒有任何欄位規則，所以完全錯位。

### 2.5 web：Status 卡

`.status-grid { grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)) }`（`App.css:129`）。688px 內容扣掉左右各 20px padding = 648px → **3 欄，每欄 ~204px**。
實測三欄的 label 起點在 x 316 / 535 / 755（欄距 219~220 px，扣掉 12 px gap 後每欄內容寬約 207 px）。
同時，label 是 0.75rem uppercase + `letter-spacing: 0.05em`（`App.css:139-142`），而字串是 20+ 字元（`COMPACTION PROACTIVENESS`、`SHARED MEMORY IN THP`）→ label 本身就把欄位撐爆。這張卡塞了 **12 筆**資料，而 VRAM 卡只放一句話且**不顯示數值** → 資訊密度兩極。可用寬度只有 648 px，label 長度卻不受欄寬約束。

### 2.6 web：設計系統不存在

`:root` 只有顏色與 `--radius`（`App.css:4-19`）。實測：

- **15 種間距值**（只算 `padding` / `margin` / `gap` 宣告）：0.2, 0.25, 0.3, 0.35, 0.4, 0.5, 0.6, 0.75, 1, 1.25, 1.5, 3, 4 rem，加 8px 與 10px 兩個絕對值
- **11 種 `font-size` 值，分屬三種互不相容的公式**：8 種絕對值（0.75/0.8/0.85/0.9/0.95/1 rem，0.82em，0.85em）+ 3 種 `clamp()`：`clamp(14px,1.2vw,18px)`（html）、`clamp(1.4rem,3vw,2rem)`（h1）、`clamp(1rem,2vw,1.25rem)`（h2）
- 對齊 hack：`.header .version { margin-left: -0.5rem; align-self: flex-end }`（`App.css:86-90`）→ 版本文字看起來漂浮
- 可供性：`.btn-toggle` 與 `.btn` 形狀相同（同 radius、同 border 規則），差異只有填色 → 「開關」看起來像「按鈕」

### 2.7 native（Fyne v2.7.4）

**對的部分（保留，不要推翻）**：

- `contentMaxWidth = 1100`（`ui/fyneui/layout.go:27`）在 1280 生效且置中：swap 大小 Select 實測 x **97..1182**，左右內縮各約 97 px。
- 4K 的 re-exec 生效：run log `scale: display width 3840 px → re-exec with FYNE_SCALE=2`；視窗佔螢幕 80%（x 384..3455），內容 cap 1100 logical 置中（內容左邊界在 physical 822）。

**錯的部分**：

1. **Status 兩欄清單的中間空白 478 px**：左欄文字 x 105..173，右欄文字 x 651..798。
2. **左右兩欄語意交錯**，讀起來對不上。左欄是 Swap 檔案、ZRAM 大小、Swap 總量、VRAM、共享記憶體 never、碎片整理 1、頁面鎖定不公平性 5；右欄是 Swap 狀態、ZRAM 狀態、Swappiness、大記憶體頁面 madvise、壓縮積極性 20、頁面鎖定不公平性 1。同一欄混了「名稱+值」與不同主題。
3. **控制件尺寸懸殊**：Select 寬 **1,085 px**、「調整 Swap」按鈕滿版；同一卡內 Swappiness 的輸入框與「設定」按鈕只有幾十 px 寬。
4. **Card 沒有可見邊界**（暗色主題下實測截圖看不出邊框）→ 六個區塊視覺上連成一條長欄，這是「擠在一起」感受的主要來源。
5. **Checkbox 未勾選時對比過低**：方塊幾乎看不見；文字「頁面鎖定不公平性 (5)」在 fold 附近時只剩文字可讀。
6. **1280×800 一屏內只露出 3/6 個區塊**（系統狀態、Swap 設定、記憶體設定開頭）。

### 2.8 已排除的懷疑（不要再查）

- ❌「native 三個區塊標題字級不一致」：**錯**。實測 CJK 字面 advance：系統狀態 23.5 px、Swap 設定 23 px、記憶體設定 23.6 px —— 相同。看起來不同是因為拉丁字 `S`/`p` 的上下延伸與詞寬較長。
- ❌「web 卡間距不一致」：**錯**。實測卡間距 20~21 px（六張卡全部一致）。
- ❌「native 的 1100px 上限沒生效」：**錯**。1280 與 3840 都生效且置中。

---

## 3. 診斷對照表

| 症狀（用戶用語） | 實測 | 位置 |
|---|---|---|
| 大小排序不整齊 | 15 種間距值、11 種 font-size | `App.css` 全文，無 token |
| 中間空一大片、右邊擠 | Memory 行空白 395~539 px、右側群組起點只有 864/903 | `App.css:189` `space-between` |
| 按鈕沒對齊 | 按鈕起點 592 vs 526、寬 110 vs 65 | `App.css:156-181` 無控制件欄 |
| 擠在一起 | Status 12 筆塞 3 欄 ~204px、uppercase 長 label | `App.css:129,139-142` |
| 不美觀（4K） | 內容 688px = 18% 視窗寬 | `App.css:34` |
| 六區塊像一條長欄 | Card 無可見邊界 | `ui/fyneui/` theme 未覆蓋 Card 邊框 |
| 表單控件拉太長 | native Select 1,085 px | Select 無寬度上限 |
| 找不到常用的東西 | Presets（最高頻動作）排在第 4 區塊 | `App.tsx` 區塊順序 |

---

## 4. 設計系統規格（web 與 native 共用同一套 token 值）

數值以**邏輯 px** 表示；native 端由 Fyne theme 與 `theme.Size()` 對映。

### 4.1 間距階梯（8px 網格）

| Token | px | 用途 |
|---|---|---|
| `--s-1` | 4 | label 與其值的間隔 |
| `--s-2` | 8 | 控件內 padding（垂直）、chip 內距 |
| `--s-3` | 12 | 列內相鄰元素 gap |
| `--s-4` | 16 | 卡內相鄰列 gap |
| `--s-5` | 24 | 卡內 padding |
| `--s-6` | 32 | 卡與卡之間 |
| `--s-7` | 48 | 區塊（section）之間、頁面上下留白 |

規則：CSS 只準出現這 7 個值。現有 15 種值全部對應到最接近的一級後移除。

### 4.2 字級階梯（比例 1.25）

| Token | px | 用途 |
|---|---|---|
| `--fs-micro` | 12 | status label（大寫、`letter-spacing 0.04em`） |
| `--fs-sm` | 14 | 次要文字、hint、版本號 |
| `--fs-body` | 16 | 一般 label、輸入框、按钮文字 |
| `--fs-md` | 20 | 數值（status value） |
| `--fs-lg` | 25 | 區塊標題（h2） |
| `--fs-xl` | 31 | 頁面標題（h1） |

規則：`clamp()` 只準出現在 `html` 的 base font 一處，其餘全部用 rem 相對值。base font 由 `--ui-scale` 控制（見 §5.1），這樣 4K 只要調一個變數。現有 11 種 font-size 全部對應到最接近的一級後移除。

### 4.3 控件尺寸（兩條路徑都要守）

| 項目 | 規則 | 現況違反例 |
|---|---|---|
| 控件高度 | desktop 36 px；Deck 觸控 44 px | native 的 Swappiness 輸入框與「設定」按鈕明顯低於此 |
| 控件寬度上限 | **select／input 一律 ≤ 320 px**，不準滿版 | native Select 1,085 px |
| 按鈕最小寬 | 96 px | native「設定」~30 px |
| 同卡內動作按鈕 | 全部對齊到同一條垂直線（走 §5.2 grid 第 3 欄） | 按鈕起點 592 vs 526 |
| 開關 | 真 switch：track 40×22 px + knob，狀態由 knob 位置 + 顏色 + 文字三重表達 | `.btn-toggle` 看起來像按鈕 |

### 4.4 顏色語意（沿用現有 palette，只補規則）

- `--accent` 只用於：區塊標題、focus 邊框、主要按鈕。
- 狀態用三色固定語意：符合 recommended = `--green`；非 recommended 但合法 = `--orange`；讀不到 = `--text-muted` + 文字 `unknown`（沿用 P0.10 的 `status.unknown` 鍵）。
- 卡背景要能看出邊界。web：卡背景 `#16213e` 對頁面背景 `#1a1a2e` 的亮度差過小，邊框 `--border` 對比也要拉高。
  native：實測視窗背景 `#171718`，Card 的預設階差在其上看不出邊界 → 要顯式覆蓋 Card 的邊框色（見 §5.6）。

---

## 5. 版面規格

**本輪只施作 native**（決策 4）。§5.1 與 §5.3 的 CSS 片段是 web 那一輪的規格，**現在不動**；§4 的 token
值兩邊共用，native 端透過自訂 `fyne.Theme` 對映（§5.7）。本輪要做的是 §5.2、§5.4、§5.6、§5.7。

### 5.1 容器策略（web，本輪不做）

```css
html { font-size: calc(16px * var(--ui-scale)); }        /* 唯一的縮放控制點 */
:root { --ui-scale: 1; }
@media (min-width: 1920px) { :root { --ui-scale: 1.35; } }
@media (min-width: 2560px) { :root { --ui-scale: 1.6; } }

#root { width: min(1120px, 92vw); margin-inline: auto; }
```

- Deck 1280×800：`min(1120px, 92vw)` 給 1,120 px（現況 688 px）。六張卡變寬，但控制件仍受 §4.3 的 320 px 上限約束。
- 3840×2160（無 OS 縮放的最壞情況，就是 Playwright harness 的環境）：內容 1,120 px、base font 25.6 px。
  兩側空白各 1,360 px，是內容寬的 1.2 倍——這是**刻意保留的空白**，不是把 select 拉長 3 倍。
- 真機的 4K 桌面通常有 OS 縮放（例如 200%），瀏覽器看到的 viewport 約 1,920 CSS px → 內容佔比會自然拉高。
  所以截圖驗證要用無縮放的 3840（最壞情況），真機則用實際桌面環境看（§7）。
- **控制件不隨容器變寬**：容器變寬只帶來「更多留白 + 更大字」。這是與現況最大的差別。

### 5.2 表單網格（web + native 同一個邏輯）

所有「label + 控制件 + 動作」的列一律走這個 grid：

```
label 欄              控制件欄                 動作欄
minmax(160px,1fr)     minmax(140px,320px)      auto
        └ gap: var(--s-4) ┘        └ gap: var(--s-3) ┘
```

- 覆蓋範圍：Swap（大小 + Resize）、Swappiness（+ Apply）、Game Data（兩個 select 各佔 1 列，Sync／Cleanup 對齊動作欄）。
- native 對映：`widget.NewForm()`（Fyne 的 Form 免費給你 label 欄對齊），**不要**用 `container.NewBorder` + 滿版 Select。Select 要 `Resize(fyne.NewSize(320, h))` 並放在 Form 的 field 欄內。

### 5.3 Memory 五行：名稱 / 值 / 開關 三欄

native 現況**根本沒有值欄**：值是塞在 label 的括號裡（實測截圖：「大記憶體頁面 (madvise)」、
「共享記憶體 (never)」）。本輪要拆成三欄：名稱（1fr）／值（固定寬、右對齊）／開關。

下面是 web 那一輪的 CSS（本輪不動，留給後續）：

```css
.memory-item { display: grid;
  grid-template-columns: 1fr var(--value-col, 110px) 40px;
  gap: var(--s-4); align-items: center; }
.memory-value { text-align: right; font-variant-numeric: tabular-nums; }
```

- 五行的值起點必須相同（`--value-col` 固定寬），開關全部對齊同一條線。
- 數值用 `tabular-nums`，避免 `20 / 5 / 1` 寬度不同造成的抖動。
- native：三欄用 `layout.NewFormLayout()` 或 `container.NewGridWithColumns(3)`，但**值欄要固定寬**（`Resize` 寫死），不能用會隨 MinSize 飄移的預設。

### 5.4 Status 拆成兩張卡 + 定義列表

拆法照 `core.GetUnitStates()` 的邊界，不重新發明分類：

- **卡 A「Swap 與記憶體」**：Swap 檔案、Swap 大小、ZRAM 大小、ZRAM 狀態、Swap 總量、Swappiness、VRAM（含「讀不到時顯示 `unknown`」的現況處理）。
- **卡 B「Kernel 參數」**：五個 unit 參數，每列 `label(micro) / 值(fs-md) / 是否 recommended(chip)`。

兩張卡內部都用**兩欄定義列表**（label 欄固定寬，不 uppercase 到撐爆欄位），**不用** `auto-fit minmax(200px,1fr)` 的迷你卡。uppercase 只準用在 ≤12 字元的 label；更長的（`Compaction Proactiveness`）用一般大小寫。

native 對應：現在的 `container.NewGridWithColumns(2)` 讓語意交錯（§2.7 第 2 點）。改成一張卡一條直式定義列表，或兩張卡各一條，**不準**讓左欄/右欄混不同主題。

### 5.5 VRAM 卡：[DECIDED] 合併進卡 A

現況是一張只放說明的空卡。已定：改成卡 A 的一行 `VRAM: <值 or unknown>（BIOS 控制）`，
說明文字變 hint。理由：`GetVRAMValue()` 依賴 `glxinfo`，讀不到時（ssh session 下實測過報錯值）
更要省空間。對應檔案：`ui/fyneui/sections_vram.go` 的資料餵進卡 A，卡本身消失（本輪只處理 native）。

### 5.6 native：Card 邊界與 checkbox

**Card 沒有邊框線是框架行為**，不是配色問題：Fyne v2.7.4 的 `Card` renderer 用
`widget.NewShadowingRenderer(objects, widget.CardLevel)`，也就是 Card 只有**陰影**沒有邊框；
theme 也沒有 Card 適用的邊框色名（`ColorNameBorder` 不存在，只有 `theme.ColorNameSeparator`）。
實測在 `#171718` 背景上看不出邊界，所以六個區塊連成一條長欄。

三條修法（選 a 或 b）：

- **(a)** 區塊外層蓋一層 `canvas.Rectangle`（`StrokeColor = th.Color(theme.ColorNameSeparator, v)`、
  `StrokeWidth = 1`、`CornerRadius` 對齊 §4 的 radius），內容放其上。
- **(b)** 不用 Card，改用自訂的 card-like 容器（同一個 Rectangle 描邊 + 背景色），好處是 padding 可控。
- **(c)** 底線：區塊之間至少加一條 1px 分隔線（`theme.ColorNameSeparator`）。

**Checkbox**：未勾選方塊要能看到邊框；列高對齊 §4.3 的 44 px。要「整列可點」得自己做（外層 Row 的
Tappable 轉呼 `Check.Toggle()`），但**不得**因此繞過 `syncing` 守衛（P0.10 規則，見 `todo.md` UI-1 紅線）。

### 5.7 native：token 對映與 nav rail

**token 對映到 Fyne theme**。v2.7.4 的 `fyne.Theme` interface 只有四個方法（`theme.go:28`）：
`Color(ThemeColorName, ThemeVariant)`、`Font(TextStyle)`、`Icon(ThemeIconName)`、`Size(ThemeSizeName)`。
沒有 Style 層，也沒有 `NewTheme` 建構子（只有 `theme.DefaultTheme()`）→ 寫一層 wrapper 包住
`theme.DefaultTheme()`，把 §4 的 token 映到這些 size name：

| §4 token | Fyne size name |
|---|---|
| `--s-1` … `--s-5` | `theme.SizeNamePadding`、`theme.SizeNameInnerPadding`（雨層就夠，其餘由呼叫端加） |
| `--s-6`（區塊之間） | 自己加在內容容器的 spacer／`MinSize` 上（Fyne 沒有 block spacing 這個 size name） |
| `--fs-body` | `theme.SizeNameText` |
| `--fs-micro` | `theme.SizeNameCaptionText`（名稱是 `helperText`） |
| `--fs-lg` | `theme.SizeNameSubHeadingText` |
| `--fs-xl` | `theme.SizeNameHeadingText` |
| 分隔線 1 px | `theme.SizeNameSeparatorThickness` |

注意 `SizeNameText` 改大會帶動所有 widget 的高度（Select、Button、Check 跟著變），這是 intended；
但 `ui/fyneui/fonts.go` 的 CJK 字體覆蓋要過同一個 token，不能一處改字級一處不改。

**nav rail**：

- 根佈局改成 `container.NewBorder(top, nil, navRail, nil, vScroll)`（nav rail 放第三參數）。
  現況是 `NewBorder(top, nil, nil, nil, vScroll)`（`ui/fyneui/app.go`）。
- nav rail 內容 = 5 個 `widget.Button`（或 vertical 的 `widget.RadioGroup`），寬 200 px。
- 捲動用 `vScroll.ScrollToOffset(fyne.NewPos(0, y))`。v2.7.4 只有 `ScrollToOffset`、`ScrollToTop`、
  `ScrollToBottom`，**沒有** `ScrollTo(float32)`。
- **y 要從實際佈局結果讀**：點下去時用 `sectionObj.Position().Y`（相對內容容器）。
  用 `MinSize().Height` 累加會錯：換行與字級讓實際高度與 MinSize 不同。
- 寬度門檻用現成的 `effectiveScale()`（`ui/fyneui/layout.go`）換算 logical 寬；<1200 時不建 nav rail。
- nav rail **不得**註冊進 `addRefresher`，也不得被 5 秒刷新重建（重建會丢鍵盤焦點與滾動位置）。
  只有語言切換才全量重建（現況規則）。
- 標籤走 i18n：新增 5 個鍵，`i18n/locales/en.json` 與 `zh-TW.json` 同步；drop-in 覆蓋机制不變。

---

## 6. 資訊架構

現況六區塊固定直式順序：Status → Swap → Memory → VRAM → Presets → Game Data。問題：**Presets（Recommended／Stock 是最高頻動作）排第 4**，1280×800 下 fold 內看不到；VRAM 佔一張空卡。

### [DECIDED] 2026-11-11（用戶）

- **單頁 + 左側 nav rail**：≥1200 logical px 時左側 200 px 錨點列；Deck 1280×800 與更窄時隱藏。
- **新區塊順序**：Status（卡 A + 卡 B，VRAM 在 A 內）→ Presets → Swap → Memory → Game Data。
  Presets 一次寫五個 kernel 參數，放在 Status 正下方可以讓使用者立刻在 Status 看到效果。
- nav rail 的項目是 **5 個**（不是六個）：系統狀態、Presets、Swap、記憶體、遊戲資料。

### [REJECTED] 響應式 tab

寬螢幕用 tab、窄螢幕單頁。**否決理由**：抵觸 AGENTS.md「Single page, vertical single column — no tabs」
與 `docs/fyne-ui-plan.md` §0 的 `[DECIDED]`；而且兩套導航要各自維護、各自驗收。

### [REJECTED] 摺疊區塊

**否決理由**：多一層互動，而且 task 進行中（swap resize 要幾十秒）時「收起的區塊有進度」要額外處理。

原項目用 5 個 tab、每 tab 整整一屏，這是它「看起來比較不擠」的主因。nav rail 拿到同一個好處
（一次專注一個區塊），但不增加第二套內容組織。

---

## 7. 驗收條件（Definition of Done）

### native（本輪）

- [ ] 單一控制件寬 ≤ 400 px（現況 native Select 實測 1,085 px）
- [ ] 同一卡內所有 label 起點 x 相同（±2 px）；Swap 的「大小 + 調整按鈕」與「Swappiness + 設定按鈕」
      兩行要落到同一組欄位（native 現況：兩行的控件位置由 Flex 決定，沒有共同欄位）
- [ ] Memory 五行的**值欄起點** x 相同（±2 px），且值不藏在 label 括號裡
- [ ] Status 卡內中間空白 ≤ 240 px（現況 native 兩欄之間 478 px）
- [ ] 五個區塊的邊界可見（Card 描邊或分隔線，§5.6）
- [ ] 1280×800 一屏至少露出 4 個區塊（現況 3 個）；Presets 要在 fold 內
- [ ] VRAM 不再是獨立卡（§5.5）
- [ ] nav rail：≥1200 logical px 時可見、5 個項目；點任一項目後 `vScroll.Offset().Y` 達到該區塊的
      `Position().Y`（±4 px）；<1200 時不存在
- [ ] nav rail 在 5 秒刷新後不被重建（用 fake sudo / 計數器斷言無特權命令，且 nav rail 的 widget 參照不變）
- [ ] `ui/fyneui/` 內不得出現 ad-hoc 尺寸常數（間距與字級一律走 §5.7 的 theme wrapper）

### web（本輪不做，留待後續那一輪）

- [ ] @3840（無 OS 縮放）：內容寬 ≥ 1,000 px（現況 688 px）且 base font ≥ 20 CSS px（現況 18 px）
- [ ] @3840：兩側空白各 ≤ 內容寬 × 1.35（現況各約 1,578 px，是內容寬的 2.3 倍）
- [ ] @1280：內容寬 / 視窗寬 ≥ 85%（現況 54%）
- [ ] `App.css` 的間距值種類 ≤ 7 種（現況 15 種）、`font-size:` 只準引用 token（現況 11 種）
- [ ] 行內中間空白 ≤ 240 px（現況 395~539 px）

截圖矩陣（前後都要存，檔名加 `-before` / `-after`）。本輪只做 native 那兩列；web 兩列留給 web 那一輪：

| | 1280×800 | 2560×1440 | 3840×2160 |
|---|---|---|---|
| native（Xvfb，zh-TW／en） | ✅ 本輪 | ✅ 本輪 | ✅ 本輪 |
| web（Playwright） | 本輪跳過 | 本輪跳過 | 本輪跳過 |

真機（用戶在場，Deck sshd 開著時可直接驗）：
- [ ] 1280×800 與 4K 兩條都實際看一眼，文字可讀、控制件可點
- [ ] nav rail：4K 上出現且可點；Deck 1280×800 上不出現（或者不妨礙）
- [ ] 五個 memory 開關在真機上操作一次，確認 P0.10 的 `syncing` 守衛仍成立（**改 UI 不能把 AGENTS.md
      的「同步不等於輸入」規則弄壞**：改版面後要重跑 `ui/fyneui/memory_state_test.go`）
- [ ] swap／gamedata 的 `Select` 換實作後要重跑 `ui/fyneui/swap_select_test.go`（P0.8 的下拉跟隨真實大小）
- [ ] 驗收過程中翻轉的 kernel 參數 / tmpfiles.d / log 狀態要收尾並列進回報（AGENTS.md 規則）

單測不可退步（改 native 必跑）：
- `ui/fyneui/memory_state_test.go`、`ui/fyneui/swap_select_test.go`、`ui/fyneui/refresh_test.go`

---

## 8. 本企劃明確不做

- **本輪不動 `web/src/`**（決策 4）。web 的 4K 與設計 token 缺陷仍在（§2.2、§2.6），留待那一輪。
- **不換 GUI 框架**（不引入 Slint／egui／Tauri／wails）。Fyne 的具體瑕疵在 §5.6、§5.7 內修。
- 不改 `core/` 的任何行為、API、locale 鍵語意（新增 nav rail 標籤鍵除外）。
- 不在 web 端補 sudo 密碼輸入（P0.5 移除的缺口是獨立議題）。
- 不動垂直節奏（§2.8：實測本來就是一致的）。
- 不在本文件處理 log 降噪／rotation、hugepages 持久化（那是 todo.md P0.9／P0.10 的待辦）。

---

## 附錄 A — 量測腳本（重現本文件的數字）

### A.1 web（從已存的截圖量）

```python
# python3 measure_web.py
from PIL import Image
import collections
for name in ('steam-deck-1280x800.png','4k-3840x2160.png'):
    im = Image.open('web/tests/screenshots/'+name).convert('RGB')
    px, W, H = im.load(), *im.size
    y = H//4
    row = [px[x,y] for x in range(W)]
    pagebg = row[0]
    card = collections.Counter(row).most_common(2)[1][0]
    xs = [x for x,c in enumerate(row) if c == card]
    print(name, im.size, 'content', min(xs), max(xs),
          'w=%d'%(max(xs)-min(xs)), 'ratio=%.2f'%((max(xs)-min(xs))/W))
```

Memory 行的 dead gap：先掃出六張卡的 y 範圍（看 x=300 這一整行是否為卡背景 `#16213e`），再到
`.memory-item` 所在 y 帶找文字像素（排除卡背景與邊框色），按 x 聚成 cluster（缺口 > 25 px 斷開）→
得到 §2.3 那張表的數字。**控制件與按鈕要用填色定位，不要用文字亮度**：`--accent-dim` = `#1a6fa0`
找按鈕、`--bg-input` = `#0f0f23` 找 select／input（§2.4 的數字就是這樣量的）。

### A.2 native（Xvfb + scrot）

`cmd/desktop` 需 `CGO_ENABLED=1`。`-ui native`。Xvfb + llvmpipe 可渲染（`docs/fyne-ui-plan.md` §5：`w.Canvas().Capture()` 在此環境恆黑，**一律用 scrot**）。

```bash
cat > /tmp/shot.sh <<'EOF'
#!/bin/bash
rm -f /tmp/.X97-lock /tmp/.X11-unix/X97          # 必清，否則 Xvfb 起不來、scrot 會連到舊 lock
pkill -9 -f cryoutils-ng-desktop; sleep 1
Xvfb :97 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
sleep 3
( DISPLAY=:97 /tmp/cryoutils-ng-desktop -ui native </dev/null >/tmp/run.log 2>&1 & )
sleep 10
DISPLAY=:97 scrot /tmp/native-1280.png
pkill -9 -f cryoutils-ng-desktop; pkill -9 Xvfb
EOF
nohup bash /tmp/shot.sh >/tmp/shot.out 2>&1 </dev/null &
```

三個環境眉角（都踩過）：

1. **Xvfb 一定要先清 `/tmp/.X<N>-lock` 與 `/tmp/.X11-unix/X<N>`**。殘留時 Xvfb 起不來，scrot 只會回 `Can't open X display ... It *is* running, yeah?`。
2. **在 agent 沙箱裡不能把 Xvfb + app 直接放前台**同一條命令跑（會拿不到 exit code）。要写成 script 再 `nohup bash script &`。
3. 4K 用 `-screen 0 3840x2160x24`；re-exec 與 scale 由 app 自己做，驗收看 run log 那行 `scale: display width 3840 px → re-exec with FYNE_SCALE=2`。

### A.3 native 的列位置量法

在同一 y 帶找亮字（`r,g,b > 150`），按 x 聚成 cluster（gap > 60px 斷開）→ 得到左欄與右欄的 x 範圍。§2.7 的 478 px 空白就是 `651 - 173`。標題字級比對用「CJK cluster 寬度 ÷ 字數」，不要用整行 bbox 高度（拉丁字的上下延伸會誤判，見 §2.8）。

---

## 附錄 B — 現況速查

### B.1 native（本輪要動的點）

| 檔案 | 位置 | 現況 | 要改成 |
|---|---|---|---|
| `ui/fyneui/app.go` | 根佈局 | `NewBorder(top, nil, nil, nil, vScroll)` | `NewBorder(top, nil, navRail, nil, vScroll)`（§5.7） |
| `ui/fyneui/app.go` | theme | 直接用 Fyne 預設 theme | 新增 theme wrapper，映 §4 token（§5.7） |
| 新增 `ui/fyneui/theme.go`（建議名） | — | 不存在 | 包住 `theme.DefaultTheme()`，覆蓋 `Size()`／`Color()` |
| `ui/fyneui/layout.go` | `contentMaxWidth` | 1100 logical；沒考慮 nav rail 的 200 px | 留 nav rail 寬度後重算內容上限（§4.3） |
| `ui/fyneui/sections_status.go` | 整個 | `NewGridWithColumns(2)` → 兩欄語意交錯、中間 478 px | 卡 A／卡 B + 固定 label 寬的定義列表（§5.4） |
| `ui/fyneui/sections_vram.go` | 整個 | 獨立卡，不顯示數值 | 資料併進卡 A 的一行，卡消失（§5.5） |
| `ui/fyneui/sections_presets.go` | 建立順序 | 第 4 個區塊 | 上移到 Status 正下方（決策 2） |
| `ui/fyneui/sections_swap.go` | Select 建立處 | Select 寬 1,085 px、按鈕滿版 | `widget.NewForm` + Select 上限 320 px（§5.2） |
| `ui/fyneui/sections_memory.go` | checkbox 列 | 值塞在 label 括號裡（無值欄）、未勾方塊對比低 | 三欄：名稱／固定寬值欄／開關（§5.3）；**保留 `syncing` 守衛** |
| `ui/fyneui/fonts.go` | 字體與字級 | 與 token 分開 | 走同一個 theme wrapper（§5.7） |
| `i18n/locales/{en,zh-TW}.json` | — | 無 nav rail 鍵 | 新增 5 個鍵（兩邊同步） |

### B.2 web（本輪不動，web 那一輪再動）

| 檔案 | 行 | 現況 | 要改成 |
|---|---|---|---|
| `web/src/App.css` | 4-19 | 只有顏色 token | 加 §4.1/§4.2/§4.3 token |
| `web/src/App.css` | 21 | `clamp(14px,1.2vw,18px)` | `calc(16px * var(--ui-scale))` + 三個 breakpoint |
| `web/src/App.css` | 34 | `max-width: 720px` | `min(1120px, 92vw)` |
| `web/src/App.css` | 86-90 | `.version` 負 margin | 放進 header grid |
| `web/src/App.css` | 129 | `auto-fit minmax(200px,1fr)` | 兩欄定義列表（§5.4） |
| `web/src/App.css` | 156-181 | `.swap-row` flex | `.form-grid`（§5.2） |
| `web/src/App.css` | 189-193 | `.memory-item` space-between | 三欄 grid + `tabular-nums` |
| `web/src/App.css` | 196+ | `.btn-toggle` | 真 switch（§4.3） |
| `web/src/App.tsx` | 區塊順序 | Presets 第 4 | 上移（§6） |
