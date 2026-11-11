#!/usr/bin/env bash
# UI-1 screenshot baseline (native Fyne UI).
#
# Builds cmd/desktop with CGO_ENABLED=1, then renders the native page under
# Xvfb at each resolution x language and captures it with scrot.
# Fyne's own w.Canvas().Capture() reads a black front buffer under
# Xvfb + llvmpipe, so every capture here goes through scrot.
#
# Usage: capture.sh <tag> ["<res> ..."] ["<lang> ..."]
#   tag  : file-name prefix, e.g. "before" or "after-step3"
#   res  : WxH, default "1280x800 2560x1440 3840x2160"
#   lang : i18n code, default "zh-TW en"
#
# Run it detached: nohup bash tools/ui1shots/capture.sh before >/tmp/ui1cap.log 2>&1 &
# The app must not be started in the foreground of a tool call (no exit code),
# and Xvfb refuses to start when a stale lock is present.
set -u

GO=${GO:-/usr/local/go/bin/go}
SHOTS_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_DIR=$(cd "$SHOTS_DIR/../.." && pwd)
BUILD_DIR="$SHOTS_DIR/.build"
BASELINE_DIR="$SHOTS_DIR/baseline"
LOCK_DIR="$SHOTS_DIR/.build/lock"
BIN="$BUILD_DIR/cryoutils-ng-desktop"

TAG=${1:-before}
RESOLUTIONS=${2:-"1280x800 2560x1440 3840x2160"}
LANGUAGES=${3:-"zh-TW en"}
DISPLAY_NUM=97
APP_WAIT=${APP_WAIT:-15}

PROFILE_DIR="$HOME/.cryoutils_ng"
CONFIG_FILE="$PROFILE_DIR/ui.json"

mkdir -p "$BUILD_DIR" "$BASELINE_DIR" "$PROFILE_DIR"

# Two capture runs at once share one X display and step on each other: scrot then
# leaves extra files named "..._000.png" behind and the screenshots no longer
# belong to one build.
if ! mkdir "$LOCK_DIR" 2>/dev/null; then
  echo "Another capture run is in progress ($LOCK_DIR exists). Remove it if the run died."
  exit 1
fi
trap 'rmdir "$LOCK_DIR" 2>/dev/null' EXIT

echo "== build (CGO_ENABLED=1) =="
if ! (cd "$REPO_DIR" && CGO_ENABLED=1 "$GO" build -o "$BIN" ./cmd/desktop); then
  echo "BUILD FAILED"
  exit 1
fi
echo "binary: $(ls -l "$BIN" | awk '{print $5" bytes"}')"

scale_from_log() {
  grep -o 'scale: display width [0-9]* px → [^"]*' "$1" 2>/dev/null | tr '\n' ' | '
  grep -o 'FYNE_SCALE=[0-9.]*' "$1" 2>/dev/null | head -1
}

capture_one() {
  local res=$1 lang=$2
  local width=${res%x*} height=${res#*x}
  local png="$BASELINE_DIR/$TAG-$res-$lang.png"
  local runlog="$BUILD_DIR/run-$res-$lang.log"

  printf '{"language":"%s"}\n' "$lang" > "$CONFIG_FILE"

  pkill -9 -f 'cryoutils-ng-desktop' >/dev/null 2>&1
  sleep 1
  rm -f "/tmp/.X$DISPLAY_NUM-lock" "/tmp/.X11-unix/X$DISPLAY_NUM"

  Xvfb ":$DISPLAY_NUM" -screen 0 "${width}x${height}x24" -nolisten tcp \
    > "$BUILD_DIR/xvfb-$res.log" 2>&1 &
  local xvfb_pid=$!
  sleep 3

  ( DISPLAY=":$DISPLAY_NUM" "$BIN" -ui native </dev/null > "$runlog" 2>&1 & )
  sleep "$APP_WAIT"

  if ! DISPLAY=":$DISPLAY_NUM" scrot "$png" 2> "$BUILD_DIR/scrot-$res-$lang.err"; then
    echo "  SCROT FAILED $res/$lang: $(cat "$BUILD_DIR/scrot-$res-$lang.err")"
  fi

  # Kill the app first, then the server, and by PID: pkill -9 on Xvfb makes bash
  # print a job-control "Killed" line that looks like a failure in the log.
  pkill -9 -f 'cryoutils-ng-desktop' >/dev/null 2>&1
  sleep 1
  kill "$xvfb_pid" >/dev/null 2>&1
  wait "$xvfb_pid" 2>/dev/null
  rm -f "/tmp/.X$DISPLAY_NUM-lock" "/tmp/.X11-unix/X$DISPLAY_NUM"
  sleep 1

  local size="missing"
  [ -f "$png" ] && size=$(identify -format '%wx%h %b' "$png" 2>/dev/null || ls -l "$png" | awk '{print $5}')
  echo "  $res $lang -> $(basename "$png") [$size] scale: $(scale_from_log "$runlog")"
}

echo "== captures (tag=$TAG) =="
for res in $RESOLUTIONS; do
  for lang in $LANGUAGES; do
    capture_one "$res" "$lang"
  done
done

printf '{"language":"zh-TW"}\n' > "$CONFIG_FILE"
echo "== done =="
ls -1 "$BASELINE_DIR"/$TAG-*.png 2>/dev/null | wc -l | sed 's/^/files with tag /'
