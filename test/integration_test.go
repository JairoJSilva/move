package test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"migrations-engine/pkg/audit"
	"migrations-engine/pkg/engine"
	"migrations-engine/pkg/ratelimit"
	"migrations-engine/pkg/scanner"
)

// Helper: cria arquivo de teste com tamanho e mtime específicos
func createTestFile(t *testing.T, path string, size int64, modTime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("falha ao criar pasta para %s: %v", path, err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("falha ao criar arquivo %s: %v", path, err)
	}
	defer f.Close()

	if size > 0 {
		buf := make([]byte, 64*1024)
		var written int64
		for written < size {
			toWrite := int64(len(buf))
			if size-written < toWrite {
				toWrite = size - written
			}
			// Preenche com bytes pseudo-aleatórios deterministicos
			for i := range toWrite {
				buf[i] = byte((int64(i) + written) % 251)
			}
			n, err := f.Write(buf[:toWrite])
			if err != nil {
				t.Fatalf("falha ao escrever em %s: %v", path, err)
			}
			written += int64(n)
		}
	}

	_ = f.Close()
	if !modTime.IsZero() {
		_ = os.Chtimes(path, modTime, modTime)
	}
}

// Helper: aguarda a finalização do job do Engine
func waitForJobCompletion(t *testing.T, eng *engine.Engine, timeout time.Duration) engine.TelemetryMetrics {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		m := eng.GetMetrics()
		if m.Status == engine.StateCompleted || m.Status == engine.StateFailed || m.Status == engine.StateCancelled {
			return m
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timeout aguardando conclusão do job (status atual: %s)", eng.GetMetrics().Status)
	return eng.GetMetrics()
}

// ---------------------------------------------------------------------------
// 1. Full Sync: Migração Completa com Múltiplos Arquivos e Pastas Aninhadas
// ---------------------------------------------------------------------------
func TestIntegration_FullSync(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source_disk")
	destDir := filepath.Join(tempDir, "dest_disk")
	auditDir := filepath.Join(tempDir, "audit_dir")

	now := time.Now().Truncate(time.Second)

	filesSpec := []struct {
		relPath string
		size    int64
		modTime time.Time
	}{
		{"empty.txt", 0, now.Add(-10 * time.Hour)},
		{"small_doc.txt", 512, now.Add(-5 * time.Hour)},
		{"medium_blob.bin", 64 * 1024, now.Add(-2 * time.Hour)},
		{"large_data.dat", 2 * 1024 * 1024, now.Add(-24 * time.Hour)},
		{filepath.Join("nested", "level1", "config.json"), 1024, now.Add(-30 * time.Minute)},
		{filepath.Join("nested", "level1", "level2", "archive.tar"), 512 * 1024, now.Add(-12 * time.Hour)},
		{filepath.Join("nested", "sibling", "readme.md"), 2048, now.Add(-1 * time.Hour)},
	}

	var expectedTotalBytes int64
	for _, spec := range filesSpec {
		createTestFile(t, filepath.Join(srcDir, spec.relPath), spec.size, spec.modTime)
		expectedTotalBytes += spec.size
	}

	eng := engine.NewEngine()

	var progressUpdates []engine.TelemetryMetrics
	var mu sync.Mutex
	eng.SetProgressCallback(func(m engine.TelemetryMetrics) {
		mu.Lock()
		progressUpdates = append(progressUpdates, m)
		mu.Unlock()
	})

	cfg := engine.JobConfig{
		JobID:          "fullsync-test-01",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull,
		},
		Concurrency: 4,
		AuditDir:    auditDir,
	}

	if err := eng.StartJob(cfg); err != nil {
		t.Fatalf("falha ao iniciar Full Sync: %v", err)
	}

	metrics := waitForJobCompletion(t, eng, 10*time.Second)

	if metrics.Status != engine.StateCompleted {
		t.Fatalf("esperado status COMPLETED, obteve %s", metrics.Status)
	}
	if metrics.FilesCopied != int64(len(filesSpec)) {
		t.Fatalf("esperado %d arquivos copiados, obteve %d", len(filesSpec), metrics.FilesCopied)
	}
	if metrics.FilesFailed != 0 {
		t.Fatalf("esperado 0 falhas, obteve %d", metrics.FilesFailed)
	}
	if metrics.BytesTransferred != expectedTotalBytes {
		t.Fatalf("esperado %d bytes transferidos, obteve %d", expectedTotalBytes, metrics.BytesTransferred)
	}

	// Validação ponta a ponta dos arquivos no destino: Conteúdo, Hashes xxHash64 e ModTime
	for _, spec := range filesSpec {
		srcFile := filepath.Join(srcDir, spec.relPath)
		dstFile := filepath.Join(destDir, spec.relPath)

		dstFi, err := os.Stat(dstFile)
		if err != nil {
			t.Fatalf("arquivo de destino não encontrado: %s (%v)", spec.relPath, err)
		}

		if dstFi.Size() != spec.size {
			t.Errorf("[%s] tamanho divergente: esperado %d, destino tem %d", spec.relPath, spec.size, dstFi.Size())
		}

		// Validação de tolerância de timestamp (até 2 segundos para compensar resolução de FS)
		diffSecs := math.Abs(dstFi.ModTime().Sub(spec.modTime).Seconds())
		if diffSecs > 2.0 {
			t.Errorf("[%s] modTime divergente: src=%v, dest=%v (diff=%fs)", spec.relPath, spec.modTime, dstFi.ModTime(), diffSecs)
		}

		// Verificação de integridade via xxHash64
		srcHash, err := engine.ComputeFileHashXX64(srcFile)
		if err != nil {
			t.Fatalf("falha ao calcular hash de origem para %s: %v", spec.relPath, err)
		}
		dstHash, err := engine.ComputeFileHashXX64(dstFile)
		if err != nil {
			t.Fatalf("falha ao calcular hash de destino para %s: %v", spec.relPath, err)
		}

		if srcHash != dstHash {
			t.Errorf("[%s] hashes xxHash64 divergentes: src=%s, dst=%s", spec.relPath, srcHash, dstHash)
		}
	}

	// Validação do Relatório Executivo
	summary := eng.GetExecutiveSummary()
	if summary == nil {
		t.Fatalf("resumo executivo nulo")
	}
	if summary.TotalFilesCopied != int64(len(filesSpec)) {
		t.Errorf("summary: esperado %d arquivos copiados, obteve %d", len(filesSpec), summary.TotalFilesCopied)
	}
	if summary.TotalBytesTransferred != expectedTotalBytes {
		t.Errorf("summary: esperado %d bytes, obteve %d", expectedTotalBytes, summary.TotalBytesTransferred)
	}
	if len(summary.Errors) != 0 {
		t.Errorf("summary: esperado 0 erros, obteve %d", len(summary.Errors))
	}
}

