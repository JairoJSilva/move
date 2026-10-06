package api

import (
	"migrations-engine/pkg/scanner"
)

// BrowseRequest representa a requisição para explorar um caminho no host.
type BrowseRequest struct {
	Path      string `json:"path"`
	Operation string `json:"operation"` // "read" ou "write"
}

// BrowseEntry representa um arquivo ou diretório retornado na navegação.
type BrowseEntry struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	IsDir      bool   `json:"is_dir"`
	Size       int64  `json:"size"`
	ModTime    string `json:"mod_time"`
	IsReadable bool   `json:"is_readable"`
	IsWritable bool   `json:"is_writable"`
}

// BrowseResponse lista o conteúdo de uma pasta e status de permissão.
type BrowseResponse struct {
	CurrentPath string        `json:"current_path"`
	IsReadable  bool          `json:"is_readable"`
	IsWritable  bool          `json:"is_writable"`
	Entries     []BrowseEntry `json:"entries"`
}

// PreviewRequest parâmetros para contagem prévia e estimativa.
type PreviewRequest struct {
	SourceDir      string                `json:"source_dir"`
	DestinationDir string                `json:"destination_dir"`
	Filters        scanner.FilterOptions `json:"filters"`
}

// RateLimitRequest parâmetros para ajuste dinâmico de vazão.
type RateLimitRequest struct {
	MaxBandwidthMB float64 `json:"max_bandwidth_mb"`
	MaxIOPS        int64   `json:"max_iops"`
}

// WSMessage encapsula eventos enviados pelo canal WebSocket.
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
