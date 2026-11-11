#!/usr/bin/env python3
"""Geometry measurements for native UI screenshots (UI-1).

Colours are sampled from the image itself; nothing about the theme is hard-coded,
so a theme change moves the measurements instead of silently breaking them.

What it reports per screenshot:

  content      x extent of the page body (everything below the header row),
               its width and how far it sits from the screen centre.
               This is the number the cappedCenterLayout 1100 px cap controls.
  sections     groups of text lines separated by >= SECTION_GAP_ROWS empty rows.
               The top-most section is the header.
  lines        per text line: y range, the x clusters inside it, and the widest
               empty column run in that line ("dead gap" = the not-aligned number)
  control      the longest horizontal run of one non-background colour, which is
               a Select or Entry fill (the 1085 px problem in one number)

Usage:
  measure.py <png>...                       # table
  measure.py <png>... --json out.json       # machine readable
  measure.py <png>... --json out.json --compare baseline.json
"""
import argparse
import collections
import json
import sys

from PIL import Image

TOL = 30                 # colour distance that counts as "not the local background"
LINE_GAP_ROWS = 3        # empty rows that end a text line
SECTION_GAP_ROWS = 22    # empty rows that end a section
CLUSTER_GAP = 60         # empty columns that split a line into x clusters
MIN_LINE_H = 6           # thinner than this is an artefact, not text


def diff(a, b):
    return abs(a[0] - b[0]) + abs(a[1] - b[1]) + abs(a[2] - b[2])


def dominant(img, boxes):
    px = img.load()
    counter = collections.Counter()
    for x0, y0, x1, y1 in boxes:
        for y in range(y0, y1):
            for x in range(x0, x1):
                counter[px[x, y]] += 1
    return counter.most_common(1)[0][0] if counter else (0, 0, 0)


def row_flags(img, x0, x1, bg, y0=0, y1=None):
    """Per row in [y0, y1): does anything differ from bg between x0 and x1 (every 2 px)."""
    px = img.load()
    y1 = img.size[1] if y1 is None else y1
    flags = []
    for y in range(y0, y1):
        flags.append(any(diff(px[x, y], bg) > TOL for x in range(x0, x1, 2)))
    return flags


def group_rows(flags, gap, min_h):
    """Consecutive True rows, allowing up to `gap` blank rows inside one group."""
    out, y = [], 0
    while y < len(flags):
        if not flags[y]:
            y += 1
            continue
        start, blanks = y, 0
        while y < len(flags) and (flags[y] or blanks < gap):
            blanks = 0 if flags[y] else blanks + 1
            y += 1
        end = min(y - blanks, len(flags) - 1)
        if end - start + 1 >= min_h:
            out.append((start, end))
    return out


def page_background(img):
    """The app's own background colour.

    Sampled as the most common colour of the whole frame, not from the image
    edge: at 2560x1440 and 3840x2160 the window is 80% of the screen, so the
    edge is the X root (black) instead of the Fyne background.
    """
    w, h = img.size
    px = img.load()
    counter = collections.Counter()
    for y in range(0, h, 3):
        for x in range(0, w, 3):
            counter[px[x, y]] += 1
    return counter.most_common(1)[0][0]


def window_rect(img, bg):
    """Bounding box of the app window: everything painted in the app background.

    Also proves windowSizeForScreen() (full screen at <=1280, 80% above that).
    """
    w, h = img.size
    px = img.load()
    x0, y0, x1, y1 = w, h, -1, -1
    for y in range(h):
        for x in range(0, w, 2):
            if diff(px[x, y], bg) <= TOL:
                x0, x1 = min(x0, x), max(x1, x)
                y0, y1 = min(y0, y), max(y1, y)
    if x1 < x0:
        return (0, 0, w - 1, h - 1)
    return (x0, y0, x1, y1)


def line_clusters(img, line, x_lo, x_hi):
    x_lo, x_hi = max(0, x_lo), min(img.size[0] - 1, x_hi)
    if x_hi <= x_lo:
        return {"clusters": [], "dead_gap": 0}
    y0, y1 = line
    bg = dominant(img, [(x_lo, y0, x_hi + 1, y0 + 2)])
    px = img.load()
    hits = [False] * (x_hi - x_lo + 1)
    for y in range(y0, y1 + 1):
        for i, x in enumerate(range(x_lo, x_hi + 1)):
            if diff(px[x, y], bg) > TOL:
                hits[i] = True
    clusters, i = [], 0
    while i < len(hits):
        if not hits[i]:
            i += 1
            continue
        start, blanks = i, 0
        while i < len(hits) and (hits[i] or blanks < CLUSTER_GAP):
            blanks = 0 if hits[i] else blanks + 1
            i += 1
        clusters.append((x_lo + start, x_lo + i - blanks))
    gaps = [clusters[k + 1][0] - clusters[k][1] - 1 for k in range(len(clusters) - 1)]
    return {"clusters": [[a, b] for a, b in clusters],
            "dead_gap": max(gaps) if gaps else 0}


def widest_solid_run(img, bg, box):
    """Longest horizontal run of one colour inside the window, other than the
    app background. That run is a Select or Entry fill (or its border)."""
    px = img.load()
    x_lo, y_lo, x_hi, y_hi = box
    best = (0, None, 0, 0, 0)
    for y in range(y_lo + 1, y_hi, 2):
        run, start, cur = 0, 0, None
        for x in range(x_lo, x_hi + 1):
            c = px[x, y]
            if diff(c, bg) > TOL:
                if c != cur:
                    cur, start, run = c, x, 1
                else:
                    run += 1
            else:
                cur, run = None, 0
            if run > best[0]:
                best = (run, c, x - run + 1, x, y)
    return {"run": best[0], "colour": list(best[1]) if best[1] else None,
            "x0": best[2], "x1": best[3], "y": best[4]}


