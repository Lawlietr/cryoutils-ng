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
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cryoutils-ng/core"
)

const bindAddr = "127.0.0.1"

func main() {
	// Ensure install directory exists
	if err := os.MkdirAll(core.InstallDirectory, 0755); err != nil {
		log.Panic(err)
	}
	// Set up logging
	os.Remove(core.LogFilePath)
	logFile, err := os.OpenFile(core.LogFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()

	// Create engine
	e := core.NewEngine(logFile, logFile)
	e.InfoLog.Println("Current Version:", core.CurrentVersionNumber)

	// Generate random token
	token, err := generateToken()
	if err != nil {
		log.Panic(err)
	}

	// Get a free port
	port, err := getFreePort()
	if err != nil {
		log.Panic(err)
	}

	// Set up API handlers
	mux := http.NewServeMux()
	setupAPIRoutes(mux, e, token)

	// Serve embedded web build (Phase 4)
	// Strip query string from path before serving — http.FileServer treats
	// the entire path (including ?token=xxx from the browser URL) as a filename.
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/")
		if r.URL.Path == "" || r.URL.Path == "index.html" {
			data, err := fs.ReadFile(WebFS, "index.html")
			if err != nil {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write(data)
			return
		}
		http.FileServer(http.FS(WebFS)).ServeHTTP(w, r)
	}))

	// Start server
	addr := net.JoinHostPort(bindAddr, strconv.Itoa(port))
	url := fmt.Sprintf("http://%s/?token=%s", addr, token)

	e.InfoLog.Println("Starting desktop server on", url)
	fmt.Println(url)

	// Parse CLI flags for browser control
	var noBrowser bool
	var overrideBrowser string
	flag.BoolVar(&noBrowser, "no-browser", false, "Only print the URL, do not open a browser")
	flag.StringVar(&overrideBrowser, "browser", "", "Force use of a specific browser path (e.g. /usr/bin/chromium)")
	flag.Parse()

	// Open browser using the fallback chain (app-window → flatpak → steam → xdg-open)
	go func() {
		time.Sleep(500 * time.Millisecond)
		method := openURL(url, noBrowser, overrideBrowser, e.InfoLog)
		e.InfoLog.Printf("Browser launch method: %s", method)
	}()

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		e.ErrorLog.Println("Server error:", err)
		os.Exit(1)
	}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func getFreePort() (int, error) {
	addr, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer addr.Close()
	_, port, err := net.SplitHostPort(addr.Addr().String())
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(port)
}
