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
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"cryoutils-ng/core"
)

const bindAddr = "127.0.0.1"

func main() {
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
	mux.Handle("/", http.FileServer(http.FS(WebFS)))

	// Start server
	addr := net.JoinHostPort(bindAddr, strconv.Itoa(port))
	url := fmt.Sprintf("http://%s/?token=%s", addr, token)

	e.InfoLog.Println("Starting desktop server on", url)
	fmt.Println(url)

	// Open browser
	go func() {
		time.Sleep(500 * time.Millisecond)
		_ = exec.Command("xdg-open", url).Start()
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
