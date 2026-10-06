//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// openFolder shows dir in Windows Explorer without flashing a console window.
func openFolder(dir string) error {
	explorer := "explorer.exe"
	if winDir := os.Getenv("WINDIR"); winDir != "" {
		p := filepath.Join(winDir, "explorer.exe")
		if _, err := os.Stat(p); err == nil {
			explorer = p
		}
	} else if _, err := os.Stat(`C:\Windows\explorer.exe`); err == nil {
		explorer = `C:\Windows\explorer.exe`
	}

	cmd := exec.Command(explorer, dir)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}
