package tray

import (
	"fmt"
	"net/url"
	"strings"
)

// Config armazena os parâmetros e callbacks do controlador de bandeja do MoveOps.
type Config struct {
	Title    string
	Port     int
	URL      string
	OnOpen   func()
	OnStatus func() string
	OnToggle func() (isPaused bool, message string)
	OnQuit   func()
}

// TrayController define a interface comum do ciclo de vida da bandeja.
type TrayController interface {
	Run() error
	Notify(title, message string) error
	Stop()
}

// ValidateBrowserURL valida se a URL informada é segura para abertura em navegadores.
// Bloqueia injeção de comandos de shell, injeção de parâmetros (-*, /*) e esquemas arbitrários (javascript:, file:, etc.).
func ValidateBrowserURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("URL não pode ser vazia")
	}

	// Previne injeção de parâmetros/argumentos de CLI (CWE-88)
	if strings.HasPrefix(rawURL, "-") || strings.HasPrefix(rawURL, "/") {
		return fmt.Errorf("URL não pode iniciar com traço ou barra")
	}

	// Bloqueia caracteres de controle e interpolação de shell
	for _, r := range rawURL {
		if r == '&' || r == '|' || r == ';' || r == '$' || r == '`' || r == '<' || r == '>' || r == '"' || r == '\'' || r == '\r' || r == '\n' || r == 0 {
			return fmt.Errorf("URL contém caracteres ilegais para chamada de navegador")
		}
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("formato de URL inválido: %w", err)
	}

	// Apenas esquemas HTTP e HTTPS são estritamente permitidos
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("esquema de protocolo não permitido: %q (apenas http e https são aceitos)", parsed.Scheme)
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL não contém host válido")
	}

	return nil
}

// OpenBrowser abre a URL informada no navegador padrão do sistema operacional de forma assíncrona após validação de segurança.
func OpenBrowser(rawURL string) error {
	if err := ValidateBrowserURL(rawURL); err != nil {
		return fmt.Errorf("abertura de navegador bloqueada por validação de segurança: %w", err)
	}
	return openBrowserPlatform(rawURL)
}
