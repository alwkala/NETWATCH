//go:build windows

package notifier

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unicode/utf16"
)

// WindowsNotifier sends native Windows 10/11 toast notifications
// via the user-mode Windows Runtime ToastNotificationManager.
type WindowsNotifier struct {
	logger *slog.Logger
}

// New creates a platform-appropriate Notifier for Windows.
func New(logger *slog.Logger) Notifier {
	return &WindowsNotifier{logger: logger}
}

func (w *WindowsNotifier) NotifyNewDevice(title, details string) {
	w.send(title, details)
}

func (w *WindowsNotifier) NotifyDeviceOffline(title, details string) {
	w.send(title, details)
}

func (w *WindowsNotifier) NotifyNetworkChange(title, details string) {
	w.send(title, details)
}

func (w *WindowsNotifier) send(title, details string) {
	// Execute asynchronously so notification dispatch never blocks the scan engine.
	go func() {
		if err := showToast(title, details); err != nil {
			if w.logger != nil {
				w.logger.Debug("failed to display windows toast", "err", err, "title", title)
			}
		}
	}()
}

func showToast(title, details string) error {
	powershellPath := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	if _, err := os.Stat(powershellPath); err != nil {
		powershellPath = "powershell.exe"
	}

	// PowerShell script to create and show Windows Runtime Toast
	script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null
$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$nodes = $template.GetElementsByTagName('text')
$nodes.Item(0).AppendChild($template.CreateTextNode(%q)) > $null
$nodes.Item(1).AppendChild($template.CreateTextNode(%q)) > $null
$toast = [Windows.UI.Notifications.ToastNotification]::new($template)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('NETWATCH').Show($toast)
`, title, details)

	// Encode as UTF-16LE base64 for PowerShell -EncodedCommand
	encoded := encodeUTF16LEBase64(script)

	cmd := exec.Command(powershellPath, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-EncodedCommand", encoded)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

func encodeUTF16LEBase64(s string) string {
	runes := []rune(s)
	u16 := utf16.Encode(runes)
	buf := make([]byte, len(u16)*2)
	for i, r := range u16 {
		buf[i*2] = byte(r)
		buf[i*2+1] = byte(r >> 8)
	}
	return base64.StdEncoding.EncodeToString(buf)
}
