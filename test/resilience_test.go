package test

import (
	"crypto/rand"
	"fmt"
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

// ---------------------------------------------------------------------------
// 1. Simulação de Interrupção no Meio da Transferência e Retomada com Delta Sync
// ---------------------------------------------------------------------------
func TestResilience_MidTransferCancellation_And_DeltaSyncResume(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "resilience_src")
	destDir := filepath.Join(tempDir, "resilience_dst")
	auditDir := filepath.Join(tempDir, "resilience_audit")

	totalFiles := 15
	fileSize := int64(1024 * 1024) // 1MB cada
	t0 := time.Now().Add(-2 * time.Hour).Truncate(time.Second)

	for i := 0; i < totalFiles; i++ {
		fname := fmt.Sprintf("dataset_chunk_%02d.dat", i)
		createTestFile(t, filepath.Join(srcDir, fname), fileSize, t0)
	}

	// 1.1 Inicia Job 1 com throttling baixo para permitir interceptação em voo
	eng1 := engine.NewEngine()

	var copiedCount1 atomic.Int64
	eng1.SetEventCallback(func(evt audit.AuditEvent) {
		if evt.Action == audit.ActionCopied {
			copiedCount1.Add(1)
		}
	})

	cfg1 := engine.JobConfig{
		JobID:          "resilience-cancel-job-01",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull,
		},
		MaxBandwidthMB: 1.5, // 1.5 MB/s controlado
		Concurrency:    2,
		AuditDir:       auditDir,
	}

	if err := eng1.StartJob(cfg1); err != nil {
		t.Fatalf("falha ao iniciar Job 1: %v", err)
	}

	// Aguarda transferir entre 2 e 5 arquivos
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		if copiedCount1.Load() >= 2 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Força cancelamento no meio da transferência
	if err := eng1.StopJob(); err != nil {
		t.Fatalf("falha ao cancelar Job 1: %v", err)
	}

	// Aguarda estado CANCELLED e drenagem dos workers
	m1 := waitForJobCompletion(t, eng1, 5*time.Second)
	if m1.Status != engine.StateCancelled {
		t.Fatalf("esperado status CANCELLED no Job 1, obteve %s", m1.Status)
	}

	initialCopied := m1.FilesCopied
	if initialCopied >= int64(totalFiles) {
		t.Fatalf("teste de resiliência inválido: todos os arquivos foram copiados antes da interrupção (%d/%d)", initialCopied, totalFiles)
	}
	if initialCopied == 0 {
		t.Fatalf("nenhum arquivo foi copiado antes da interrupção")
	}

	t.Logf("Job 1 interrompido com sucesso: %d/%d arquivos copiados", initialCopied, totalFiles)

	// 1.2 Valida que os arquivos já copiados estão íntegros e que nenhum arquivo parcial corrompido ficou no destino
	for i := 0; i < totalFiles; i++ {
		fname := fmt.Sprintf("dataset_chunk_%02d.dat", i)
		dstPath := filepath.Join(destDir, fname)
		srcPath := filepath.Join(srcDir, fname)

		if fi, err := os.Stat(dstPath); err == nil {
			// Se o arquivo existe no destino, deve ter exatamente o tamanho completo e hash válido
			if fi.Size() != fileSize {
				t.Fatalf("arquivo parcial corrompido encontrado no destino após cancelamento: %s (tamanho: %d != %d)", fname, fi.Size(), fileSize)
			}
			srcH, _ := engine.ComputeFileHashXX64(srcPath)
			dstH, _ := engine.ComputeFileHashXX64(dstPath)
			if srcH != dstH {
				t.Fatalf("arquivo corrompido no destino: %s", fname)
			}
		}
	}

	// 1.3 Passo 2: Retomada (Resume) com Delta Sync
	engResume := engine.NewEngine()

	var skippedResumeCount atomic.Int64
	var copiedResumeCount atomic.Int64
	engResume.SetEventCallback(func(evt audit.AuditEvent) {
		if evt.Action == audit.ActionSkipped {
			skippedResumeCount.Add(1)
		} else if evt.Action == audit.ActionCopied {
			copiedResumeCount.Add(1)
		}
	})

	cfgResume := engine.JobConfig{
		JobID:          "resilience-resume-job-02",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeDelta,
		},
		Concurrency: 4,
		AuditDir:    auditDir,
	}

	if err := engResume.StartJob(cfgResume); err != nil {
		t.Fatalf("falha ao iniciar job de retomada: %v", err)
	}

	mResume := waitForJobCompletion(t, engResume, 10*time.Second)

	if mResume.Status != engine.StateCompleted {
		t.Fatalf("retomada falhou com status: %s", mResume.Status)
	}

	// Validação das métricas de delta sync na retomada
	expectedSkipped := initialCopied
	expectedCopied := int64(totalFiles) - initialCopied

	if mResume.FilesSkipped != expectedSkipped {
		t.Errorf("delta sync resume: esperado %d arquivos pulados, obteve %d", expectedSkipped, mResume.FilesSkipped)
	}
	if mResume.FilesCopied != expectedCopied {
		t.Errorf("delta sync resume: esperado %d arquivos copiados, obteve %d", expectedCopied, mResume.FilesCopied)
	}
	if mResume.FilesFailed != 0 {
		t.Errorf("delta sync resume: esperado 0 falhas, obteve %d", mResume.FilesFailed)
	}

	// 1.4 Validação Criptográfica Global: 100% dos arquivos no destino com integridade xxHash64
	for i := 0; i < totalFiles; i++ {
		fname := fmt.Sprintf("dataset_chunk_%02d.dat", i)
		dstPath := filepath.Join(destDir, fname)
		srcPath := filepath.Join(srcDir, fname)

		srcHash, err1 := engine.ComputeFileHashXX64(srcPath)
		dstHash, err2 := engine.ComputeFileHashXX64(dstPath)
		if err1 != nil || err2 != nil {
			t.Fatalf("falha ao ler arquivo pós-retomada (%s): %v / %v", fname, err1, err2)
		}
		if srcHash != dstHash {
			t.Fatalf("corrupção de dados pós-retomada em %s: src=%s dst=%s", fname, srcHash, dstHash)
		}
	}

	t.Logf("Retomada concluída: %d arquivos verificados com 100%% de integridade xxHash64", totalFiles)
}

