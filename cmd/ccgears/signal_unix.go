//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

func ignoreSIGINT() {
	signal.Ignore(syscall.SIGINT)
}

func resetSIGINT() {
	signal.Reset(syscall.SIGINT)
}

// killClaudeParent walks up the process tree from our PID to find the
// claude process, then sends it SIGINT. Only kills that specific process.
func killClaudeParent() {
	pid := os.Getpid()
	for {
		ppid := getParentPID(pid)
		if ppid <= 1 {
			break // reached init, stop
		}
		name := getProcessName(ppid)
		if name == "claude" || name == "node" {
			// Found it — send SIGINT to just this process
			_ = syscall.Kill(ppid, syscall.SIGINT)
			return
		}
		pid = ppid
	}
	// Fallback: kill direct parent
	_ = syscall.Kill(syscall.Getppid(), syscall.SIGINT)
}

// getParentPID reads the parent PID of a given PID via ps.
func getParentPID(pid int) int {
	out, err := exec.Command("ps", "-o", "ppid=", "-p", fmt.Sprintf("%d", pid)).Output()
	if err != nil {
		return 0
	}
	var ppid int
	_, _ = fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &ppid)
	return ppid
}

// getProcessName reads the process name of a given PID via ps.
func getProcessName(pid int) string {
	out, err := exec.Command("ps", "-o", "comm=", "-p", fmt.Sprintf("%d", pid)).Output()
	if err != nil {
		return ""
	}
	name := strings.TrimSpace(string(out))
	return filepath.Base(name)
}
