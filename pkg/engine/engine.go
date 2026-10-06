package engine

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"migrations-engine/pkg/audit"
	"migrations-engine/pkg/platform"
	"migrations-engine/pkg/ratelimit"
	"migrations-engine/pkg/scanner"
)

// Engine coordena os jobs de migração, os workers e a telemetria.
type Engine struct {
	mu           sync.RWMutex
	currentJob   *JobConfig
	limiter      *ratelimit.Limiter
	logger       *audit.Logger
	state        JobState
	phase        JobPhase
	cancelFunc   context.CancelFunc
	pauseCond    *sync.Cond
	isPaused     bool

	// Contadores atômicos para máxima performance em concorrência
	bytesTransferred atomic.Int64
	filesCopied      atomic.Int64
	filesSkipped     atomic.Int64
	filesFailed      atomic.Int64
	totalFilesDisc   atomic.Int64
	totalBytesDisc   atomic.Int64
	activeWorkers    atomic.Int32

	// Métricas dinâmicas
	startTime        time.Time
	currentFile      string
	currentThroughput float64
	currentIOPS      int64
	lastBytesSample  int64
	lastFilesSample  int64
	lastSampleTime   time.Time

	// Callbacks de evento
	onProgressUpdate func(metrics TelemetryMetrics)
	onAuditEvent     func(evt audit.AuditEvent)
}

// NewEngine cria uma nova instância do motor de migração.
func NewEngine() *Engine {
	e := &Engine{
		state:     StateIdle,
		pauseCond: sync.NewCond(&sync.Mutex{}),
	}
	return e
}

// SetProgressCallback define um listener para atualizações de telemetria periódicas.
func (e *Engine) SetProgressCallback(fn func(metrics TelemetryMetrics)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onProgressUpdate = fn
}

// SetEventCallback define um listener para notificações individuais de arquivo migrado.
func (e *Engine) SetEventCallback(fn func(evt audit.AuditEvent)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onAuditEvent = fn
}

// isValidJobID valida se o ID do job contém apenas caracteres alfanuméricos seguros.
func isValidJobID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// StartJob inicia um novo processo de migração.
func (e *Engine) StartJob(config JobConfig) error {
	e.mu.Lock()
	if e.state == StateRunning || e.state == StatePaused {
		e.mu.Unlock()
		return fmt.Errorf("um trabalho de migração já está em andamento (job: %s)", e.currentJob.JobID)
	}

	if config.JobID == "" {
		config.JobID = fmt.Sprintf("job-%d", time.Now().Unix())
	} else if !isValidJobID(config.JobID) {
		e.mu.Unlock()
		return fmt.Errorf("job_id inválido: apenas caracteres alfanuméricos, '-' e '_' são permitidos (path traversal prevenido)")
	}
	// Aplica constantes de fallback seguro de produção caso não venham informadas no JobConfig
	if config.Concurrency <= 0 {
		config.Concurrency = DefaultSafeConcurrency
	}
	if config.MaxBandwidthMB <= 0 {
		config.MaxBandwidthMB = DefaultSafeBandwidthMB
	}
	if config.MaxIOPS <= 0 {
		config.MaxIOPS = DefaultSafeIOPS
	}
	if config.AuditDir == "" {
		config.AuditDir = filepath.Join(".", "audit_logs")
	}

	logger, err := audit.NewLogger(config.JobID, config.AuditDir)
	if err != nil {
		e.mu.Unlock()
		return fmt.Errorf("falha ao inicializar logger de auditoria: %w", err)
	}

	e.currentJob = &config
	e.logger = logger
	e.limiter = ratelimit.NewLimiter(config.MaxBandwidthMB, config.MaxIOPS)
	e.state = StateRunning
	e.phase = PhaseBaseline
	e.isPaused = false
	e.startTime = time.Now()
	e.lastSampleTime = e.startTime

	// Zera contadores
	e.bytesTransferred.Store(0)
	e.filesCopied.Store(0)
	e.filesSkipped.Store(0)
	e.filesFailed.Store(0)
	e.totalFilesDisc.Store(0)
	e.totalBytesDisc.Store(0)
	e.activeWorkers.Store(0)
	e.currentFile = ""
	e.currentThroughput = 0
	e.currentIOPS = 0
	e.lastBytesSample = 0
	e.lastFilesSample = 0

	ctx, cancel := context.WithCancel(context.Background())
	e.cancelFunc = cancel
	e.mu.Unlock()

	// Configura prioridade de I/O em background para mitigar impacto de E/S no host
	if err := platform.SetBackgroundPriority(); err != nil {
		log.Printf("[Engine] Aviso: não foi possível definir prioridade de I/O em background: %v", err)
	} else {
		log.Printf("[Engine] Prioridade de I/O em background (IDLE/Background Mode) ativada")
	}

	// Inicia a rotina de execução em background
	go e.runPipeline(ctx, config)

	return nil
}