def measure(path):
    img = Image.open(path).convert("RGB")
    w, h = img.size
    bg = page_background(img)
    wx0, wy0, wx1, wy1 = window_rect(img, bg)
    # The vertical scrollbar sits against the right edge of the scroll container.
    # It is not part of the content column, so it must not widen the measurement.
    scroll_bar_w = 16
    x_lo, x_hi = wx0, wx1 - scroll_bar_w
    # Only rows inside the window: above and below it is the X root (black), which
    # is not the app background and would otherwise read as one giant content band.
    all_rows = [(a + wy0, b + wy0)
                for a, b in group_rows(row_flags(img, x_lo, x_hi, bg, wy0, wy1 + 1),
                                       LINE_GAP_ROWS, MIN_LINE_H)]
    if not all_rows:
        return {"file": path, "error": "nothing but background found"}

    header_y1 = all_rows[0][1]
    body_top = header_y1 + 1
    body_rows = [r for r in all_rows if r[0] >= body_top]

    px = img.load()
    left, right = w, -1
    for y in range(body_top, wy1 + 1, 2):
        for x in range(x_lo, x_hi + 1):
            if diff(px[x, y], bg) > TOL:
                left = min(left, x)
                break
        for x in range(x_hi, x_lo - 1, -1):
            if diff(px[x, y], bg) > TOL:
                right = max(right, x)
                break

    out = {"file": path, "screen": [w, h], "background": list(bg),
           "window": {"x0": wx0, "y0": wy0, "x1": wx1, "y1": wy1,
                      "w": wx1 - wx0 + 1, "h": wy1 - wy0 + 1,
                      "screen_ratio": round((wx1 - wx0 + 1) / w, 3)},
           "header_bottom_y": header_y1,
           "content": {"left": left, "right": right, "width": right - left + 1,
                       "ratio": round((right - left + 1) / w, 3),
                       "centre_offset": round((left + right) / 2 - (wx0 + wx1) / 2, 1)},
           "widest_control": widest_solid_run(img, bg, (wx0, wy0, wx1, wy1))}

    sections, cur = [], [body_rows[0]] if body_rows else []
    for prev, nxt in zip(body_rows, body_rows[1:]):
        if nxt[0] - prev[1] >= SECTION_GAP_ROWS:
            sections.append(cur)
            cur = [nxt]
        else:
            cur.append(nxt)
    if cur:
        sections.append(cur)

    out["section_count"] = len(sections)
    out["sections"] = []
    for sec in sections:
        y0, y1 = sec[0][0], sec[-1][1]
        out["sections"].append({
            "y": [y0, y1], "height": y1 - y0 + 1, "lines": len(sec),
            "line_detail": [dict(y=list(ln), **line_clusters(img, ln, left, right))
                            for ln in sec]})
    return out


def show(m):
    print(f"\n{m['file']}  screen={m.get('screen')}  header_bottom_y={m.get('header_bottom_y')}")
    if "error" in m:
        print("  ", m["error"])
        return
    c, win = m["content"], m["window"]
    print(f"  window {win['w']}x{win['h']} at ({win['x0']},{win['y0']}) "
          f"= {win['screen_ratio']} of screen width")
    print(f"  content {c['left']}..{c['right']} w={c['width']} ratio={c['ratio']} "
          f"centre_offset={c['centre_offset']}")
    wc = m["widest_control"]
    print(f"  sections={m['section_count']}  widest_control={wc['run']}px "
          f"(x {wc['x0']}..{wc['x1']} @y{wc['y']} rgb={wc['colour']})")
    for i, sec in enumerate(m["sections"]):
        print(f"    section {i}: y {sec['y'][0]}..{sec['y'][1]} "
              f"h={sec['height']:3d} lines={sec['lines']}")
        for ln in sec["line_detail"]:
            cl = " ".join(f"{a}..{b}" for a, b in ln["clusters"])
            print(f"        y {ln['y'][0]:4d}..{ln['y'][1]:4d} dead_gap={ln['dead_gap']:4d} "
                  f"[{cl}]")


def compare(m, base):
    ref = base.get(m["file"].replace("\\", "/").split("/")[-1])
    if not ref:
        print("  compare: no baseline entry for this file")
        return
    print(f"  vs baseline: content width {m['content']['width'] - ref['content']['width']:+d} px, "
          f"window width {m['window']['w'] - ref['window']['w']:+d} px, "
          f"sections {m['section_count'] - ref['section_count']:+d}, "
          f"widest control {m['widest_control']['run'] - ref['widest_control']['run']:+d} px")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("pngs", nargs="+")
    ap.add_argument("--json", dest="json_out")
    ap.add_argument("--compare")
    ap.add_argument("--quiet", action="store_true")
    args = ap.parse_args()

    results = [measure(p) for p in args.pngs]
    if not args.quiet:
        for m in results:
            show(m)
        if args.compare:
            print("\n== deltas ==")
            base = json.load(open(args.compare))
            for m in results:
                compare(m, base)
    if args.json_out:
        payload = {m["file"].split("/")[-1]: m for m in results}
        json.dump(payload, open(args.json_out, "w"), indent=1, ensure_ascii=False)
        print(f"wrote {args.json_out}")
    return 0 if all("error" not in m for m in results) else 1


if __name__ == "__main__":
    sys.exit(main())
