#!/usr/bin/env bash
# CryoUtils NG — P0.9 + P0.10 acceptance check (read-only; changes nothing).
#
# Proves three things about the five unit parameters (hugepages, shared memory
# in THP, compaction proactiveness, khugepaged defrag, page_lock_unfairness):
#
#   1. they are DETECTABLE as plain `deck`, with no password prompt (P0.9);
#   2. `status` reports the value the kernel really holds, and the boolean is
#      derived from the recommended value, not from "non-zero" (P0.10);
#   3. the application does not write anything by itself — leaving its window
#      open untouched must leave the kernel alone (P0.10: Fyne's
#      Check.SetChecked() used to fire OnChanged, so building the memory card
#      flipped every parameter that already sat on the recommended value).
#
# Build/ship (Dev VM):  scp tools/p09check/p09_check.sh deck@<deck>:/tmp/
# Run on the Deck:      bash /tmp/p09_check.sh
#
# For step 3 to mean anything, keep CryoUtils NG (native UI) OPEN while the
# script counts down, and do not click anything.
BIN="$HOME/.cryoutils_ng/cryoutils-ng"
LOG="$HOME/.cryoutils_ng/cryoutils_ng.log"
fail=0

# What this application calls "recommended" (core/config.go). A boolean in
# `status` means "the current value equals this", not "non-zero", so a stock
# Deck legitimately reads false for compaction / defrag / page_lock.
declare -A REC=(
  [hugepages]=always
  [shmem_enabled]=advise
  [compaction_proactiveness]=0
  [khugepaged_defrag]=0
  [page_lock_unfairness]=1
)
declare -A PATHS=(
  [hugepages]=/sys/kernel/mm/transparent_hugepage/enabled
  [shmem_enabled]=/sys/kernel/mm/transparent_hugepage/shmem_enabled
  [compaction_proactiveness]=/proc/sys/vm/compaction_proactiveness
  [khugepaged_defrag]=/sys/kernel/mm/transparent_hugepage/khugepaged/defrag
  [page_lock_unfairness]=/proc/sys/vm/page_lock_unfairness
)
# app status key -> parameter key
declare -A APPKEY=(
  [HugePages]=hugepages
  [ShMem]=shmem_enabled
  [CompactionProactiveness]=compaction_proactiveness
  [Defrag]=khugepaged_defrag
  [PageLockUnfairness]=page_lock_unfairness
)

# The kernel's own answer: sysfs/procfs list the active choice inside [brackets].
kernel_value() {
  local p=${PATHS[$1]} v
  [ -r "$p" ] || { printf '?\n'; return; }
  v=$(cat "$p")
  if printf '%s' "$v" | grep -q '\['; then
    printf '%s\n' "$v" | sed -e 's/.*\[\(.*\)\].*/\1/'
  else
    printf '%s\n' "$v" | tr -d ' \n'
  fi
}

echo "== 0. 佈署中的二進位 =="
md5sum "$BIN" "$HOME/.cryoutils_ng/cryoutils-ng-desktop" 2>/dev/null

echo "== 1. 以 deck 身份（不給密碼）跑 status：不得出現 'Unable to get' / 密碼提示 =="
before=0; [ -f "$LOG" ] && before=$(wc -l < "$LOG")
out=$("$BIN" status 2>&1)
newlog=""; [ -f "$LOG" ] && newlog=$(tail -n +$((before+1)) "$LOG")
if printf '%s\n' "$newlog" | grep -qi "unable to get"; then
  echo "FAIL: status 期間 log 仍有 Unable to get"; fail=1
else
  echo "PASS: 讀取期間 log 無 Unable to get"
fi
if printf '%s\n' "$out" | grep -qi "password for"; then
  echo "FAIL: 出現密碼提示（讀取不該要 root）"; fail=1
else
  echo "PASS: 無密碼提示、未碰 tty"
fi

