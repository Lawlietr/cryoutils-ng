// CryoUtils NG — throwaway acceptance probe for P0.9 (not part of the app).
//
// It runs as an ordinary user against the real sudo path, which is the only way
// to exercise core.runSudoWithInput the same way the native UI does:
// engine.Password is set from stdin, then the memory-parameter toggles run for
// real, and every step prints the value that the kernel actually holds.
//
// Build (Dev VM):
//
//	CGO_ENABLED=0 /usr/local/go/bin/go build -o tools/p09probe ./tools/p09probe
//
// Run on the Deck (it asks for the sudo password itself, input not echoed):
//
//	~/.cryoutils_ng/p09probe
//
// A scripted run also works: `echo '<password>' | p09probe`.
//
// It always puts the values back the way it found them.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cryoutils-ng/core"
)

type param struct {
	name   string
	path   string
	rec    string
	def    string
	toggle func() error
	set    func() error
	revert func() error
}

func raw(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return "<unreadable: " + err.Error() + ">"
	}
	return strings.TrimSpace(string(b))
}

// current picks the active value out of a list-style sysfs file, without using
// any core helper, so the probe measures the kernel independently.
func current(path string) string {
	s := raw(path)
	for _, f := range strings.Fields(s) {
		if strings.HasPrefix(f, "[") {
			return strings.Trim(f, "[]")
		}
	}
	return s
}

// readPassword pulls the sudo password out of stdin when something is piped
// in (scripted runs), otherwise it prompts on the terminal with echo turned off
// through `stty`. Both paths end with the same string; nothing is ever written
// to disk. The earlier version only read stdin, so typing the command by hand
// looked like a hang with no output at all.
func readPassword() string {
	if fi, err := os.Stdin.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, "cannot open /dev/tty; pipe the password in instead: echo '<password>' | p09probe")
			return ""
		}
		defer tty.Close()
		fmt.Fprint(os.Stderr, "sudo password (input hidden): ")
		if c := exec.Command("stty", "-echo", "-F", "/dev/tty"); c.Run() != nil {
			// Older stty without -F: drive it through the tty as stdin instead.
			c := exec.Command("stty", "-echo")
			c.Stdin = tty
			c.Run()
		}
		pw, _ := bufio.NewReader(tty).ReadString('\n')
		fmt.Fprintln(os.Stderr)
		e := exec.Command("stty", "echo", "-F", "/dev/tty")
		if e.Run() != nil {
			e = exec.Command("stty", "echo")
			e.Stdin = tty
			e.Run()
		}
		return strings.TrimRight(pw, "\r\n")
	}
	pw, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimRight(pw, "\r\n")
}

func main() {
	fmt.Println("CryoUtils NG P0.9 acceptance probe - five unit parameters, real sudo, values restored at the end.")
	pw := readPassword()
	if pw == "" {
		fmt.Println("no password read; every privileged step will be refused (err=sudo password not provided)")
	}

	e := core.NewEngine(os.Stderr, os.Stderr)
	e.Password = pw
	e.SetProgressCallback(func(msg string) { fmt.Println("  progress:", msg) })

	fmt.Printf("euid=%d sudo_password_from_stdin=%d chars\n", os.Geteuid(), len(pw))

	params := []param{
		{"hugepages", core.UnitMatrix["hugepages"], core.RecommendedHugePages, core.DefaultHugePages, e.ToggleHugePages, e.SetHugePages, e.RevertHugePages},
		{"shmem_enabled", core.UnitMatrix["shmem_enabled"], core.RecommendedShMem, core.DefaultShMem, e.ToggleShMem, e.SetShMem, e.RevertShMem},
		{"compaction_proactiveness", core.UnitMatrix["compaction_proactiveness"], core.RecommendedCompactionProactiveness, core.DefaultCompactionProactiveness, e.ToggleCompactionProactiveness, e.SetCompactionProactiveness, e.RevertCompactionProactiveness},
		{"defrag", core.UnitMatrix["defrag"], core.RecommendedHugePageDefrag, core.DefaultHugePageDefrag, e.ToggleDefrag, e.SetDefrag, e.RevertDefrag},
		{"page_lock_unfairness", core.UnitMatrix["page_lock_unfairness"], core.RecommendedPageLockUnfairness, core.DefaultPageLockUnfairness, e.TogglePageLockUnfairness, e.SetPageLockUnfairness, e.RevertPageLockUnfairness},
	}

	failures := 0
	for i, p := range params {
		before := current(p.path)
		fmt.Printf("\n[%d] %s  path=%s  before=%q\n", i, p.name, p.path, before)

		err1 := p.toggle()
		mid := current(p.path)
		fmt.Printf("    toggle #1 -> err=%v value=%q\n", err1, mid)
		if err1 != nil {
			failures++
		}
		if mid == before {
			fmt.Printf("    !! the value did not change\n")
			failures++
		}

		err2 := p.toggle()
		after := current(p.path)
		fmt.Printf("    toggle #2 -> err=%v value=%q\n", err2, after)
		if err2 != nil {
			failures++
		}
		if after != before {
			fmt.Printf("    !! not restored (before=%q after=%q) - restore by hand\n", before, after)
			failures++
		}

		// The second pair of calls also checks the persistence layer, which is what
		// makes a setting survive a reboot (/etc/tmpfiles.d/<param>.conf).
		errSet := p.set()
		fmt.Printf("    Set()      -> err=%v value=%q\n", errSet, current(p.path))
		if errSet != nil {
			failures++
		}
		conf := "/etc/tmpfiles.d/" + p.name + ".conf"
		if _, err := os.Stat(conf); err == nil {
			body, _ := os.ReadFile(conf)
			fmt.Printf("    persistence OK: %s -> %q\n", conf, strings.TrimSpace(string(body)))
		} else {
			fmt.Printf("    persistence MISSING: %s (%v)\n", conf, err)
		}
		errRev := p.revert()
		fmt.Printf("    Revert()   -> err=%v value=%q\n", errRev, current(p.path))
		if errRev != nil {
			failures++
		}

		// Leave the kernel exactly as it was found.
		if current(p.path) != before {
			switch {
			case before == p.rec:
				_ = p.set()
			case before == p.def:
				_ = p.revert()
			default:
				fmt.Printf("    !! original value %q is neither recommended nor default - restore by hand\n", before)
				failures++
			}
			fmt.Printf("    restored to %q (target was %q)\n", current(p.path), before)
		}
	}

	fmt.Printf("\n== GetStatusSummary() as an ordinary user (no password needed to read) ==\n")
	for _, k := range []string{"HugePages", "ShMem", "CompactionProactiveness", "Defrag", "PageLockUnfairness", "Swappiness", "SwapFile", "SwapSizeGB", "ZramActive", "TotalSwapGB"} {
		fmt.Printf("  %-24s = %s\n", k, e.GetStatusSummary()[k])
	}

	fmt.Printf("\n== final kernel state (must match what it was at the start) ==\n")
	for _, p := range params {
		fmt.Printf("  %-26s %s\n", p.name, current(p.path))
	}

	fmt.Printf("\nprobe failures counted: %d\n", failures)
}
