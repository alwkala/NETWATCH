//go:build !windows

package appdata

import (
	"os/exec"
	"runtime"
)

// OpenURL launches target URL using the platform opener (open or xdg-open).
func OpenURL(targetURL string) error {
	opener := "xdg-open"
	if runtime.GOOS == "darwin" {
		opener = "open"
	}
	return exec.Command(opener, targetURL).Start()
}
