//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func ignoreSIGINT() {
	// Windows doesn't support SIGINT ignore the same way; no-op.
}

func resetSIGINT() {
	// Windows doesn't support SIGINT reset; no-op.
}

// killClaudeParent on Windows sends a taskkill to the parent process.
func killClaudeParent() {
	ppid := os.Getppid()
	_ = exec.Command("taskkill", "/PID", fmt.Sprintf("%d", ppid), "/F").Run()
}
