//go:build windows

package appdata

import (
	"os/exec"
	"syscall"
)

// OpenURL launches target URL in the user's default browser on Windows.
func OpenURL(targetURL string) error {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}