// PauseJob pausa a migração sem perder o estado.
func (e *Engine) PauseJob() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StateRunning {
		return fmt.Errorf("migração não está em execução (estado atual: %s)", e.state)
	}

	e.pauseCond.L.Lock()
	e.isPaused = true
	e.state = StatePaused
	e.pauseCond.L.Unlock()

	return nil
}

// ResumeJob retoma a migração pausada.
func (e *Engine) ResumeJob() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StatePaused {
		return fmt.Errorf("migração não está pausada (estado atual: %s)", e.state)
	}

	e.pauseCond.L.Lock()
	e.isPaused = false
	e.state = StateRunning
	e.pauseCond.Broadcast()
	e.pauseCond.L.Unlock()

	return nil
}

// StopJob cancela graciosamente o trabalho atual.
func (e *Engine) StopJob() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StateRunning && e.state != StatePaused {
		return nil
	}

	if e.isPaused {
		// Acorda goroutines travadas no pause para que possam sair via ctx
		e.pauseCond.L.Lock()
		e.isPaused = false
		e.pauseCond.Broadcast()
		e.pauseCond.L.Unlock()
	}

	if e.cancelFunc != nil {
		e.cancelFunc()
	}

	e.state = StateCancelled
	return nil
}

// UpdateLimits ajusta dinamicamente a taxa de transferência e IOPS em tempo de execução.
func (e *Engine) UpdateLimits(bandwidthMB float64, iops int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.currentJob != nil {
		e.currentJob.MaxBandwidthMB = bandwidthMB
		e.currentJob.MaxIOPS = iops
	}
	if e.limiter != nil {
		e.limiter.UpdateLimits(bandwidthMB, iops)
	}
}

// GetMetrics retorna o estado atual e a telemetria do motor.
func (e *Engine) GetMetrics() TelemetryMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()

	now := time.Now()
	elapsed := int64(0)
	if !e.startTime.IsZero() {
		elapsed = int64(now.Sub(e.startTime).Seconds())
	}

	totalBytes := e.totalBytesDisc.Load()
	transferred := e.bytesTransferred.Load()
	progress := 0.0
	if totalBytes > 0 {
		progress = (float64(transferred) / float64(totalBytes)) * 100.0
		if progress > 100.0 {
			progress = 100.0
		}
	}

	eta := int64(0)
	if e.currentThroughput > 0 && totalBytes > transferred {
		remainingBytes := totalBytes - transferred
		eta = int64(float64(remainingBytes) / (e.currentThroughput * 1024 * 1024))
	}

	limitBandwidth := 0.0
	limitIOPS := int64(0)
	jobID := ""
	if e.currentJob != nil {
		jobID = e.currentJob.JobID
		limitBandwidth = e.currentJob.MaxBandwidthMB
		limitIOPS = e.currentJob.MaxIOPS
	}

	return TelemetryMetrics{
		JobID:                jobID,
		Status:               e.state,
		CurrentPhase:         e.phase,
		CurrentThroughputMBs: e.currentThroughput,
		CurrentIOPS:          e.currentIOPS,
		LimitBandwidthMB:     limitBandwidth,
		LimitIOPS:            limitIOPS,
		TotalFilesDiscovered: e.totalFilesDisc.Load(),
		FilesCopied:          e.filesCopied.Load(),
		FilesSkipped:         e.filesSkipped.Load(),
		FilesFailed:          e.filesFailed.Load(),
		TotalBytesDiscovered: totalBytes,
		BytesTransferred:     transferred,
		ProgressPercent:      progress,
		ActiveWorkers:        int(e.activeWorkers.Load()),
		CurrentFile:          e.currentFile,
		ElapsedSeconds:       elapsed,
		ETASeconds:           eta,
		StartTime:            e.startTime,
	}
}