// ---------------------------------------------------------------------------
// 2. Delta Sync: Apenas Arquivos Modificados/Novos são Copiados
// ---------------------------------------------------------------------------
func TestIntegration_DeltaSync(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "delta_src")
	destDir := filepath.Join(tempDir, "delta_dst")
	auditDir := filepath.Join(tempDir, "delta_audit")

	t0 := time.Now().Add(-2 * time.Hour).Truncate(time.Second)

	// Arquivos iniciais
	createTestFile(t, filepath.Join(srcDir, "doc_unchanged_1.txt"), 1024, t0)
	createTestFile(t, filepath.Join(srcDir, "doc_unchanged_2.txt"), 2048, t0)
	createTestFile(t, filepath.Join(srcDir, "doc_to_modify_content.txt"), 4096, t0)
	createTestFile(t, filepath.Join(srcDir, "doc_to_touch_time.txt"), 1024, t0)

	// Passo 1: Executa Baseline Full Sync
	engFull := engine.NewEngine()
	cfgFull := engine.JobConfig{
		JobID:          "baseline-job",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull,
		},
		Concurrency: 2,
		AuditDir:    auditDir,
	}
	if err := engFull.StartJob(cfgFull); err != nil {
		t.Fatalf("falha ao iniciar baseline: %v", err)
	}
	mFull := waitForJobCompletion(t, engFull, 5*time.Second)
	if mFull.FilesCopied != 4 {
		t.Fatalf("baseline: esperado 4 arquivos copiados, obteve %d", mFull.FilesCopied)
	}

	// Passo 2: Aplica alterações pontuais na origem
	tModified := time.Now().Truncate(time.Second)

	// Arquivo 1 alterado: conteúdo diferente e tamanho diferente
	createTestFile(t, filepath.Join(srcDir, "doc_to_modify_content.txt"), 8192, tModified)

	// Arquivo 2 alterado: mesmo tamanho, mas timestamp atualizado (+2 horas)
	createTestFile(t, filepath.Join(srcDir, "doc_to_touch_time.txt"), 1024, tModified)

	// Arquivo 3 novo: adicionado na origem
	createTestFile(t, filepath.Join(srcDir, "doc_brand_new.txt"), 3072, tModified)

	// Passo 3: Executa Delta Sync
	engDelta := engine.NewEngine()
	cfgDelta := engine.JobConfig{
		JobID:          "delta-job-01",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeDelta,
		},
		Concurrency: 2,
		AuditDir:    auditDir,
	}

	var skippedFilesCount atomic.Int64
	var copiedFilesCount atomic.Int64
	engDelta.SetEventCallback(func(evt audit.AuditEvent) {
		if evt.Action == audit.ActionSkipped {
			skippedFilesCount.Add(1)
		} else if evt.Action == audit.ActionCopied {
			copiedFilesCount.Add(1)
		}
	})

	if err := engDelta.StartJob(cfgDelta); err != nil {
		t.Fatalf("falha ao iniciar Delta Sync: %v", err)
	}

	mDelta := waitForJobCompletion(t, engDelta, 5*time.Second)

	if mDelta.Status != engine.StateCompleted {
		t.Fatalf("delta sync: esperado status COMPLETED, obteve %s", mDelta.Status)
	}

	// Esperado: 3 arquivos copiados (modify_content, touch_time, brand_new)
	// e 2 arquivos pulados (unchanged_1, unchanged_2)
	if mDelta.FilesCopied != 3 {
		t.Fatalf("delta sync: esperado 3 arquivos copiados, obteve %d", mDelta.FilesCopied)
	}
	if mDelta.FilesSkipped != 2 {
		t.Fatalf("delta sync: esperado 2 arquivos pulados (delta identical), obteve %d", mDelta.FilesSkipped)
	}
	if copiedFilesCount.Load() != 3 {
		t.Errorf("delta events: esperado 3 eventos de cópia, obteve %d", copiedFilesCount.Load())
	}
	if skippedFilesCount.Load() != 2 {
		t.Errorf("delta events: esperado 2 eventos de skip, obteve %d", skippedFilesCount.Load())
	}

	// Validação de integridade no destino pós-delta
	expectedFiles := []string{
		"doc_unchanged_1.txt",
		"doc_unchanged_2.txt",
		"doc_to_modify_content.txt",
		"doc_to_touch_time.txt",
		"doc_brand_new.txt",
	}

	for _, fname := range expectedFiles {
		srcPath := filepath.Join(srcDir, fname)
		dstPath := filepath.Join(destDir, fname)

		srcHash, err := engine.ComputeFileHashXX64(srcPath)
		if err != nil {
			t.Fatalf("falha ao calcular hash de origem para %s: %v", fname, err)
		}
		dstHash, err := engine.ComputeFileHashXX64(dstPath)
		if err != nil {
			t.Fatalf("falha ao calcular hash de destino para %s: %v", fname, err)
		}
		if srcHash != dstHash {
			t.Errorf("[%s] divergência de hash após Delta Sync: src=%s dst=%s", fname, srcHash, dstHash)
		}
	}
}

