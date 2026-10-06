//go:build !windows

package platform

import (
	"path/filepath"
	"strings"
)

// MaxPathLength representa o limite PATH_MAX comum em distribuições Linux (4096 bytes/caracteres).
const MaxPathLength = 4096

// IsLongPath verifica se o caminho excede o limite comum de sistema de arquivos do Linux.
func IsLongPath(path string) bool {
	return len(path) >= MaxPathLength
}

// NormalizePath normaliza o caminho para padrões POSIX/Linux.
// Limpa caminhos redundantes e trata eventuais prefixos do Windows (\\?\, \\?\UNC\)
// recebidos inadvertidamente via chamadas de API ou configurações cruzadas.
func NormalizePath(path string) string {
	if path == "" {
		return ""
	}

	cleaned := filepath.Clean(path)

	// Remove prefixos estendidos de rede do Windows
	if strings.HasPrefix(cleaned, `\\?\UNC\`) {
		cleaned = "/" + filepath.ToSlash(strings.TrimPrefix(cleaned, `\\?\UNC\`))
	} else if strings.HasPrefix(cleaned, `\\?\`) {
		trimmed := strings.TrimPrefix(cleaned, `\\?\`)
		// Trata letras de unidade Windows como C:\... convertendo para /...
		if len(trimmed) >= 2 && trimmed[1] == ':' {
			cleaned = "/" + filepath.ToSlash(trimmed[2:])
		} else {
			cleaned = filepath.ToSlash(trimmed)
		}
	}

	return filepath.Clean(cleaned)
}

// ToExtendedPath em Linux retorna o caminho normalizado, visto que o padrão \\?\ é restrito ao Windows.
func ToExtendedPath(path string) string {
	return NormalizePath(path)
}
