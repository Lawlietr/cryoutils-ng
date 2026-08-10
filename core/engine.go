// Package core provides the UI-independent engine for CryoUtils NG.
// All tuning logic lives here; UI backends (CLI, web server, future Decky)
// call into this package.
package core

// Engine is the main struct that replaces the old global CryoUtils.
// It holds loggers, sudo password, and the OnProgress callback.
type Engine struct {
	// TODO: Phase 1 — populate fields from internal.Config
}
