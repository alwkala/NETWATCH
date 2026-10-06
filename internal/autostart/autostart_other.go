//go:build !windows

package autostart

type OtherManager struct{}

func New() Manager {
	return &OtherManager{}
}

func (o *OtherManager) Set(enabled bool) error {
	return nil
}

func (o *OtherManager) IsEnabled() (bool, error) {
	return false, nil
}
