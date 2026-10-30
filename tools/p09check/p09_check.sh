#!/usr/bin/env bash
# CryoUtils NG — P0.9 acceptance check (read-only; changes nothing).
#
# Proves the five unit parameters are DETECTABLE without root and without any
# password prompt, and that nothing else (swap/zram/log perms) got damaged.
#
# Build/ship (Dev VM):  scp tools/p09check/p09_check.sh deck@<deck>:/tmp/
# Run on the Deck:      bash /tmp/p09_check.sh
#
# Booleans in `status` mean "the value equals the RECOMMENDED value", not
# "non-zero": compaction_proactiveness (recommended 0, stock 20), khugepaged/
# defrag (recommended 0, stock 1) and page_lock_unfairness (recommended 1,
# stock 5) are therefore expected to read false on a stock Deck.
BIN="$HOME/.cryoutils_ng/cryoutils-ng"
LOG="$HOME/.cryoutils_ng/cryoutils_ng.log"
fail=0

echo "== 0. 佈署中的二進位 =="
md5sum "$BIN" "$HOME/.cryoutils_ng/cryoutils-ng-desktop" 2>/dev/null

echo "== 1. 以 deck 身份（不給密碼）跑 status：不得出現 'Unable to get' =="
before=0; [ -f "$LOG" ] && before=$(wc -l < "$LOG")
out=$("$BIN" status 2>&1)
newlog=""; [ -f "$LOG" ] && newlog=$(tail -n +$((before+1)) "$LOG")
if printf '%s\n' "$newlog" | grep -qi "unable to get"; then
  echo "FAIL: status 期間 log 仍有 Unable to get"; fail=1
else
  echo "PASS: 讀取期間 log 無 Unable to get"
fi
printf '%s\n' "$out" | grep -E "HugePages|ShMem|CompactionProactiveness|Defrag|PageLockUnfairness" | sed 's/^/    /'

echo "== 2. app 讀值 vs kernel 實際值（右邊是 kernel ground truth）=="
g(){ printf '%s\n' "$out" | awk -F': ' -v k="$1" '$1==k{print $2}'; }
k(){ sed -e 's/.*\[\(.*\)\].*/\1/' "$1" 2>/dev/null; }
printf '    %-26s app=%-6s kernel=%s\n' "hugepages"                "$(g HugePages)"                "$(k /sys/kernel/mm/transparent_hugepage/enabled)"
printf '    %-26s app=%-6s kernel=%s\n' "shmem_enabled"            "$(g ShMem)"                    "$(k /sys/kernel/mm/transparent_hugepage/shmem_enabled)"
printf '    %-26s app=%-6s kernel=%s\n' "compaction_proactiveness" "$(g CompactionProactiveness)"  "$(k /proc/sys/vm/compaction_proactiveness)"
printf '    %-26s app=%-6s kernel=%s\n' "khugepaged/defrag"        "$(g Defrag)"                   "$(k /sys/kernel/mm/transparent_hugepage/khugepaged/defrag)"
printf '    %-26s app=%-6s kernel=%s\n' "page_lock_unfairness"     "$(g PageLockUnfairness)"       "$(k /proc/sys/vm/page_lock_unfairness)"
if [ "$(g HugePages)" = true ] && [ "$(g ShMem)" = true ]; then
  echo "PASS: HugePages / ShMem 讀到 true（這台 Deck 實際是 always / advise；修之前顯示 false）"
else
  echo "FAIL: HugePages=$(g HugePages) ShMem=$(g ShMem)"; fail=1
fi

echo "== 3. 讀取不得啟動 sudo（不得有任何密碼提示）=="
if printf '%s\n' "$out" | grep -qi "password for"; then echo "FAIL: 出現密碼提示"; fail=1
else echo "PASS: 無密碼提示、未碰 tty"; fi

echo "== 4. log 權限（应为 600，且不含密碼）=="
if [ -f "$LOG" ]; then
  perm=$(stat -c %a "$LOG"); echo "    log mode: $perm"
  [ "$perm" = 600 ] && echo "PASS: log 只有本人可讀" || { echo "FAIL: log mode=$perm"; fail=1; }
else
  echo "    （尚未產生 log —— 首次跑 native UI 後會產生）"
fi

echo "== 5. 沒有波及別的子系統 =="
printf '    swapfile: %s 個 | ' "$(grep -c '/home/swapfile' /proc/swaps)"
awk 'NR>1 && $1 ~ /zram/ {printf "zram0 仍在（size %s GB, used %s, priority %s）\n", $3/1048576, $4, $5}' /proc/swaps
printf '    /etc/tmpfiles.d 自訂档: %s 個（驗收後預期 0）\n' "$(ls /etc/tmpfiles.d/ 2>/dev/null | wc -l)"

echo
[ $fail -eq 0 ] && echo "RESULT: 全部 PASS" || echo "RESULT: 有 FAIL —— 把上方輸出整段貼回來"
