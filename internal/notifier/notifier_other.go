//go:build !windows

package notifier

import "log/slog"

// New returns a no-op notifier on non-Windows platforms.
func New(logger *slog.Logger) Notifier {
	return NoopNotifier{}
}
