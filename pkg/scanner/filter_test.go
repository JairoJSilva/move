package scanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return nil }

func TestEvaluateItem_Filters(t *testing.T) {
	fi := mockFileInfo{
		name:    "relatorio_2024.pdf",
		size:    1024,
		mode:    0644,
		modTime: time.Date(2024, 5, 10, 10, 0, 0, 0, time.UTC),
		isDir:   false,
	}

	// 1. Teste Exclude Pattern
	opts := FilterOptions{
		Mode:            ModeFull,
		ExcludePatterns: []string{"*.pdf"},
	}
	needsCopy, reason, _ := EvaluateItem("/src/relatorio_2024.pdf", "relatorio_2024.pdf", fi, opts)
	if needsCopy {
		t.Fatalf("esperado falso para arquivo excluído por padrão, obteve true (%s)", reason)
	}

	// 2. Teste Year Filter Match
	optsYear := FilterOptions{
		Mode: ModeYear,
		Year: 2024,
	}
	needsCopy, _, _ = EvaluateItem("/src/relatorio_2024.pdf", "relatorio_2024.pdf", fi, optsYear)
	if !needsCopy {
		t.Fatalf("esperado true para ano coincidente (2024)")
	}

	// 3. Teste Year Filter Mismatch
	optsYearMismatch := FilterOptions{
		Mode: ModeYear,
		Year: 2023,
	}
	needsCopy, _, _ = EvaluateItem("/src/relatorio_2024.pdf", "relatorio_2024.pdf", fi, optsYearMismatch)
	if needsCopy {
		t.Fatalf("esperado false para ano divergente (2023)")
	}

	// 4. Teste DateRange
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	optsRange := FilterOptions{
		Mode:      ModeDateRange,
		StartDate: &start,
		EndDate:   &end,
	}
	needsCopy, _, _ = EvaluateItem("/src/relatorio_2024.pdf", "relatorio_2024.pdf", fi, optsRange)
	if !needsCopy {
		t.Fatalf("esperado true para data dentro do intervalo")
	}
}

func TestEvaluateItem_DeltaIdentical(t *testing.T) {
	tempDir := t.TempDir()
	destFile := filepath.Join(tempDir, "teste.txt")

	content := []byte("conteudo idêntico de teste")
	if err := os.WriteFile(destFile, content, 0644); err != nil {
		t.Fatalf("falha ao criar arquivo temporário: %v", err)
	}

	destFi, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("falha no stat: %v", err)
	}

	opts := FilterOptions{
		Mode:           ModeDelta,
		DestinationDir: tempDir,
	}

	needsCopy, reason, _ := EvaluateItem("/src/teste.txt", "teste.txt", destFi, opts)
	if needsCopy {
		t.Fatalf("esperado needsCopy=false no modo delta para arquivo idêntico, motivo: %s", reason)
	}
}