// GetExecutiveSummary consolida o relatório final do job atual.
func (e *Engine) GetExecutiveSummary() *audit.ExecutiveSummary {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.logger == nil || e.currentJob == nil {
		return nil
	}

	events := e.logger.GetEventsBuffer()
	summary := audit.GenerateSummary(e.currentJob.JobID, e.startTime, time.Now(), events)
	return &summary
}

// runPipeline gerencia o ciclo completo de varredura e cópia com workers concorrentes.
func (e *Engine) runPipeline(ctx context.Context, cfg JobConfig) {
	// Garante encerramento limpo do contexto e goroutines filhas ao finalizar o pipeline
	defer func() {
		e.mu.Lock()
		if e.cancelFunc != nil {
			e.cancelFunc()
		}
		e.mu.Unlock()
	}()

	// Goroutine de telemetria e broadcast periódico (a cada 500ms)
	telemetryTicker := time.NewTicker(500 * time.Millisecond)
	defer telemetryTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-telemetryTicker.C:
				e.updateTelemetrySample(now)
				metrics := e.GetMetrics()
				e.mu.RLock()
				callback := e.onProgressUpdate
				e.mu.RUnlock()
				if callback != nil {
					callback(metrics)
				}
			}
		}
	}()

	itemsChan := make(chan scanner.ScannedItem, cfg.Concurrency*4)

	// Inicia o Scanner concorrente
	go func() {
		_ = scanner.Scan(ctx, cfg.SourceDir, cfg.DestinationDir, cfg.Filters, itemsChan)
	}()

	var wg sync.WaitGroup
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			e.workerLoop(ctx, cfg, itemsChan)
		}(i)
	}

	// Aguarda todos os workers finalizarem
	wg.Wait()

	// Finalização
	e.mu.Lock()
	if e.state == StateRunning {
		e.state = StateCompleted
	}
	if e.logger != nil {
		_ = e.logger.Close()
	}
	e.mu.Unlock()

	// Emite última atualização de telemetria
	e.mu.RLock()
	callback := e.onProgressUpdate
	e.mu.RUnlock()
	if callback != nil {
		callback(e.GetMetrics())
	}
}

