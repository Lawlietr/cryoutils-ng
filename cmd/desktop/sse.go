// CryoUtilities
// Copyright (C) 2023 CryoByte33 and contributors to the CryoUtilities project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"cryoutils-ng/core"
)

// handleProgress serves an SSE stream of progress events.
// Clients connect here first, then authenticate via /api/auth before calling
// privileged endpoints. The SSE stream is open for the lifetime of the connection.
func handleProgress(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-core.ProgressChannel:
			if !ok {
				return
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: progress\ndata: %s\n\n", data)
			flusher.Flush()
		case <-time.After(30 * time.Second):
			// Send a keepalive comment every 30s to prevent proxy timeouts
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// isLocalhost checks whether the remote address is localhost.
func isLocalhost(addr net.Addr) bool {
	host, _, _ := net.SplitHostPort(addr.String())
	return host == "127.0.0.1" || host == "::1" || host == "localhost"
}
