//go:build windows

package appdata

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// OpenFolder opens dir in Windows Explorer without flashing a console window.
func OpenFolder(dir string) error {
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