// ---------------------------------------------------------------------------
// 2. Simulação de Crash Abrupto com Arquivo Parcial Órfão Truncado no Destino
// ---------------------------------------------------------------------------
func TestResilience_CrashRecoveryWithOrphanPartialFile_DeltaSync(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "crash_src")
	destDir := filepath.Join(tempDir, "crash_dst")
	auditDir := filepath.Join(tempDir, "crash_audit")

	t0 := time.Now().Add(-1 * time.Hour).Truncate(time.Second)

	// Cria 6 arquivos na origem
	type testItem struct {
		name string
		size int64
	}
	items := []testItem{
		{"valid_01.dat", 1024 * 1024},
		{"valid_02.dat", 512 * 1024},
		{"orphan_partial.dat", 2 * 1024 * 1024}, // 2MB na origem
		{"corrupted_data.dat", 1024 * 1024},     // 1MB
		{"pending_01.dat", 256 * 1024},
		{"pending_02.dat", 768 * 1024},
	}

	for _, it := range items {
		createTestFile(t, filepath.Join(srcDir, it.name), it.size, t0)
	}

	// Simula estado do destino deixado por uma queda abrupta de energia / SIGKILL:
	// 1. valid_01 e valid_02 já haviam sido copiados com sucesso e metadados preservados
	createTestFile(t, filepath.Join(destDir, "valid_01.dat"), 1024*1024, t0)
	createTestFile(t, filepath.Join(destDir, "valid_02.dat"), 512*1024, t0)

	// 2. orphan_partial.dat foi interrompido na metade (apenas 300KB gravados em vez de 2MB)
	createTestFile(t, filepath.Join(destDir, "orphan_partial.dat"), 300*1024, t0.Add(-10*time.Minute))

	// 3. corrupted_data.dat tem tamanho igual mas bytes e timestamp divergentes
	badBytes := make([]byte, 1024*1024)
	_, _ = rand.Read(badBytes)
	_ = os.MkdirAll(destDir, 0755)
	_ = os.WriteFile(filepath.Join(destDir, "corrupted_data.dat"), badBytes, 0644)
	_ = os.Chtimes(filepath.Join(destDir, "corrupted_data.dat"), t0.Add(5*time.Minute), t0.Add(5*time.Minute))

	// 4. pending_01 e pending_02 não existem no destino

	// Executa retomada com Delta Sync
	eng := engine.NewEngine()

	cfg := engine.JobConfig{
		JobID:          "crash-recovery-job",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeDelta,
		},
		Concurrency: 4,
		AuditDir:    auditDir,
	}

	if err := eng.StartJob(cfg); err != nil {
		t.Fatalf("falha ao iniciar job de recuperação de crash: %v", err)
	}

	metrics := waitForJobCompletion(t, eng, 10*time.Second)

	if metrics.Status != engine.StateCompleted {
		t.Fatalf("recuperação falhou com status %s", metrics.Status)
	}

	// Esperado:
	// 2 arquivos idênticos pulados (valid_01, valid_02)
	// 4 arquivos copiados/corrigidos (orphan_partial truncado, corrupted_data divergente, pending_01, pending_02)
	if metrics.FilesSkipped != 2 {
		t.Errorf("esperado 2 arquivos pulados, obteve %d", metrics.FilesSkipped)
	}
	if metrics.FilesCopied != 4 {
		t.Errorf("esperado 4 arquivos copiados/recuperados, obteve %d", metrics.FilesCopied)
	}
	if metrics.FilesFailed != 0 {
		t.Errorf("esperado 0 falhas, obteve %d", metrics.FilesFailed)
	}

	// Validação final de integridade e xxHash64 de TODOS os 6 arquivos
	for _, it := range items {
		srcP := filepath.Join(srcDir, it.name)
		dstP := filepath.Join(destDir, it.name)

		fi, err := os.Stat(dstP)
		if err != nil {
			t.Fatalf("arquivo não encontrado após recuperação: %s", it.name)
		}
		if fi.Size() != it.size {
			t.Fatalf("tamanho incorreto em %s: esperado %d, obteve %d", it.name, it.size, fi.Size())
		}

		sHash, err1 := engine.ComputeFileHashXX64(srcP)
		dHash, err2 := engine.ComputeFileHashXX64(dstP)
		if err1 != nil || err2 != nil {
			t.Fatalf("erro ao calcular hashes (%s): %v / %v", it.name, err1, err2)
		}
		if sHash != dHash {
			t.Fatalf("arquivo corrompido após recuperação pós-crash (%s): src=%s dst=%s", it.name, sHash, dHash)
		}
	}
}

