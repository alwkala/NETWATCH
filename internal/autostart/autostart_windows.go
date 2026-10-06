//go:build windows

package autostart

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName    = "NetWatch"
)

type WindowsManager struct{}

func New() Manager {
	return &WindowsManager{}
}

// Set configures or removes the NetWatch entry in HKCU\Software\Microsoft\Windows\CurrentVersion\Run.
func (w *WindowsManager) Set(enabled bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("open HKCU run key: %w", err)
	}
	defer k.Close()

	if enabled {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("get executable path: %w", err)
		}
		absExe, err := filepath.Abs(exe)
		if err != nil {
			absExe = exe
		}
		// Write the executable path with --minimized argument
		cmdLine := fmt.Sprintf(`"%s" --minimized`, absExe)
		if err := k.SetStringValue(appName, cmdLine); err != nil {
			return fmt.Errorf("set run key value: %w", err)
		}
		return nil
	}

	// Remove value if present
	err = k.DeleteValue(appName)
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("delete run key value: %w", err)
	}
	return nil
}

// IsEnabled checks whether the NetWatch entry is active in HKCU Run.
func (w *WindowsManager) IsEnabled() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("open HKCU run key: %w", err)
	}
	defer k.Close()

	val, _, err := k.GetStringValue(appName)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	return val != "", nil
}
