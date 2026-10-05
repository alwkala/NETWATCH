// Package appdata locates the per-user data directory.
package appdata

import (
	"os"
	"path/filepath"
	"runtime"
)

// Dir returns the NETWATCH data directory:
//
//	Windows: %LOCALAPPDATA%\NetWatch\data
//	other:   $XDG_DATA_HOME/netwatch  (or ~/.local/share/netwatch)
func Dir() (string, error) {
	if runtime.GOOS == "windows" {
		if base := os.Getenv("LOCALAPPDATA"); base != "" {
			return filepath.Join(base, "NetWatch", "data"), nil
		}
	}
	if base := os.Getenv("XDG_DATA_HOME"); base != "" {
		return filepath.Join(base, "netwatch"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Local", "NetWatch", "data"), nil
	}
	return filepath.Join(home, ".local", "share", "netwatch"), nil
}

// DBPath is the SQLite file inside dir.
func DBPath(dir string) string { return filepath.Join(dir, "network.db") }