func (e *Engine) workerLoop(ctx context.Context, cfg JobConfig, itemsChan <-chan scanner.ScannedItem) {
	for {
		// Checagem de pause
		e.pauseCond.L.Lock()
		for e.isPaused {
			e.pauseCond.Wait()
		}
		e.pauseCond.L.Unlock()

		select {
		case <-ctx.Done():
			return
		case item, ok := <-itemsChan:
			if !ok {
				return
			}

			// Atualiza contadores globais descobertos
			if !item.IsDir {
				e.totalFilesDisc.Add(1)
				e.totalBytesDisc.Add(item.Size)
			}

			// Se for diretório, apenas garante criação no destino
			if item.IsDir {
				_ = os.MkdirAll(NormalizePath(item.DestPath), item.Mode)
				continue
			}

			e.activeWorkers.Add(1)
			e.mu.Lock()
			e.currentFile = item.RelPath
			e.mu.Unlock()

			// Trata item que não precisa de cópia (filtros / delta)
			if !item.NeedsCopy {
				e.filesSkipped.Add(1)
				e.activeWorkers.Add(-1)

				reason := item.SkipReason
				if reason == "" {
					reason = "filtered"
				}

				e.logAndNotify(audit.AuditEvent{
					EventID:        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
					JobID:          cfg.JobID,
					Timestamp:      time.Now(),
					Action:         audit.ActionSkipped,
					SourcePath:     item.SourcePath,
					DestPath:       item.DestPath,
					SizeBytes:      item.Size,
					SourceMTime:    item.ModTime,
					Status:         audit.StatusSkipped,
					ErrorMessage:   &reason,
				})
				continue
			}

			// Verifica checkpoint (resiliência / resume)
			if e.logger.IsCompleted(item.SourcePath) {
				e.filesSkipped.Add(1)
				e.activeWorkers.Add(-1)
				continue
			}

			// Executa cópia do arquivo em streaming
			fi, err := os.Stat(NormalizePath(item.SourcePath))
			if err != nil {
				e.filesFailed.Add(1)
				e.activeWorkers.Add(-1)
				errMsg := err.Error()
				e.logAndNotify(audit.AuditEvent{
					EventID:      fmt.Sprintf("evt-%d", time.Now().UnixNano()),
					JobID:        cfg.JobID,
					Timestamp:    time.Now(),
					Action:       audit.ActionFailed,
					SourcePath:   item.SourcePath,
					DestPath:     item.DestPath,
					SizeBytes:    item.Size,
					Status:       audit.StatusError,
					ErrorMessage: &errMsg,
				})
				continue
			}

			copyResult := CopyFileStream(ctx, item.SourcePath, item.DestPath, fi, e.limiter)

			durationMs := copyResult.Duration.Milliseconds()
			var throughputMBs float64
			if durationMs > 0 {
				throughputMBs = (float64(copyResult.BytesTransferred) / (1024 * 1024)) / (float64(durationMs) / 1000.0)
			}

			if copyResult.Error != nil {
				e.filesFailed.Add(1)
				e.activeWorkers.Add(-1)
				errMsg := copyResult.Error.Error()
				e.logAndNotify(audit.AuditEvent{
					EventID:        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
					JobID:          cfg.JobID,
					Timestamp:      time.Now(),
					Action:         audit.ActionFailed,
					SourcePath:     item.SourcePath,
					DestPath:       item.DestPath,
					SizeBytes:      item.Size,
					DurationMs:     durationMs,
					SourceMTime:    item.ModTime,
					Status:         audit.StatusError,
					ErrorMessage:   &errMsg,
				})
			} else {
				e.filesCopied.Add(1)
				e.bytesTransferred.Add(copyResult.BytesTransferred)
				e.activeWorkers.Add(-1)

				e.logAndNotify(audit.AuditEvent{
					EventID:        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
					JobID:          cfg.JobID,
					Timestamp:      time.Now(),
					Action:         audit.ActionCopied,
					SourcePath:     item.SourcePath,
					DestPath:       item.DestPath,
					SizeBytes:      copyResult.BytesTransferred,
					DurationMs:     durationMs,
					ThroughputMBs:  throughputMBs,
					SourceHashXX64: copyResult.SourceHash,
					DestHashXX64:   copyResult.DestHash,
					SourceMTime:    item.ModTime,
					DestMTime:      item.ModTime,
					Status:         audit.StatusSuccess,
				})
			}
		}
	}
}

func (e *Engine) updateTelemetrySample(now time.Time) {
	e.mu.Lock()
	defer e.mu.Unlock()

	elapsedSecs := now.Sub(e.lastSampleTime).Seconds()
	if elapsedSecs <= 0 {
		return
	}

	currentBytes := e.bytesTransferred.Load()
	currentFiles := e.filesCopied.Load()

	deltaBytes := currentBytes - e.lastBytesSample
	deltaFiles := currentFiles - e.lastFilesSample

	if deltaBytes < 0 {
		deltaBytes = 0
	}
	if deltaFiles < 0 {
		deltaFiles = 0
	}

	e.currentThroughput = (float64(deltaBytes) / (1024 * 1024)) / elapsedSecs
	e.currentIOPS = int64(float64(deltaFiles) / elapsedSecs)

	e.lastBytesSample = currentBytes
	e.lastFilesSample = currentFiles
	e.lastSampleTime = now
}

func (e *Engine) logAndNotify(evt audit.AuditEvent) {
	if e.logger != nil {
		_ = e.logger.LogEvent(evt)
	}
	e.mu.RLock()
	cb := e.onAuditEvent
	e.mu.RUnlock()
	if cb != nil {
		cb(evt)
	}
}