// ---------------------------------------------------------------------------
// 3. Integridade Inline xxHash64: Validação Rigorosa e Detecção de Corrupção
// ---------------------------------------------------------------------------
func TestIntegration_InlineIntegrity_xxHash64(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "integrity_src.bin")
	dstFile := filepath.Join(tempDir, "integrity_dst.bin")

	// Gera 1MB com conteúdo pseudo-aleatório
	payload := make([]byte, 1024*1024)
	if _, err := rand.Read(payload); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcFile, payload, 0644); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(srcFile)
	if err != nil {
		t.Fatal(err)
	}

	// Executa cópia via CopyFileStream
	limiter := ratelimit.NewLimiter(0, 0)
	res := engine.CopyFileStream(context.Background(), srcFile, dstFile, fi, limiter)

	if res.Error != nil {
		t.Fatalf("cópia válida falhou com erro: %v", res.Error)
	}

	if res.SourceHash == "" || res.DestHash == "" {
		t.Fatal("hashes gerados não podem ser vazios")
	}

	if res.SourceHash != res.DestHash {
		t.Fatalf("discrepância inline detectada: src=%s dst=%s", res.SourceHash, res.DestHash)
	}

	// Validação com cálculo independente
	expectedHash, err := engine.ComputeFileHashXX64(srcFile)
	if err != nil {
		t.Fatal(err)
	}
	if res.SourceHash != expectedHash {
		t.Fatalf("hash retornado (%s) diverge do cálculo independente (%s)", res.SourceHash, expectedHash)
	}

	// Valida que o arquivo foi gravado corretamente no destino
	dstData, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(payload, dstData) {
		t.Fatal("conteúdo gravado no destino diverge byte a byte da origem")
	}
}

