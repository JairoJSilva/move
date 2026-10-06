package tray

import (
	"testing"
)

func TestTrayController(t *testing.T) {
	opened := false
	statusChecked := false
	toggled := false
	quit := false

	cfg := Config{
		Title: "MoveOps Test",
		Port:  8080,
		URL:   "http://localhost:8080",
		OnOpen: func() {
			opened = true
		},
		OnStatus: func() string {
			statusChecked = true
			return "Status: OK"
		},
		OnToggle: func() (bool, string) {
			toggled = true
			return false, "Retomado"
		},
		OnQuit: func() {
			quit = true
		},
	}

	trayCtrl := NewTray(cfg)
	if trayCtrl == nil {
		t.Fatal("NewTray retornou nil")
	}

	errNotify := trayCtrl.Notify("Teste", "Mensagem de teste")
	if errNotify != nil {
		t.Fatalf("falha no Notify: %v", errNotify)
	}

	cfg.OnOpen()
	if !opened {
		t.Fatal("OnOpen não executado")
	}

	resStatus := cfg.OnStatus()
	if !statusChecked || resStatus != "Status: OK" {
		t.Fatal("OnStatus falhou")
	}

	_, msg := cfg.OnToggle()
	if !toggled || msg != "Retomado" {
		t.Fatal("OnToggle falhou")
	}

	cfg.OnQuit()
	if !quit {
		t.Fatal("OnQuit falhou")
	}

	trayCtrl.Stop()
}

func TestValidateBrowserURL(t *testing.T) {
	// 1. URLs válidas e esperadas
	validURLs := []string{
		"http://localhost:8080",
		"http://127.0.0.1:8082",
		"http://localhost:8082/dashboard",
		"https://moveops.internal.corp:8443",
		"https://sub.domain.example.com/path?foo=bar#section",
	}

	for _, u := range validURLs {
		if err := ValidateBrowserURL(u); err != nil {
			t.Errorf("ValidateBrowserURL(%q) rejeitou indevidamente URL legítima: %v", u, err)
		}
	}

	// 2. Vetores maliciosos de injeção de comando, injeção de parâmetros e protocolos inseguros
	maliciousURLs := []string{
		"",                                       // Vazia
		"   ",                                    // Espaços em branco
		"http://localhost:8080 & calc.exe",       // Shell injection (&)
		"http://localhost:8080 | notepad",        // Shell pipe (|)
		"http://localhost:8080; reboot",          // Command separator (;)
		"http://localhost:8080 && echo pwned",    // Shell AND
		"http://localhost:8080 || echo pwned",    // Shell OR
		"http://localhost:8080`calc.exe`",        // Backtick evaluation
		"http://localhost:8080$(calc.exe)",       // Subshell evaluation ($)
		"http://localhost:8080 > /tmp/hacked",    // Redirection (>)
		"http://localhost:8080\ncalc.exe",        // Newline injection
		"http://localhost:8080\r\ncalc.exe",      // CRLF injection
		"http://localhost:8080\x00calc.exe",      // Null byte injection
		"javascript:alert(1)",                    // Protocolo script
		"file:///etc/passwd",                     // Protocolo local file
		"file:///C:/Windows/win.ini",             // Protocolo local Windows file
		"smb://attacker.com/evil",                // Protocolo SMB
		"data:text/html,<script>alert(1)</script>", // Data URI
		"--disable-web-security",                 // CLI Argument Injection (CWE-88)
		"-a evil",                                // CLI Option Injection
		"/bin/sh",                                // Path absoluto
		"cmd.exe",                                // Nome de executável
	}

	for _, bad := range maliciousURLs {
		if err := ValidateBrowserURL(bad); err == nil {
			t.Errorf("ValidateBrowserURL(%q) falhou em bloquear vetor malicioso!", bad)
		}
	}
}
