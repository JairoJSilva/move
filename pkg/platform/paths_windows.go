//go:build windows

package platform

import (
	"path/filepath"
	"strings"
)

// MaxPathLength define o limite MAX_PATH tradicional da API Win32 (260 caracteres).
const MaxPathLength = 260

// IsLongPath retorna true se o caminho atingir ou exceder o limite tradicional de 260 caracteres.
func IsLongPath(path string) bool {
	return len(path) >= MaxPathLength
}

// NormalizePath formata o caminho para o ambiente Windows, aplicando o prefixo estendido \\?\
// (ou \\?\UNC\ para compartilhamentos de rede) para superar o limite de 260 caracteres da API Win32.
func NormalizePath(path string) string {
	if path == "" {
		return ""
	}

	// Normaliza separadores de diretório e limpa redundâncias (. e ..)
	cleaned := filepath.Clean(path)

	// Se já possui o prefixo estendido UNC: \\?\UNC\
	if strings.HasPrefix(cleaned, `\\?\UNC\`) {
		return cleaned
	}

	// Se já possui o prefixo estendido de drive local: \\?\
	if strings.HasPrefix(cleaned, `\\?\`) {
		return cleaned
	}

	// Compartilhamento de rede UNC: \\servidor\compartilhamento\...
	if strings.HasPrefix(cleaned, `\\`) {
		serverShare := strings.TrimPrefix(cleaned, `\\`)
		return `\\?\UNC\` + serverShare
	}

	// Caminho absoluto de unidade de disco: C:\... ou D:\...
	if len(cleaned) >= 2 && cleaned[1] == ':' {
		return `\\?\` + cleaned
	}

	// Caminhos relativos que ultrapassem 260 caracteres:
	// A API Win32 exige caminho absoluto para o prefixo \\?\ funcionar.
	if len(cleaned) >= MaxPathLength {
		if abs, err := filepath.Abs(cleaned); err == nil {
			if strings.HasPrefix(abs, `\\`) {
				return `\\?\UNC\` + strings.TrimPrefix(abs, `\\`)
			}
			return `\\?\` + abs
		}
	}

	return cleaned
}

// ToExtendedPath garante a conversão do caminho para o padrão estendido \\?\ do Windows.
func ToExtendedPath(path string) string {
	return NormalizePath(path)
}
