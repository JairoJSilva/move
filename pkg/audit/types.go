package audit

import (
	"time"
)

// AuditAction representa a ação executada sobre o arquivo.
type AuditAction string

const (
	ActionCopied  AuditAction = "COPIED"
	ActionSkipped AuditAction = "SKIPPED"
	ActionFailed  AuditAction = "FAILED"
)

// AuditStatus representa o resultado final da operação.
type AuditStatus string

const (
	StatusSuccess AuditStatus = "SUCCESS"
	StatusSkipped AuditStatus = "SKIPPED"
	StatusError   AuditStatus = "ERROR"
)

// AuditEvent define o registro estruturado de cada arquivo processado.
type AuditEvent struct {
	EventID        string      `json:"event_id"`
	JobID          string      `json:"job_id"`
	Timestamp      time.Time   `json:"timestamp"`
	Action         AuditAction `json:"action"`
	SourcePath     string      `json:"source_path"`
	DestPath       string      `json:"dest_path"`
	SizeBytes      int64       `json:"size_bytes"`
	DurationMs     int64       `json:"duration_ms"`
	ThroughputMBs  float64     `json:"throughput_mbs"`
	SourceHashXX64 string      `json:"source_hash_xx64"`
	DestHashXX64   string      `json:"dest_hash_xx64"`
	SourceMTime    time.Time   `json:"source_mtime"`
	DestMTime      time.Time   `json:"dest_mtime"`
	Retries        int         `json:"retries"`
	Status         AuditStatus `json:"status"`
	ErrorMessage   *string     `json:"error_message"`
}

// ExecutiveSummary sumariza as métricas globais para relatórios executivos.
type ExecutiveSummary struct {
	JobID                 string        `json:"job_id"`
	StartTime             time.Time     `json:"start_time"`
	EndTime               time.Time     `json:"end_time"`
	DurationSeconds       float64       `json:"duration_seconds"`
	TotalFilesExamined    int64         `json:"total_files_examined"`
	TotalFilesCopied      int64         `json:"total_files_copied"`
	TotalFilesSkipped     int64         `json:"total_files_skipped"`
	TotalFilesFailed      int64         `json:"total_files_failed"`
	TotalBytesTransferred int64         `json:"total_bytes_transferred"`
	AverageThroughputMBs  float64       `json:"average_throughput_mbs"`
	PeakThroughputMBs     float64       `json:"peak_throughput_mbs"`
	AverageIOPS           float64       `json:"average_iops"`
	Errors                []AuditEvent  `json:"errors,omitempty"`
}
