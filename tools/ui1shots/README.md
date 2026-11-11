# `tools/ui1shots` — native UI 截圖基線與幾何量測（UI-1）

UI-1 每一步都要能回答兩件事：**改動有沒有效果**、**有沒有回歸舊缺陷**。
這個目錄提供那個基準。設計規格與數字出自 `docs/ui-redesign-plan.md`。

## 兩個腳本

| 腳本 | 作用 |
|---|---|
| `capture.sh` | 用 `CGO_ENABLED=1` 建置 `cmd/desktop`，在 Xvfb 依解析度 × 語言渲染 native 頁面，用 `scrot` 存 PNG |
| `measure.py` | 從 PNG 量出欄位邊界、區塊分組、每行的 x 簇與最大空白（dead gap）、最寬控件 |

Fyne 自己的 `w.Canvas().Capture()` 在 Xvfb + llvmpipe 下讀 front buffer 恆黑，
所以截圖一律走 `scrot`（`docs/fyne-ui-plan.md` §5 已記錄）。

## 怎麼跑

```bash
# 截圖要在背景跑：Xvfb 與 app 不能放在 tool call 的前台（拿不到 exit code）
nohup bash tools/ui1shots/capture.sh before >/tmp/ui1cap.log 2>&1 &
tail -f /tmp/ui1cap.log

python3 tools/ui1shots/measure.py tools/ui1shots/baseline/before-*.png \
    --json tools/ui1shots/baseline/before-geometry.json

# 之後每一步：先截 after，再對比基線
nohup bash tools/ui1shots/capture.sh after-step3 >/tmp/ui1cap.log 2>&1 &
python3 tools/ui1shots/measure.py tools/ui1shots/baseline/after-step3-*.png \
    --compare tools/ui1shots/baseline/before-geometry.json
```

`capture.sh <tag> [解析度...] [語言...]`：預設 `1280x800 2560x1440 3840x2160` ×
`zh-TW en`。語言是寫 `~/.cryoutils_ng/ui.json` 選的（`{"language":"en"}`），
腳本跑完會還原成 `zh-TW`。

## 檔名與由來

`baseline/before-<解析度>-<語言>.png`，共 6 張，2026-11-11 在 Dev VM 產生：

- Dev VM 是 LXC，**沒有 swap file、沒有 zram、`glxinfo` 拿到 llvmpipe**，
  所以卡上的數值是假的（實測 `ZRAM 大小: 0 GB`、`VRAM: 93GB`）。
  **這些 PNG 是幾何基準，不是 SteamOS 的真實狀態**，不可當數值引用。
- 可用來引用的只有版面結構：視窗尺寸、內容寬度、置中偏移、控件寬度、行對齊。
- OCR（`tesseract -l chi_tra`）已驗證中文實際渲染成字（`系統狀態`、`Swap 檔案`、
  `共享記憶體`…）→ 字型 gate 通過，不是方框。

## 基線數字（2026-11-11）

| 解析度 | 視窗 | 內容左..右 | 內容寬 | 置中偏移 | 最寬控件 |
|---|---|---|---|---|---|
| 1280×800 | 1279×800（全螢幕） | 91..1188 | **1098** | +0.5 | **1086 px** |
| 2560×1440 | 2047×1152（80%） | 267..2283 | 2017 | −4.0 | 2004 px |
| 3840×2160 | 3071×1728（80%） | 823..3016 | **2194** | +0.5 | 2172 px |

這組數字同時證實 `docs/ui-redesign-plan.md` §2.7 的三件事：

1. `contentMaxWidth = 1100` 生效而且置中：1280 下內容寬 1098、偏移 0.5 px；
   4K 下內容寬 2194 ≈ 1100 logical × scale 2。
2. 2560×1440 不會被上限擋住：視窗 1024 logical < 1100 → 內容用到 2017 px。
   **upper bound 只在 4K 咬人**，所以驗 4K 才看得到 cap 的價值。
3. Select 實寬 1086 px（文件記 1,085 px）→ Step 4 要把它壓到 ≤320 logical px。
4. Status 各行 dead gap 431–475 px（文件 §2.7 記 478 px）→ Step 3 的定義列表要消掉它。

## 量測方式（改 theme 時要注意）

### 頁面背景怎麼找（Step 2 之後改了）
舊版取「整張圖最常見的顏色」當頁面背景。Section 有了面色之後，**卡片面色變成最常見的顏色**，
腳本会把卡片當背景（實測發生過：六張圖全部量出 `sections=0`）。現在的做法：
1. 從四角判斷視窗外是什麼：全螢幕時四角就是頁面色，2560/3840 時是 X root 的黑。
2. 視窗範圍 = 與那個顏色不同的像素外框（順帶驗證 `windowSizeForScreen()`：
   ≤1280 全螢幕、更大 80%）。
3. 頁面背景 = 視窗頂部 header 帶的最常見顏色（header 不是卡片）。

### 卡片面（card_faces）
- 面色由**面積**找出（內容區內最常見的顏色，且要與頁面背景差過 TOL），不寫死色值。
- `runs` = 一條垂直掃描線跨越的卡片面數。同一張卡内的文字行會打斷面色，所以短於
  `max(20, 2% × 掃描高度)` 的斷算為同一張：scale 1 卡内行距 16 px、scale 2 是 32 px，
  都不会被误判；卡与卡之間是 `--s-6` 32 logical px，会被正確切開。
- `coverage` = 內容區被卡片面覆蓋的比例，**雜訊最低，適合當閘門值**。
- 實測對照：Step 2 之前 `runs=0`、`coverage=0`；Step 2 之後 `coverage` 約 0.4–0.5。
  「六個區塊連成一條長欄」現在是一個數字，不是一句話。

### 其他
- 所有顏色都從圖檔本身取樣，**不寫死**任何 theme 色值。theme 換色會讓數字移動，
  而不是讓腳本壞掉。
- **`widest_control` 在 Step 2 之後不能直接當「控件寬」用**：最長的一段同色現在是卡片面本身
  （1280 從 1086 → 1096 px 就是卡片面，不是 Select）。Step 4 要卡「Select ≤ 320 logical px」
  之前，得先把與卡片面同色的像素排除，或改測控件的**描邊**而不是填充。
- 捲軸不算內容：視窗右緣 16 px 會被排除，否則內容寬會被捲軸灌水。
- 行（line）= 空行 ≤3 px 内相連；區塊（section）= 行之間空 ≥22 px；
  一行内的 x 簇以 60 px 空白切開。
  **注意**：卡片面出現之後，「行／區塊」分組失去意義（面色填滿每一行），
  Step 2 之後請看 `card_faces`、`content`、`widest_control`；`sections` 只對沒有面色
  的圖有意義。

## 尚未做（UI-1 Step 7 的範圍）

把這幾個數字變成斷言（例如「內容寬 ≤ 1100×scale + 20 px」、「Select 控件寬 ≤
320×scale」、「Memory 每一行的值欄起點相同，差 ≤ 2 px」、「`card_faces` ≥ 可見卡數」），
並放進 CI 或 `tools/` 的 runner。現在 `--compare` 只印差值，還不會好壞判定。
