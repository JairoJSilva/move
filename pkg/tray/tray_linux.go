//go:build linux

package tray

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

type linuxTray struct {
	cfg   Config
	mu    sync.Mutex
	stopC chan struct{}
}

// NewTray inicializa a bandeja e integração com desktop no Linux.
func NewTray(cfg Config) TrayController {
	if cfg.Title == "" {
		cfg.Title = "MoveOps"
	}
	return &linuxTray{
		cfg:   cfg,
		stopC: make(chan struct{}),
	}
}

func (t *linuxTray) Run() error {
	log.Printf("[MoveOps] Controlador de bandeja/desktop Linux ativo na porta %d", t.cfg.Port)

	// Emite notificação inicial de inicialização no desktop se ambiente gráfico estiver disponível
	_ = t.Notify("MoveOps Iniciado", fmt.Sprintf("Painel de migração ativo em %s", t.cfg.URL))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-t.stopC:
		log.Printf("[MoveOps] Bandeja Linux encerrada.")
	case sig := <-sigChan:
		log.Printf("[MoveOps] Sinal recebido na bandeja (%v). Encerrando...", sig)
		if t.cfg.OnQuit != nil {
			t.cfg.OnQuit()
		}
	}

	return nil
}

// findIconPath localiza o caminho de um ícone disponível para o notify-send
func findIconPath() string {
	candidates := []string{
		"/usr/share/icons/hicolor/scalable/apps/moveops.png",
		"/usr/share/pixmaps/moveops.png",
		"moveops.png",
		"logo-n-fundo.png",
	}
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}
	return ""
}

func (t *linuxTray) Notify(title, message string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Se notify-send estiver disponível no PATH, emite balão nativo no GNOME/KDE/XFCE
	if notifyBin, err := exec.LookPath("notify-send"); err == nil {
		args := []string{"-a", "MoveOps"}
		if icon := findIconPath(); icon != "" {
			args = append(args, "-i", icon)
		}
		args = append(args, title, message)
		_ = exec.Command(notifyBin, args...).Run()
	}

	log.Printf("[MoveOps] Notificação Desktop: [%s] %s", title, message)
	return nil
}

func (t *linuxTray) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	select {
	case <-t.stopC:
	default:
		close(t.stopC)
	}
}

// openBrowserPlatform invoca xdg-open após a validação rigorosa de URL realizada em ValidateBrowserURL.
func openBrowserPlatform(rawURL string) error {
	cmd := exec.Command("xdg-open", rawURL)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar xdg-open para %s: %w", rawURL, err)
	}
	return nil
}