echo "== 2. app 的 value 必須等於 kernel 的實際值，boolean 必須由 recommended 推導 =="
g(){ printf '%s\n' "$out" | awk -F': ' -v k="$1" '$1==k{print $2}'; }
for app in "${!APPKEY[@]}"; do
  par=${APPKEY[$app]}
  truth=$(kernel_value "$par")
  gotv=$(g "${app}Value")
  gotb=$(g "$app")
  wantb=$([ "$truth" = "${REC[$par]}" ] && echo true || echo false)
  printf '    %-24s kernel=%-8s appValue=%-8s appBool=%-5s (bool 應=%s)\n' "$par" "$truth" "$gotv" "$gotb" "$wantb"
  [ "$gotv" = "$truth" ] || { echo "FAIL: $par 的 value 讀錯"; fail=1; }
  [ "$gotb" = "$wantb" ] || { echo "FAIL: $par 的布爾值與 recommended 判斷不一致"; fail=1; }
done

echo "== 3. 五個 Value 鍵都要存在（UI 靠它們顯示真實值，讀不到時要顯示 unknown 而非 false）=="
for app in "${!APPKEY[@]}"; do
  v=$(g "${app}Value")
  case "$v" in
    "") echo "FAIL: status 缺少 ${app}Value 鍵"; fail=1 ;;
    unknown) echo "NOTE: ${app}Value = unknown（這台機器讀不到該參數，UI 會顯示 unknown）" ;;
    *) echo "PASS: ${app}Value = $v" ;;
  esac
done

echo "== 4. log 權限（必須 600，且不含密碼）=="
if [ -f "$LOG" ]; then
  perm=$(stat -c %a "$LOG"); echo "    log mode: $perm"
  [ "$perm" = 600 ] && echo "PASS: log 只有本人可讀" || { echo "FAIL: log mode=$perm"; fail=1; }
  if grep -qi "password for\|\[sudo\]" "$LOG"; then
    echo "FAIL: log 內出現 sudo 密碼提示殘跡"; fail=1
  else
    echo "PASS: log 內沒有 sudo 密碼殘跡"
  fi
else
  echo "    （尚未產生 log —— 首次跑 native UI 後會產生）"
fi

echo "== 5. app 開著不動，自己不得寫任何參數 =="
echo "    請確認 CryoUtils NG（native UI）已開啟且不去點它；10 秒後比對 kernel 值"
for i in 10 9 8 7 6 5 4 3 2 1; do printf '    %s\r' "$i"; sleep 1; done
printf '\n'
declare -A SNAP
for par in "${!PATHS[@]}"; do SNAP[$par]=$(kernel_value "$par"); done
# Let the UI run through at least two 5 s refresh cycles.
sleep 12
changed=0
for par in "${!PATHS[@]}"; do
  now=$(kernel_value "$par")
  if [ "$now" != "${SNAP[$par]}" ]; then
    echo "FAIL: $par 在你沒點擊的情況下被改寫：${SNAP[$par]} -> $now"; changed=1; fail=1
  fi
done
[ $changed -eq 0 ] && echo "PASS: 22 秒（含兩個刷新週期）內五個參數值皆未變動"
writes=$(printf '%s\n' "$newlog" | grep -cE "Writing (always|madvise|advise|never|0|1|20|5) for param")
if [ "${writes:-0}" != 0 ]; then
  echo "FAIL: 這段期間 log 出現 $writes 次參數寫入"; fail=1
else
  echo "PASS: log 未出現任何參數寫入"
fi

echo "== 6. 沒有波及別的子系統 =="
printf '    swapfile: %s 個 | ' "$(grep -c '/home/swapfile' /proc/swaps)"
awk 'NR>1 && $1 ~ /zram/ {printf "zram0 仍在（size %s GB, used %s, priority %s）\n", $3/1048576, $4, $5}' /proc/swaps
printf '    /etc/tmpfiles.d 自訂檔: %s 個\n' "$(ls /etc/tmpfiles.d/ 2>/dev/null | wc -l)"
for f in /etc/tmpfiles.d/*.conf; do [ -e "$f" ] || continue; printf '      %s: %s\n' "$(basename "$f")" "$(tail -1 "$f" | awk '{print $NF}')"; done
[ -d "$HOME/.cry_utilities" ] && echo "    原版 CryoUtilities 目錄仍在（共存正常）"

echo
[ $fail -eq 0 ] && echo "RESULT: 全部 PASS" || echo "RESULT: 有 FAIL —— 把上方輸出整段貼回來"
exit $fail
