package autostart

// Manager handles configuring the application to launch on user login.
type Manager interface {
	Set(enabled bool) error
	IsEnabled() (bool, error)
}