// ---------------------------------------------------------------------------
// 4. Token Bucket Rate Limiter: Teste de Vazão e Respeito aos Limites
// ---------------------------------------------------------------------------
func TestIntegration_RateLimiter_ThroughputControl(t *testing.T) {
	// Teste 1: Validação de controle de largura de banda
	// Limite: 2 MB/s. Burst: 2 MB (mínimo).
	// Transferindo 4 MB:
	// O primeiro consumo utiliza os tokens disponíveis de burst (~2MB).
	// Os 2MB restantes exigem ~1.0s para reposição de tokens.
	limiter := ratelimit.NewLimiter(2.0, 0)
	ctx := context.Background()

	start := time.Now()
	// Transfere 4MB em blocos de 1MB
	blockSize := int64(1024 * 1024)
	for i := 0; i < 4; i++ {
		if err := limiter.Wait(ctx, blockSize, 1); err != nil {
			t.Fatalf("erro em limiter.Wait: %v", err)
		}
	}
	elapsed := time.Since(start)

	// O tempo mínimo deve ser de pelo menos 0.8s devido ao déficit de 2MB a 2MB/s
	if elapsed < 800*time.Millisecond {
		t.Errorf("rate limiter não restringiu vazão: 4MB a 2MB/s levou apenas %v", elapsed)
	}

	// Teste 2: Validação de controle de IOPS
	// Limite: 20 IOPS. Burst: 20 ops.
	// Executa 40 operações de IOPS:
	// As primeiras 20 consomem o burst.
	// As 20 seguintes exigem 1.0s de espera.
	iopsLimiter := ratelimit.NewLimiter(0, 20)
	startIOPS := time.Now()
	for i := 0; i < 40; i++ {
		if err := iopsLimiter.Wait(ctx, 0, 1); err != nil {
			t.Fatalf("erro em iopsLimiter.Wait: %v", err)
		}
	}
	elapsedIOPS := time.Since(startIOPS)

	if elapsedIOPS < 800*time.Millisecond {
		t.Errorf("rate limiter de IOPS não restringiu vazão: 40 ops a 20 IOPS levou apenas %v", elapsedIOPS)
	}

	// Teste 3: Cancelamento de contexto imediato no Wait
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Já cancelado
	err := limiter.Wait(cancelCtx, blockSize, 1)
	if err == nil {
		t.Errorf("esperado erro de contexto cancelado, obteve nil")
	}

	// Teste 4: Ajuste dinâmico de limites em runtime (Hot-reloading)
	limiter.UpdateLimits(100.0, 1000)
	bw, iops := limiter.GetLimits()
	if bw != 100.0 || iops != 1000 {
		t.Errorf("limites após UpdateLimits inválidos: bw=%f, iops=%d", bw, iops)
	}
}

