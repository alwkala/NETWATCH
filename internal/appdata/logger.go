package appdata

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const (
	MaxLogSizeBytes = 5 * 1024 * 1024 // 5 MB
	MaxLogBackups   = 3
)

// LogPath returns the path to the main log file inside dir.
func LogPath(dir string) string {
	return filepath.Join(dir, "netwatch.log")
}

// RotatingFile implements a thread-safe rotating file writer.
type RotatingFile struct {
	mu       sync.Mutex
	dir      string
	filename string
	size     int64
	file     *os.File
}

func OpenRotatingFile(dir string) (*RotatingFile, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	rf := &RotatingFile{
		dir:      dir,
		filename: filepath.Join(dir, "netwatch.log"),
	}
	if err := rf.openCurrent(); err != nil {
		return nil, err
	}
	return rf, nil
}

func (rf *RotatingFile) openCurrent() error {
	fi, err := os.Stat(rf.filename)
	if err == nil {
		rf.size = fi.Size()
	} else if os.IsNotExist(err) {
		rf.size = 0
	} else {
		return err
	}

	f, err := os.OpenFile(rf.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	rf.file = f
	return nil
}

func (rf *RotatingFile) Write(p []byte) (n int, err error) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	writeLen := int64(len(p))
	if rf.size+writeLen > MaxLogSizeBytes {
		_ = rf.rotate()
	}

	if rf.file == nil {
		if err := rf.openCurrent(); err != nil {
			return 0, err
		}
	}

	n, err = rf.file.Write(p)
	rf.size += int64(n)
	return n, err
}

func (rf *RotatingFile) rotate() error {
	if rf.file != nil {
		_ = rf.file.Close()
		rf.file = nil
	}

	// Shift netwatch.3.log (delete), netwatch.2.log -> 3, netwatch.1.log -> 2
	for i := MaxLogBackups; i >= 1; i-- {
		oldPath := filepath.Join(rf.dir, fmt.Sprintf("netwatch.%d.log", i))
		if i == MaxLogBackups {
			_ = os.Remove(oldPath)
		} else {
			newPath := filepath.Join(rf.dir, fmt.Sprintf("netwatch.%d.log", i+1))
			_ = os.Rename(oldPath, newPath)
		}
	}

	// Rename current netwatch.log to netwatch.1.log
	firstBackup := filepath.Join(rf.dir, "netwatch.1.log")
	_ = os.Rename(rf.filename, firstBackup)

	return rf.openCurrent()
}

func (rf *RotatingFile) Close() error {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.file != nil {
		err := rf.file.Close()
		rf.file = nil
		return err
	}
	return nil
}

// NewLogger creates an slog.Logger that writes to both netwatch.log (with rotation)
// and os.Stderr. It returns the logger and a cleanup function to flush/close the file.
func NewLogger(dir string) (*slog.Logger, func(), error) {
	rf, err := OpenRotatingFile(dir)
	if err != nil {
		return nil, nil, err
	}

	mw := io.MultiWriter(os.Stderr, rf)
	handler := slog.NewTextHandler(mw, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)

	cleanup := func() {
		_ = rf.Close()
	}

	return logger, cleanup, nil
}
