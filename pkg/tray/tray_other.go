//go:build !windows && !linux

package tray

import (
	"fmt"
	"os/exec"
	"runtime"
)

type fallbackTray struct {
	cfg   Config
	stopC chan struct{}
}

func NewTray(cfg Config) TrayController {
	return &fallbackTray{
		cfg:   cfg,
		stopC: make(chan struct{}),
	}
}

func (t *fallbackTray) Run() error {
	<-t.stopC
	return nil
}

func (t *fallbackTray) Notify(title, message string) error {
	return nil
}

func (t *fallbackTray) Stop() {
	select {
	case <-t.stopC:
	default:
		close(t.stopC)
	}
}

func openBrowserPlatform(rawURL string) error {
	cmdName := "open"
	if runtime.GOOS != "darwin" {
		cmdName = "xdg-open"
	}
	cmd := exec.Command(cmdName, rawURL)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("falha ao abrir navegador para %s: %w", rawURL, err)
	}
	return nil
}