// ---------------------------------------------------------------------------
// 5. Auditoria: Validação Estrutural e Coerência de audit_events.jsonl e audit_report.csv
// ---------------------------------------------------------------------------
func TestIntegration_AuditLogs_GenerationAndCoherence(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "audit_src")
	destDir := filepath.Join(tempDir, "audit_dst")
	auditDir := filepath.Join(tempDir, "audit_out")

	// Cria 5 arquivos de tamanhos variados
	for i := 1; i <= 5; i++ {
		fname := fmt.Sprintf("audit_test_file_%d.txt", i)
		createTestFile(t, filepath.Join(srcDir, fname), int64(i*1024), time.Now())
	}

	jobID := "audit" // Nomeia job como 'audit' para gerar audit_events.jsonl e audit_report.csv
	eng := engine.NewEngine()

	cfg := engine.JobConfig{
		JobID:          jobID,
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

	m := waitForJobCompletion(t, eng, 5*time.Second)
	if m.Status != engine.StateCompleted {
		t.Fatalf("job falhou: %s", m.Status)
	}

	// 1. Validação de audit_events.jsonl
	jsonlPath := filepath.Join(auditDir, "audit_events.jsonl")
	jsonlFile, err := os.Open(jsonlPath)
	if err != nil {
		t.Fatalf("arquivo audit_events.jsonl não foi gerado em %s: %v", auditDir, err)
	}
	defer jsonlFile.Close()

	var events []audit.AuditEvent
	scanner := bufio.NewScanner(jsonlFile)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		var evt audit.AuditEvent
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Fatalf("linha %d de audit_events.jsonl é JSON inválido: %v", lineNum, err)
		}

		// Coerência dos campos
		if evt.JobID != jobID {
			t.Errorf("linha %d: JobID esperado '%s', obteve '%s'", lineNum, jobID, evt.JobID)
		}
		if evt.EventID == "" {
			t.Errorf("linha %d: EventID vazio", lineNum)
		}
		if evt.Status != audit.StatusSuccess {
			t.Errorf("linha %d: Status esperado SUCCESS, obteve %s", lineNum, evt.Status)
		}
		if evt.Action != audit.ActionCopied {
			t.Errorf("linha %d: Action esperada COPIED, obteve %s", lineNum, evt.Action)
		}
		if evt.SourceHashXX64 == "" || evt.DestHashXX64 == "" {
			t.Errorf("linha %d: Hashes vazios (src=%s, dst=%s)", lineNum, evt.SourceHashXX64, evt.DestHashXX64)
		}
		if evt.SourceHashXX64 != evt.DestHashXX64 {
			t.Errorf("linha %d: Hashes divergentes: src=%s, dst=%s", lineNum, evt.SourceHashXX64, evt.DestHashXX64)
		}
		if evt.SizeBytes <= 0 {
			t.Errorf("linha %d: SizeBytes inválido (%d)", lineNum, evt.SizeBytes)
		}

		events = append(events, evt)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("erro ao ler jsonl: %v", err)
	}

	if len(events) != 5 {
		t.Fatalf("esperado 5 eventos em audit_events.jsonl, obteve %d", len(events))
	}

	// 2. Validação de audit_report.csv
	csvPath := filepath.Join(auditDir, "audit_report.csv")
	csvFile, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("arquivo audit_report.csv não foi gerado em %s: %v", auditDir, err)
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("falha ao interpretar audit_report.csv: %v", err)
	}

	// Cabeçalho + 5 linhas = 6 records
	if len(records) != 6 {
		t.Fatalf("esperado 6 linhas no CSV (1 header + 5 registros), obteve %d", len(records))
	}

	expectedHeaders := []string{
		"event_id", "job_id", "timestamp", "action", "status",
		"source_path", "dest_path", "size_bytes", "duration_ms",
		"throughput_mbs", "source_hash_xx64", "dest_hash_xx64",
		"source_mtime", "dest_mtime", "error_message",
	}

	for i, h := range expectedHeaders {
		if records[0][i] != h {
			t.Errorf("cabeçalho CSV coluna %d: esperado '%s', obteve '%s'", i, h, records[0][i])
		}
	}

	// Cruzamento de dados entre JSONL e CSV
	for i, evt := range events {
		csvRow := records[i+1]
		if csvRow[0] != evt.EventID {
			t.Errorf("linha CSV %d: EventID esperado %s, obteve %s", i+1, evt.EventID, csvRow[0])
		}
		if csvRow[1] != evt.JobID {
			t.Errorf("linha CSV %d: JobID esperado %s, obteve %s", i+1, evt.JobID, csvRow[1])
		}
		if csvRow[3] != string(evt.Action) {
			t.Errorf("linha CSV %d: Action esperada %s, obteve %s", i+1, evt.Action, csvRow[3])
		}
		if csvRow[4] != string(evt.Status) {
			t.Errorf("linha CSV %d: Status esperado %s, obteve %s", i+1, evt.Status, csvRow[4])
		}
		if csvRow[10] != evt.SourceHashXX64 {
			t.Errorf("linha CSV %d: SourceHash esperado %s, obteve %s", i+1, evt.SourceHashXX64, csvRow[10])
		}
		if csvRow[11] != evt.DestHashXX64 {
			t.Errorf("linha CSV %d: DestHash esperado %s, obteve %s", i+1, evt.DestHashXX64, csvRow[11])
		}
	}
}

