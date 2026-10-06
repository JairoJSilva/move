//go:build windows

package platform

import (
	"strings"
	"testing"
)

func TestWindowsPaths(t *testing.T) {
	drivePath := `C:\Data\Folder\file.txt`
	normDrive := NormalizePath(drivePath)
	if !strings.HasPrefix(normDrive, `\\?\C:\`) {
		t.Fatalf("NormalizePath(%q) = %q, esperado prefixo \\\\?\\C:\\", drivePath, normDrive)
	}

	uncPath := `\\server\share\file.txt`
	normUNC := NormalizePath(uncPath)
	if !strings.HasPrefix(normUNC, `\\?\UNC\server\share`) {
		t.Fatalf("NormalizePath(%q) = %q, esperado prefixo \\\\?\\UNC\\", uncPath, normUNC)
	}

	alreadyExtended := `\\?\C:\Data\file.txt`
	normExt := NormalizePath(alreadyExtended)
	if normExt != alreadyExtended {
		t.Fatalf("NormalizePath(%q) = %q, esperado %q", alreadyExtended, normExt, alreadyExtended)
	}
}
