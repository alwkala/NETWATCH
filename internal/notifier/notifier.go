package notifier

// Notifier abstracts system desktop notifications across platforms.
type Notifier interface {
	NotifyNewDevice(title, details string)
	NotifyDeviceOffline(title, details string)
	NotifyNetworkChange(title, details string)
}

// NoopNotifier is a silent implementation used for testing or unsupported platforms.
type NoopNotifier struct{}

func (NoopNotifier) NotifyNewDevice(title, details string)    {}
func (NoopNotifier) NotifyDeviceOffline(title, details string) {}
func (NoopNotifier) NotifyNetworkChange(title, details string) {}