// ---------------------------------------------------------------------------
// 3. Teste de Múltiplos Ciclos de Interrupção e Retomada (Cascading Crash/Resume)
// ---------------------------------------------------------------------------
func TestResilience_CascadingMultiCycle_InterruptionAndResume(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "cascading_src")
	destDir := filepath.Join(tempDir, "cascading_dst")
	auditDir := filepath.Join(tempDir, "cascading_audit")

	totalFiles := 20
	fileSize := int64(256 * 1024) // 256KB cada
	t0 := time.Now().Add(-5 * time.Hour).Truncate(time.Second)

	for i := 0; i < totalFiles; i++ {
		fname := fmt.Sprintf("chunk_%03d.bin", i)
		createTestFile(t, filepath.Join(srcDir, fname), fileSize, t0)
	}

	// Ciclo 1: Inicia e interrompe precocemente
	eng1 := engine.NewEngine()
	cfg1 := engine.JobConfig{
		JobID:          "cascade-cycle-01",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull,
		},
		MaxBandwidthMB: 1.0,
		Concurrency:    2,
		AuditDir:       auditDir,
	}
	if err := eng1.StartJob(cfg1); err != nil {
		t.Fatalf("ciclo 1 falhou ao iniciar: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
	_ = eng1.StopJob()
	m1 := waitForJobCompletion(t, eng1, 5*time.Second)
	t.Logf("Ciclo 1 finalizado: %d arquivos copiados", m1.FilesCopied)

	// Ciclo 2: Retoma com Delta Sync e interrompe novamente
	eng2 := engine.NewEngine()
	cfg2 := engine.JobConfig{
		JobID:          "cascade-cycle-02",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeDelta,
		},
		MaxBandwidthMB: 1.0,
		Concurrency:    2,
		AuditDir:       auditDir,
	}
	if err := eng2.StartJob(cfg2); err != nil {
		t.Fatalf("ciclo 2 falhou ao iniciar: %v", err)
	}
	time.Sleep(300 * time.Millisecond)
	_ = eng2.StopJob()
	m2 := waitForJobCompletion(t, eng2, 5*time.Second)
	t.Logf("Ciclo 2 finalizado: %d copiados, %d pulados", m2.FilesCopied, m2.FilesSkipped)

	// Ciclo 3: Retoma com Delta Sync e deixa concluir 100%
	eng3 := engine.NewEngine()
	cfg3 := engine.JobConfig{
		JobID:          "cascade-cycle-03",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeDelta,
		},
		Concurrency: 4,
		AuditDir:    auditDir,
	}
	if err := eng3.StartJob(cfg3); err != nil {
		t.Fatalf("ciclo 3 falhou ao iniciar: %v", err)
	}
	m3 := waitForJobCompletion(t, eng3, 10*time.Second)
	if m3.Status != engine.StateCompleted {
		t.Fatalf("ciclo 3 falhou com status %s", m3.Status)
	}
	t.Logf("Ciclo 3 finalizado com sucesso: %d copiados, %d pulados", m3.FilesCopied, m3.FilesSkipped)

	// Validação final de integridade de todos os 20 arquivos
	for i := 0; i < totalFiles; i++ {
		fname := fmt.Sprintf("chunk_%03d.bin", i)
		srcP := filepath.Join(srcDir, fname)
		dstP := filepath.Join(destDir, fname)

		sHash, err1 := engine.ComputeFileHashXX64(srcP)
		dHash, err2 := engine.ComputeFileHashXX64(dstP)
		if err1 != nil || err2 != nil {
			t.Fatalf("falha ao verificar arquivo pós-ciclos (%s): %v / %v", fname, err1, err2)
		}
		if sHash != dHash {
			t.Fatalf("corrupção de dados após múltiplos ciclos em %s: src=%s dst=%s", fname, sHash, dHash)
		}
	}
}

