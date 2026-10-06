package engine

import (
	"migrations-engine/pkg/platform"
)

// NormalizePath formata o caminho para o sistema operacional correspondente.
// Utiliza o pacote pkg/platform para tratar convenções POSIX e caminhos estendidos do Windows (\\?\).
func NormalizePath(p string) string {
	return platform.NormalizePath(p)
}