// ---------------------------------------------------------------------------
// 6. Concorrência: Stress Test com Múltiplos Workers e Sem Condição de Corrida
// ---------------------------------------------------------------------------
func TestIntegration_ConcurrencyStress(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "stress_src")
	destDir := filepath.Join(tempDir, "stress_dst")
	auditDir := filepath.Join(tempDir, "stress_audit")

	// Cria 48 arquivos com distribuição em múltiplas subpastas
	totalFiles := 48
	var totalExpectedBytes int64

	for i := 0; i < totalFiles; i++ {
		subFolder := fmt.Sprintf("worker_dir_%d", i%6)
		fname := fmt.Sprintf("file_%03d.bin", i)
		size := int64((i + 1) * 2048) // entre 2KB e 98KB
		totalExpectedBytes += size
		createTestFile(t, filepath.Join(srcDir, subFolder, fname), size, time.Now())
	}

	eng := engine.NewEngine()

	// Simula múltiplos observadores concorrentes de telemetria
	var progressCallCount atomic.Int64
	eng.SetProgressCallback(func(metrics engine.TelemetryMetrics) {
		progressCallCount.Add(1)
		// Lê campos sob concorrência
		_ = metrics.CurrentThroughputMBs
		_ = metrics.CurrentIOPS
		_ = metrics.BytesTransferred
	})

	var eventCallCount atomic.Int64
	eng.SetEventCallback(func(evt audit.AuditEvent) {
		eventCallCount.Add(1)
	})

	cfg := engine.JobConfig{
		JobID:          "stress-test-job",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull,
		},
		Concurrency: 12, // 12 workers paralelos
		AuditDir:    auditDir,
	}

	if err := eng.StartJob(cfg); err != nil {
		t.Fatalf("falha ao iniciar job de stress: %v", err)
	}

	metrics := waitForJobCompletion(t, eng, 15*time.Second)

	if metrics.Status != engine.StateCompleted {
		t.Fatalf("stress test: esperado COMPLETED, obteve %s", metrics.Status)
	}
	if metrics.FilesCopied != int64(totalFiles) {
		t.Fatalf("stress test: esperado %d arquivos copiados, obteve %d", totalFiles, metrics.FilesCopied)
	}
	if metrics.FilesFailed != 0 {
		t.Fatalf("stress test: esperado 0 falhas, obteve %d", metrics.FilesFailed)
	}
	if metrics.BytesTransferred != totalExpectedBytes {
		t.Fatalf("stress test: esperado %d bytes transferidos, obteve %d", totalExpectedBytes, metrics.BytesTransferred)
	}
	if eventCallCount.Load() != int64(totalFiles) {
		t.Errorf("stress test: callbacks de eventos esperados %d, obteve %d", totalFiles, eventCallCount.Load())
	}

	// Validação de que todos os arquivos existem no destino sem corrupção
	for i := 0; i < totalFiles; i++ {
		subFolder := fmt.Sprintf("worker_dir_%d", i%6)
		fname := fmt.Sprintf("file_%03d.bin", i)
		dstPath := filepath.Join(destDir, subFolder, fname)
		srcPath := filepath.Join(srcDir, subFolder, fname)

		sHash, err1 := engine.ComputeFileHashXX64(srcPath)
		dHash, err2 := engine.ComputeFileHashXX64(dstPath)
		if err1 != nil || err2 != nil {
			t.Fatalf("erro ao ler arquivo pós-stress (%s): srcErr=%v dstErr=%v", fname, err1, err2)
		}
		if sHash != dHash {
			t.Fatalf("corrupção detectada pós-stress em %s: src=%s dst=%s", fname, sHash, dHash)
		}
	}
}

