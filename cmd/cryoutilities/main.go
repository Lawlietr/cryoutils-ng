// CryoUtilities
// Copyright (C) 2023 CryoByte33 and contributors to the CryoUtilities project

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"cryoutils-ng/core"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/cristalhq/acmd"
)

func main() {
	// Ensure install directory exists
	if err := os.MkdirAll(core.InstallDirectory, 0755); err != nil {
		log.Panic(err)
	}
	// Delete old log file
	os.Remove(core.LogFilePath)
	// Create a log file
	logFile, err := os.OpenFile(core.LogFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()

	// Create engine with log output to file
	e := core.NewEngine(logFile, logFile)

	// Print the current version as a test
	e.InfoLog.Println("Current Version:", core.CurrentVersionNumber)

	// Provide a command structure for parsing
	cmds := []acmd.Command{
		{
			Name:        "desktop",
			Description: "Run the desktop web server",
			ExecFunc: func(context.Context, []string) error {
				return errors.New("desktop server must be run as a separate binary: go run ./cmd/desktop")
			},
		},
		{
			Name:        "swap",
			Description: "Change swap file size in increments of 1GB.",
			ExecFunc: func(_ context.Context, args []string) error {
				e.InfoLog.Println("Starting swap file resize...")
				size, err := strconv.Atoi(args[0])
				if err != nil {
					return err
				}
				err = e.ChangeSwapSize(size)
				if err != nil {
					return err
				}
				e.InfoLog.Println("Success!")
				return nil
			},
		},
		{
			Name:        "swappiness",
			Description: "Change swappiness to the specified value 0-200.",
			ExecFunc: func(_ context.Context, args []string) error {
				e.InfoLog.Println("Starting swappiness change...")
				swappiness := args[0]
				swappinessInt, err := strconv.Atoi(swappiness)
				if err != nil || swappinessInt < 0 || swappinessInt > 200 {
					return errors.New("invalid swappiness value")
				}
				err = e.ChangeSwappiness(swappiness)
				if err != nil {
					return err
				}
				e.InfoLog.Println("Success!")
				return nil
			},
		},
		{
			Name:        "hugepages",
			Description: "Enable or disable hugepages. Accepts 'true', 'false', 'enable' or 'disable'.\n\tRecommended: Enabled",
			ExecFunc: func(_ context.Context, args []string) error {
				arg := strings.ToLower(args[0])
				if arg == "true" || arg == "enable" {
					e.InfoLog.Println("Enabling HugePages...")
					return e.SetHugePages()
				} else if arg == "false" || arg == "disable" {
					e.InfoLog.Println("Disabling HugePages...")
					return e.RevertHugePages()
				}
				return errors.New("invalid argument provided")
			},
		},
		{
			Name:        "compaction_proactiveness",
			Description: "Set or revert compaction proactiveness. Accepts 'recommended' or 'stock'.",
			ExecFunc: func(_ context.Context, args []string) error {
				arg := strings.ToLower(args[0])
				if arg == "recommended" {
					e.InfoLog.Println("Setting Compaction Proactiveness...")
					return e.SetCompactionProactiveness()
				} else if arg == "stock" {
					e.InfoLog.Println("Reverting Compaction Proactiveness...")
					return e.RevertCompactionProactiveness()
				}
				return errors.New("invalid argument provided")
			},
		},
		{
			Name:        "defrag",
			Description: "Enable or disable hugepage defrag. Accepts 'true', 'false', 'enable' or 'disable'.\n\tRecommended: Disabled",
			ExecFunc: func(_ context.Context, args []string) error {
				arg := strings.ToLower(args[0])
				if arg == "true" || arg == "enable" {
					e.InfoLog.Println("Enabling HugePAge Defrag...")
					return e.RevertDefrag()
				} else if arg == "false" || arg == "disable" {
					e.InfoLog.Println("Disabling HugePage Defrag...")
					return e.SetDefrag()
				}
				return errors.New("invalid argument provided")
			},
		},
		{
			Name:        "page_lock_unfairness",
			Description: "Set or revert page lock unfairness. Accepts 'recommended' or 'stock'.",
			ExecFunc: func(_ context.Context, args []string) error {
				arg := strings.ToLower(args[0])
				if arg == "recommended" {
					e.InfoLog.Println("Setting Page Lock Unfairness...")
					return e.SetPageLockUnfairness()
				} else if arg == "stock" {
					e.InfoLog.Println("Reverting Page Lock Unfairness...")
					return e.RevertPageLockUnfairness()
				}
				return errors.New("invalid argument provided")
			},
		},
		{
			Name:        "shmem",
			Description: "Enable or disable shared memory. Accepts 'true', 'false', 'enable' or 'disable'.\n\tRecommended: Enabled",
			ExecFunc: func(_ context.Context, args []string) error {
				arg := strings.ToLower(args[0])
				if arg == "true" || arg == "enable" {
					e.InfoLog.Println("Setting Shared Memory...")
					return e.SetShMem()
				} else if arg == "false" || arg == "disable" {
					e.InfoLog.Println("Reverting Shared Memory...")
					return e.RevertShMem()
				}
				return errors.New("invalid argument provided")
			},
		},
		{
			Name:        "status",
			Description: "Print current system tuning status (scripting interface).",
			ExecFunc: func(context.Context, []string) error {
				return printStatus(e)
			},
		},
		{
			Name:        "recommended",
			Description: "Set all values to Cryo's recommendations.",
			ExecFunc: func(context.Context, []string) error {
				return e.UseRecommendedSettings()
			},
		},
		{
			Name:        "stock",
			Description: "Set all values to Valve defaults.",
			ExecFunc: func(context.Context, []string) error {
				return e.UseStockSettings()
			},
		},
	}

	// If no args are passed, assume "desktop"
	if len(os.Args) <= 1 {
		os.Args = []string{"", "desktop"}
	}

	// Basic program metadata
	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:         "cryoutils-ng",
		AppDescription:  "CryoUtils NG — Steam Deck performance utility (rewrite of CryoUtilities by CryoByte33).",
		PostDescription: "NOTE: You NEED to run this with sudo if not using GUI mode.",
		Version:         core.CurrentVersionNumber,
	})

	// Run the command parser
	if err := r.Run(); err != nil {
		e.ErrorLog.Println(err)
		os.Exit(1)
	}
}

// printStatus prints the current tuning status to stdout.
func printStatus(e *core.Engine) error {
	summary := e.GetStatusSummary()
	for k, v := range summary {
		fmt.Printf("%s: %s\n", k, v)
		e.InfoLog.Printf("%s: %s", k, v)
	}
	return nil
}
