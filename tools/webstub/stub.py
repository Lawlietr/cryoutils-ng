#!/usr/bin/env python3
"""Stand-in for `cryoutils-ng-desktop -ui web`, for UI work on the Dev VM only.

Serves the built web UI (web/dist) and answers /api/* from a JSON fixture, so
the status section can be seen in the states that only a real Deck reaches.
The Dev VM is an LXC: it has no swap file, no zram, no glxinfo, and web mode
has no sudo password entry. Running the real desktop binary there leaves every
status field empty or "unknown", which hides the parts being reviewed.

Bind is 127.0.0.1 only, mirroring the real server (AGENTS.md: loopback plus a
per-launch token). A stub on another interface would hand the whole web build
to anyone on the LAN, so a non-loopback --addr is refused unless the
WEBSTUB_ALLOW_LAN=1 environment variable is set.

    python3 tools/webstub/stub.py                       # deck-recommended
    python3 tools/webstub/stub.py --fixture deck-factory
    python3 tools/webstub/stub.py --list                # fixtures + notes

Then open http://127.0.0.1:8099 in a browser, or point Playwright at it:

    cd web && BASE_URL=http://127.0.0.1:8099 npx playwright test

See tools/webstub/README.md for what the fixtures claim and what they do not.
"""

import argparse
import functools
import json
import os
import sys
import time
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer

HERE = os.path.dirname(os.path.abspath(__file__))
REPO_ROOT = os.path.dirname(os.path.dirname(HERE))
FIXTURES_DIR = os.path.join(HERE, "fixtures")


def fixture_path(name):
    """Accept a bare fixture name, a path with .json, or an absolute path."""
    if os.path.isfile(name):
        return name
    candidate = name if name.endswith(".json") else name + ".json"
    return os.path.join(FIXTURES_DIR, candidate)


def load_fixture(path):
    with open(path, encoding="utf-8") as handle:
        raw = json.load(handle)
    # Keys starting with "_" are the fixture's own documentation and are never
    # sent to the client.
    note = raw.get("_note", "")
    status = {k: v for k, v in raw.get("status", {}).items() if not k.startswith("_")}
    libraries = raw.get("libraries", [])
    return status, libraries, note


class Handler(SimpleHTTPRequestHandler):
    """Static files out of web/dist, plus the four endpoints the UI calls."""

    def json_response(self, payload, code=200):
        body = json.dumps(payload).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(body)

    def token_ok(self):
        if not ARGS.token:
            return True
        query = self.path.split("?", 1)[1] if "?" in self.path else ""
        pairs = dict(p.split("=", 1) for p in query.split("&") if "=" in p)
        return pairs.get("token") == ARGS.token

    def api_path(self):
        """Path without the query string, or None when this is not /api/*."""
        path = self.path.split("?", 1)[0]
        return path if path.startswith("/api/") else None

    def do_GET(self):
        path = self.api_path()
        if path is None:
            self.serve_static()
            return
        if not self.token_ok():
            self.json_response({"success": False, "error": "unauthorized"}, 401)
            return
        if path == "/api/status":
            self.json_response({"success": True, "data": STATE["status"]})
        elif path == "/api/libraries":
            self.json_response({"success": True, "data": STATE["libraries"]})
        elif path == "/api/progress":
            if ARGS.no_sse:
                # Fatal for EventSource (non-200 is not retried), so a headless
                # screenshot mode that waits for the network to go idle can finish.
                self.json_response({"success": False, "error": "progress disabled"}, 501)
            else:
                self.progress_stream()
        else:
            self.json_response({"success": False, "error": "no stub for " + path}, 404)

    def do_POST(self):
        path = self.api_path()
        if path is None:
            self.json_response({"success": False, "error": "not an api path"}, 405)
            return
        if not self.token_ok():
            self.json_response({"success": False, "error": "unauthorized"}, 401)
            return
        length = int(self.headers.get("Content-Length") or 0)
        if length:
            self.rfile.read(length)
        # Nothing is written anywhere: the stub has no engine behind it. The UI
        # sees success and refreshes, so the same fixture comes back.
        self.json_response({"success": True})
        if ARGS.verbose:
            sys.stderr.write("POST %s -> success (nothing changed)\n" % path)

    def progress_stream(self):
        """Open SSE with comments only, so EventSource stays connected."""
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Cache-Control", "no-cache")
        self.send_header("Connection", "close")
        self.end_headers()
        try:
            while True:
                self.wfile.write(b": stub keepalive\n\n")
                self.wfile.flush()
                time.sleep(15)
        except (BrokenPipeError, ConnectionResetError, ValueError):
            pass

    def serve_static(self):
        path = self.path.split("?", 1)[0]
        target = self.translate_path(path)
        if not os.path.exists(target) and not path.startswith("/assets/"):
            # Single page app: unknown routes serve the shell.
            self.path = "/index.html"
        super().do_GET()

    def log_message(self, fmt, *args):
        if ARGS.verbose:
            sys.stderr.write("%s - %s\n" % (self.address_string(), fmt % args))


