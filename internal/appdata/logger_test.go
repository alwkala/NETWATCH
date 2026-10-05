package appdata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRotatingLogger(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "netwatch_log_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	logger, cleanup, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer cleanup()

	logger.Info("first test log message", "testKey", "testValue")

	logFile := filepath.Join(tempDir, "netwatch.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("expected log file to contain data, got empty")
	}
}
