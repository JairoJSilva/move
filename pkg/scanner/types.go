package scanner

import (
	"os"
	"time"
)

// FilterMode define a modalidade de seleção e filtragem da migração.
type FilterMode string

const (
	ModeFull       FilterMode = "FULL"        // Migração completa de todos os arquivos
	ModeDelta      FilterMode = "DELTA_ONLY"  // Apenas arquivos novos ou modificados
	ModeDateRange  FilterMode = "DATE_RANGE"  // Intervalo fechado de datas
	ModeYear       FilterMode = "BY_YEAR"     // Ano específico
	ModeMonth      FilterMode = "BY_MONTH"    // Mês/Ano específico
	ModeDay        FilterMode = "BY_DAY"      // Dia específico
	ModeRetention  FilterMode = "RETENTION"   // Arquivos mais antigos que X dias
)

// FilterOptions agrupa todas as regras de filtragem aplicáveis ao scanner.
type FilterOptions struct {
	Mode FilterMode `json:"mode"`

	// Parâmetros temporais
	StartDate *time.Time `json:"start_date,omitempty"` // Usado em DATE_RANGE
	EndDate   *time.Time `json:"end_date,omitempty"`   // Usado em DATE_RANGE
	Year      int        `json:"year,omitempty"`       // Usado em BY_YEAR e BY_MONTH
	Month     int        `json:"month,omitempty"`      // 1 a 12, usado em BY_MONTH
	TargetDay *time.Time `json:"target_day,omitempty"` // Usado em BY_DAY
	OlderThanDays int    `json:"older_than_days,omitempty"` // Usado em RETENTION

	// Padrões de nome e extensões
	IncludePatterns []string `json:"include_patterns,omitempty"` // Ex: ["*.pdf", "*.docx"]
	ExcludePatterns []string `json:"exclude_patterns,omitempty"` // Ex: ["*.tmp", "~*", "node_modules/**", ".git/**"]

	// Limites de tamanho em bytes (0 = sem limite)
	MinSizeBytes int64 `json:"min_size_bytes,omitempty"`
	MaxSizeBytes int64 `json:"max_size_bytes,omitempty"`

	// Diretório de destino (usado para comparação no modo DELTA)
	DestinationDir string `json:"destination_dir,omitempty"`
}

// ScannedItem representa um arquivo ou diretório inspecionado pelo scanner.
type ScannedItem struct {
	RelPath    string      `json:"rel_path"`
	SourcePath string      `json:"source_path"`
	DestPath   string      `json:"dest_path"`
	Size       int64       `json:"size"`
	ModTime    time.Time   `json:"mod_time"`
	Mode       os.FileMode `json:"mode"`
	IsDir      bool        `json:"is_dir"`
	NeedsCopy  bool        `json:"needs_copy"`
	SkipReason string      `json:"skip_reason,omitempty"`
}

// ScanSummary contém métricas sumarizadas da varredura (ex: dry-run / preview).
type ScanSummary struct {
	TotalFilesDiscovered int64 `json:"total_files_discovered"`
	TotalDirsDiscovered  int64 `json:"total_dirs_discovered"`
	TotalBytesDiscovered int64 `json:"total_bytes_discovered"`
	EligibleFilesCount   int64 `json:"eligible_files_count"`
	EligibleBytesTotal   int64 `json:"eligible_bytes_total"`
	SkippedFilesCount    int64 `json:"skipped_files_count"`
	SkippedBytesTotal    int64 `json:"skipped_bytes_total"`
}
