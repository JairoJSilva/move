package engine

import (
	"time"

	"migrations-engine/pkg/scanner"
)

// JobPhase define o estágio atual de execução da migração.
type JobPhase string

const (
	PhaseBaseline JobPhase = "PHASE_1_BASELINE"
	PhaseDelta    JobPhase = "PHASE_2_DELTA"
	PhaseCutover  JobPhase = "PHASE_3_CUTOVER"
)

// Constantes de fallback seguro de produção (Modo Produção Segura).
// Aplicadas automaticamente caso não venham informadas no JobConfig,
// evitando saturação de barramento de I/O e contenção de recursos no host.
const (
	DefaultSafeBandwidthMB float64 = 50.0 // Banda padrão segura: 50 MB/s
	DefaultSafeIOPS        int64   = 300  // IOPS padrão seguro: 300 operações/s
	DefaultSafeConcurrency int     = 4    // Concorrência padrão segura: 4 workers
)

// JobState define o estado do ciclo de vida da migração.
type JobState string

const (
	StateIdle      JobState = "IDLE"
	StateRunning   JobState = "RUNNING"
	StatePaused    JobState = "PAUSED"
	StateCompleted JobState = "COMPLETED"
	StateFailed    JobState = "FAILED"
	StateCancelled JobState = "CANCELLED"
)

// JobConfig define os parâmetros de inicialização de um job de migração.
type JobConfig struct {
	JobID          string                `json:"job_id"`
	SourceDir      string                `json:"source_dir"`
	DestinationDir string                `json:"destination_dir"`
	Filters        scanner.FilterOptions `json:"filters"`
	MaxBandwidthMB float64               `json:"max_bandwidth_mb"`
	MaxIOPS        int64                 `json:"max_iops"`
	Concurrency    int                   `json:"concurrency"` // Número de workers de cópia (default 4)
	AuditDir       string                `json:"audit_dir"`   // Diretório para salvar logs e relatórios
}

// TelemetryMetrics representa o payload de telemetria emitido via WebSocket e REST.
type TelemetryMetrics struct {
	JobID                string   `json:"job_id"`
	Status               JobState `json:"status"`
	CurrentPhase         JobPhase `json:"current_phase"`
	CurrentThroughputMBs float64  `json:"current_throughput_mbs"`
	CurrentIOPS          int64    `json:"current_iops"`
	LimitBandwidthMB     float64  `json:"limit_bandwidth_mb"`
	LimitIOPS            int64    `json:"limit_iops"`
	TotalFilesDiscovered int64    `json:"total_files_discovered"`
	FilesCopied          int64    `json:"files_copied"`
	FilesSkipped         int64    `json:"files_skipped"`
	FilesFailed          int64    `json:"files_failed"`
	TotalBytesDiscovered int64    `json:"total_bytes_discovered"`
	BytesTransferred     int64    `json:"bytes_transferred"`
	ProgressPercent      float64  `json:"progress_percent"`
	ActiveWorkers        int      `json:"active_workers"`
	CurrentFile          string   `json:"current_file"`
	ElapsedSeconds       int64    `json:"elapsed_seconds"`
	ETASeconds           int64    `json:"eta_seconds"`
	StartTime            time.Time `json:"start_time,omitempty"`
}
