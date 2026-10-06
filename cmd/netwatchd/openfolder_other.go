//go:build !windows

package main

import (
	"errors"
	"os/exec"
	"runtime"
)

// openFolder uses the platform opener where one exists (macOS open, Linux xdg-open).
func openFolder(dir string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", dir).Start()
	case "linux":
		return exec.Command("xdg-open", dir).Start()
	}
	return errors.New("opening the data folder is not supported on this platform")
}
