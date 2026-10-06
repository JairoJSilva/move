package platform

import (
	"strings"
	"testing"
)

func TestPathsNormalization(t *testing.T) {
	// Testa caminho vazio
	if got := NormalizePath(""); got != "" {
		t.Fatalf("NormalizePath(\"\") = %q, esperado \"\"", got)
	}

	// Testa caminho comum
	p := "/var/log/app/../data"
	norm := NormalizePath(p)
	if strings.Contains(norm, "..") {
		t.Fatalf("NormalizePath(%q) contem '..': %q", p, norm)
	}

	// Testa IsLongPath
	shortPath := "/var/log"
	if IsLongPath(shortPath) {
		t.Fatalf("IsLongPath(%q) retornou true para caminho curto", shortPath)
	}

	hugePath := "/" + strings.Repeat("a", 5000)
	if !IsLongPath(hugePath) {
		t.Fatalf("IsLongPath(%q) retornou false para caminho de 5001 chars", hugePath)
	}
}
