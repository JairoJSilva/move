package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"migrations-engine/pkg/ratelimit"
	"migrations-engine/pkg/scanner"
)

func TestCopyFileStream_Success(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "origem", "dados.bin")
	destFile := filepath.Join(tempDir, "destino", "dados.bin")

	if err := os.MkdirAll(filepath.Dir(srcFile), 0755); err != nil {
		t.Fatal(err)
	}

	payload := []byte("HyperSync Super High Throughput Test Data 1234567890")
	if err := os.WriteFile(srcFile, payload, 0644); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(srcFile)
	if err != nil {
		t.Fatal(err)
	}

	limiter := ratelimit.NewLimiter(0, 0)
	res := CopyFileStream(context.Background(), srcFile, destFile, fi, limiter)
	if res.Error != nil {
		t.Fatalf("falha inesperada na cópia: %v", res.Error)
	}

	if res.BytesTransferred != int64(len(payload)) {
		t.Fatalf("bytes transferidos divergentes: esperado %d, obteve %d", len(payload), res.BytesTransferred)
	}

	if res.SourceHash == "" || res.SourceHash != res.DestHash {
		t.Fatalf("hashes inválidos ou divergentes: src=%s dest=%s", res.SourceHash, res.DestHash)
	}

	// Verifica se destino existe e conteúdo confere
	destData, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(destData) != string(payload) {
		t.Fatalf("conteúdo do destino divergente do arquivo de origem")
	}
}

func TestEngine_JobLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	destDir := filepath.Join(tempDir, "dest")
	auditDir := filepath.Join(tempDir, "audit")

	_ = os.MkdirAll(srcDir, 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "doc1.txt"), []byte("teste 1"), 0644)
	_ = os.WriteFile(filepath.Join(srcDir, "doc2.txt"), []byte("teste 2"), 0644)

	eng := NewEngine()

	cfg := JobConfig{
		JobID:          "test-job-01",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull,
		},
		Concurrency: 2,
		AuditDir:    auditDir,
	}

	if err := eng.StartJob(cfg); err != nil {
		t.Fatalf("falha ao iniciar job: %v", err)
	}

	// Aguarda conclusão (máximo 5s)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		m := eng.GetMetrics()
		if m.Status == StateCompleted {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	m := eng.GetMetrics()
	if m.Status != StateCompleted {
		t.Fatalf("job não completou dentro do prazo: status=%s", m.Status)
	}

	if m.FilesCopied != 2 {
		t.Fatalf("esperado 2 arquivos copiados, obteve %d", m.FilesCopied)
	}

	summary := eng.GetExecutiveSummary()
	if summary == nil {
		t.Fatalf("resumo executivo nulo")
	}
	if summary.TotalFilesCopied != 2 {
		t.Fatalf("esperado 2 arquivos copiados no sumário executivo, obteve %d", summary.TotalFilesCopied)
	}
}

func TestEngine_DefaultSafeProductionFallback(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src_safe")
	destDir := filepath.Join(tempDir, "dest_safe")
	_ = os.MkdirAll(srcDir, 0755)

	eng := NewEngine()

	// Configuração sem especificar Concurrency, MaxBandwidthMB ou MaxIOPS
	cfg := JobConfig{
		JobID:          "safe-fallback-job",
		SourceDir:      srcDir,
		DestinationDir: destDir,
	}

	if err := eng.StartJob(cfg); err != nil {
		t.Fatalf("falha ao iniciar job com defaults seguros: %v", err)
	}
	defer func() { _ = eng.StopJob() }()

	metrics := eng.GetMetrics()
	if metrics.LimitBandwidthMB != DefaultSafeBandwidthMB {
		t.Fatalf("esperado LimitBandwidthMB = %f, obteve %f", DefaultSafeBandwidthMB, metrics.LimitBandwidthMB)
	}
	if metrics.LimitIOPS != DefaultSafeIOPS {
		t.Fatalf("esperado LimitIOPS = %d, obteve %d", DefaultSafeIOPS, metrics.LimitIOPS)
	}
	if eng.currentJob.Concurrency != DefaultSafeConcurrency {
		t.Fatalf("esperado Concurrency = %d, obteve %d", DefaultSafeConcurrency, eng.currentJob.Concurrency)
	}
}