def list_fixtures():
    for name in sorted(os.listdir(FIXTURES_DIR)):
        if not name.endswith(".json"):
            continue
        _, _, note = load_fixture(os.path.join(FIXTURES_DIR, name))
        print("%-28s %s" % (name[:-5], note))
    return 0


def main():
    global ARGS, STATE
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--fixture", default="deck-recommended",
                        help="fixture name in tools/webstub/fixtures, or a path")
    parser.add_argument("--dist", default=os.path.join(REPO_ROOT, "web", "dist"),
                        help="built web UI (default: web/dist)")
    parser.add_argument("--addr", default="127.0.0.1",
                        help="bind address (default: 127.0.0.1)")
    parser.add_argument("--port", type=int, default=8099)
    parser.add_argument("--token", default="",
                        help="require this token, like the real server does")
    parser.add_argument("--verbose", action="store_true", help="log each request")
    parser.add_argument("--no-sse", action="store_true",
                        help="/api/progress answers 501 instead of holding the connection "
                             "open; use it with headless screenshot modes that wait for "
                             "the network to go idle")
    parser.add_argument("--list", action="store_true", help="list fixtures and exit")
    ARGS = parser.parse_args()

    if ARGS.list:
        return list_fixtures()

    loopback = (ARGS.addr == "localhost" or ARGS.addr.startswith("127.")
                or ARGS.addr == "::1")
    if not loopback and os.environ.get("WEBSTUB_ALLOW_LAN") != "1":
        sys.stderr.write(
            "refusing to bind %s: this stub has no auth. Loopback only, or set "
            "WEBSTUB_ALLOW_LAN=1 if you really mean it.\n" % ARGS.addr)
        return 2

    path = fixture_path(ARGS.fixture)
    if not os.path.isfile(path):
        sys.stderr.write("no fixture %s (try --list)\n" % path)
        return 2
    status, libraries, note = load_fixture(path)

    if not os.path.isfile(os.path.join(ARGS.dist, "index.html")):
        sys.stderr.write("no web build in %s - run: cd web && npm ci && npm run build\n"
                         % ARGS.dist)
        return 2

    STATE = {"status": status, "libraries": libraries}

    # directory= has to go through the constructor: SimpleHTTPRequestHandler.__init__
    # replaces a None directory with the current working directory, so assigning the
    # class attribute would serve the wrong tree.
    handler = functools.partial(Handler, directory=ARGS.dist)
    server = ThreadingHTTPServer((ARGS.addr, ARGS.port), handler)
    server.daemon_threads = True
    print("stub    : http://%s:%d/  (no token)" % (ARGS.addr, ARGS.port)
          if not ARGS.token else "stub    : http://%s:%d/?token=%s" % (ARGS.addr, ARGS.port, ARGS.token))
    print("fixture : %s\n            %s" % (path, note))
    print("ui build: %s" % ARGS.dist)
    print("playwright: cd web && BASE_URL=http://%s:%d npx playwright test"
          % (ARGS.addr, ARGS.port))
    sys.stdout.flush()
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nstopped")
    return 0


if __name__ == "__main__":
    sys.exit(main())