// ---------------------------------------------------------------------------
// 7. Controle de Ciclo de Vida: Pause, Resume e Cancelamento Gracioso
// ---------------------------------------------------------------------------
func TestIntegration_Lifecycle_PauseResumeCancel(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "life_src")
	destDir := filepath.Join(tempDir, "life_dst")
	auditDir := filepath.Join(tempDir, "life_audit")

	// Cria arquivos suficientes com rate limit baixo para dar tempo de pausar e cancelar
	for i := 0; i < 10; i++ {
		createTestFile(t, filepath.Join(srcDir, fmt.Sprintf("file_%d.dat", i)), 1024*1024, time.Now())
	}

	eng := engine.NewEngine()

	cfg := engine.JobConfig{
		JobID:          "lifecycle-job",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull,
		},
		MaxBandwidthMB: 1.0, // Limite baixo para controle
		Concurrency:    2,
		AuditDir:       auditDir,
	}

	if err := eng.StartJob(cfg); err != nil {
		t.Fatalf("falha ao iniciar job: %v", err)
	}

	// Aguarda iniciar
	time.Sleep(100 * time.Millisecond)

	// Pausa o trabalho
	if err := eng.PauseJob(); err != nil {
		t.Fatalf("falha ao pausar job: %v", err)
	}

	m := eng.GetMetrics()
	if m.Status != engine.StatePaused {
		t.Fatalf("esperado status PAUSED, obteve %s", m.Status)
	}

	copiedDuringPause := m.FilesCopied
	time.Sleep(200 * time.Millisecond)
	mAfterSleep := eng.GetMetrics()
	if mAfterSleep.FilesCopied > copiedDuringPause+1 {
		t.Errorf("arquivos foram copiados enquanto o job estava pausado!")
	}

	// Retoma o trabalho
	if err := eng.ResumeJob(); err != nil {
		t.Fatalf("falha ao retomar job: %v", err)
	}

	mResumed := eng.GetMetrics()
	if mResumed.Status != engine.StateRunning {
		t.Fatalf("esperado status RUNNING, obteve %s", mResumed.Status)
	}

	// Cancela o trabalho
	if err := eng.StopJob(); err != nil {
		t.Fatalf("falha ao parar job: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	mFinal := eng.GetMetrics()
	if mFinal.Status != engine.StateCancelled {
		t.Fatalf("esperado status CANCELLED, obteve %s", mFinal.Status)
	}
}