// ---------------------------------------------------------------------------
// 4. Teste de Persistência de Checkpointing em Log de Auditoria
// ---------------------------------------------------------------------------
func TestResilience_CheckpointAuditLog_PersistenceAndResume(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "ckpt_src")
	destDir := filepath.Join(tempDir, "ckpt_dst")
	auditDir := filepath.Join(tempDir, "ckpt_audit")

	jobID := "ckpt-resilience-job-01"

	// Cria 12 arquivos de 512KB na origem
	totalFiles := 12
	fileSize := int64(512 * 1024)
	for i := 0; i < totalFiles; i++ {
		createTestFile(t, filepath.Join(srcDir, fmt.Sprintf("record_%02d.dat", i)), fileSize, time.Now())
	}

	// Inicia Job 1 com limite de banda estrito para interromper em voo
	eng1 := engine.NewEngine()
	var copiedCount1 atomic.Int64
	eng1.SetEventCallback(func(evt audit.AuditEvent) {
		if evt.Action == audit.ActionCopied {
			copiedCount1.Add(1)
		}
	})

	cfg1 := engine.JobConfig{
		JobID:          jobID,
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull,
		},
		MaxBandwidthMB: 1.0,
		Concurrency:    1, // 1 worker sequencial para controle preciso de checkpoint
		AuditDir:       auditDir,
	}

	if err := eng1.StartJob(cfg1); err != nil {
		t.Fatalf("falha ao iniciar job 1: %v", err)
	}

	// Aguarda copiar entre 3 e 5 arquivos
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		if copiedCount1.Load() >= 3 {
			break
		}
		time.Sleep(30 * time.Millisecond)
	}
	_ = eng1.StopJob()
	m1 := waitForJobCompletion(t, eng1, 5*time.Second)

	filesDoneFirstPass := m1.FilesCopied
	if filesDoneFirstPass >= int64(totalFiles) {
		t.Fatalf("job 1 concluiu todos os arquivos (%d/%d) antes da parada; ajuste de tempo necessário", filesDoneFirstPass, totalFiles)
	}
	t.Logf("Job 1 completou %d/%d arquivos antes da parada", filesDoneFirstPass, totalFiles)

	// Inicia Job 2 com o MESMO JobID e auditDir (simulando reinicialização de processo)
	eng2 := engine.NewEngine()
	cfg2 := engine.JobConfig{
		JobID:          jobID, // Mesmo JobID: carrega checkpoints do JSONL existente
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull, // Mesmo em Modo Full, o checkpoint do log evita re-cópia!
		},
		Concurrency: 4,
		AuditDir:    auditDir,
	}

	if err := eng2.StartJob(cfg2); err != nil {
		t.Fatalf("falha ao retomar job com checkpoints: %v", err)
	}

	m2 := waitForJobCompletion(t, eng2, 10*time.Second)
	if m2.Status != engine.StateCompleted {
		t.Fatalf("job de retomada falhou: status=%s", m2.Status)
	}

	// Arquivos que já estavam no checkpoint do JSONL foram pulados via IsCompleted
	if m2.FilesSkipped < filesDoneFirstPass {
		t.Errorf("esperado pelo menos %d arquivos pulados via checkpoint, obteve %d", filesDoneFirstPass, m2.FilesSkipped)
	}

	// Todos os 8 arquivos devem estar íntegros no destino
	for i := 0; i < totalFiles; i++ {
		fname := fmt.Sprintf("record_%02d.dat", i)
		srcP := filepath.Join(srcDir, fname)
		dstP := filepath.Join(destDir, fname)

		sHash, err1 := engine.ComputeFileHashXX64(srcP)
		dHash, err2 := engine.ComputeFileHashXX64(dstP)
		if err1 != nil || err2 != nil {
			t.Fatalf("falha ao ler hash do arquivo (%s): %v / %v", fname, err1, err2)
		}
		if sHash != dHash {
			t.Fatalf("discrepância de hash no destino para %s: src=%s dst=%s", fname, sHash, dHash)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. Teste de Limpeza de Arquivo Parcial em Cancelamento de Streaming
// ---------------------------------------------------------------------------
func TestResilience_StreamingCancel_CleansPartialDestination(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "big_stream_src.dat")
	dstPath := filepath.Join(tempDir, "big_stream_dst.dat")

	// Cria arquivo de 4MB
	createTestFile(t, srcPath, 4*1024*1024, time.Now())

	fi, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("falha ao obter stat: %v", err)
	}
	// Cria contexto que é cancelado após breve atraso durante a cópia
	var once sync.Once
	ctxCancel, cancelFunc := contextWithManualCancel()

	// Inicia streaming com limitação para forçar cancelamento no meio
	go func() {
		time.Sleep(50 * time.Millisecond)
		once.Do(func() {
			cancelFunc()
		})
	}()

	limiter := ratelimit.NewLimiter(0.5, 0) // 0.5 MB/s -> 4MB leva ~8s
	res := engine.CopyFileStream(ctxCancel, srcPath, dstPath, fi, limiter)

	if res.Error == nil {
		t.Fatalf("streaming não deveria completar antes do cancelamento")
	}

	// O arquivo de destino parcial DEVE ter sido limpo e removido pelo cancelamento
	if _, err := os.Stat(dstPath); !os.IsNotExist(err) {
		t.Errorf("arquivo parcial não foi removido após cancelamento! Destino ainda existe: %s", dstPath)
	}
}

// Helper para contexto cancelável manual
type cancelCtxHelper struct {
	done chan struct{}
	err  error
	mu   sync.Mutex
}

func contextWithManualCancel() (cancelCtxHelperContext, func()) {
	h := &cancelCtxHelper{
		done: make(chan struct{}),
	}
	return cancelCtxHelperContext{h}, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		select {
		case <-h.done:
		default:
			h.err = fmt.Errorf("context cancelled by test")
			close(h.done)
		}
	}
}

type cancelCtxHelperContext struct {
	h *cancelCtxHelper
}

func (c cancelCtxHelperContext) Deadline() (deadline time.Time, ok bool) { return time.Time{}, false }
func (c cancelCtxHelperContext) Done() <-chan struct{}                   { return c.h.done }
func (c cancelCtxHelperContext) Err() error {
	c.h.mu.Lock()
	defer c.h.mu.Unlock()
	return c.h.err
}
func (c cancelCtxHelperContext) Value(key any) any { return nil }
