package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"migrations-engine/pkg/discovery"
	"migrations-engine/pkg/engine"
	"migrations-engine/pkg/scanner"
)

// Handler gerencia o processamento das requisições REST da API.
type Handler struct {
	eng *engine.Engine
	hub *WSHub
}

// NewHandler cria uma nova instância de Handler.
func NewHandler(eng *engine.Engine, hub *WSHub) *Handler {
	return &Handler{
		eng: eng,
		hub: hub,
	}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, msg string) {
	h.respondJSON(w, status, map[string]string{"error": msg})
}

// HandleGetDisks retorna os discos e volumes descobertos no host.
func (h *Handler) HandleGetDisks(w http.ResponseWriter, r *http.Request) {
	disks, err := discovery.GetDisks()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "falha ao descobrir volumes: "+err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"disks": disks,
		"count": len(disks),
	})
}

// HandleBrowse lista o diretório especificado e valida permissões de leitura/escrita.
func (h *Handler) HandleBrowse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // Limite de 1MB para payload
	var req BrowseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "corpo de requisição inválido")
		return
	}

	if strings.ContainsRune(req.Path, 0) {
		h.respondError(w, http.StatusBadRequest, "caminho contém caracteres nulos inválidos")
		return
	}

	targetPath := req.Path
	if targetPath == "" {
		targetPath = "."
	}

	targetPath = engine.NormalizePath(targetPath)
	entries, err := os.ReadDir(targetPath)
	isReadable := (err == nil)

	// Testa permissão de escrita tentando criar um arquivo efêmero oculto
	isWritable := false
	if isReadable {
		testFile := filepath.Join(targetPath, fmt.Sprintf(".hypersync_probe_%d.tmp", time.Now().UnixNano()))
		if f, err := os.Create(testFile); err == nil {
			_ = f.Close()
			_ = os.Remove(testFile)
			isWritable = true
		}
	}

	var list []BrowseEntry
	if isReadable {
		for _, e := range entries {
			info, _ := e.Info()
			size := int64(0)
			modTime := ""
			if info != nil {
				size = info.Size()
				modTime = info.ModTime().Format(time.RFC3339)
			}

			fullPath := filepath.Join(targetPath, e.Name())
			list = append(list, BrowseEntry{
				Name:       e.Name(),
				Path:       fullPath,
				IsDir:      e.IsDir(),
				Size:       size,
				ModTime:    modTime,
				IsReadable: true,
			})
		}
	}

	h.respondJSON(w, http.StatusOK, BrowseResponse{
		CurrentPath: targetPath,
		IsReadable:  isReadable,
		IsWritable:  isWritable,
		Entries:     list,
	})
}

// HandlePreview calcula estimativas de quantidade e volume de arquivos.
func (h *Handler) HandlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "parâmetros de pré-visualização inválidos")
		return
	}

	if req.SourceDir == "" {
		h.respondError(w, http.StatusBadRequest, "source_dir é obrigatório")
		return
	}

	if strings.ContainsRune(req.SourceDir, 0) || strings.ContainsRune(req.DestinationDir, 0) {
		h.respondError(w, http.StatusBadRequest, "caminhos informados contêm caracteres nulos inválidos")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	summary, err := scanner.Preview(ctx, req.SourceDir, req.DestinationDir, req.Filters)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "erro ao calcular estimativa: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, summary)
}

// HandleStartMigration inicia um novo job de migração.
func (h *Handler) HandleStartMigration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var cfg engine.JobConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		h.respondError(w, http.StatusBadRequest, "configuração de migração inválida")
		return
	}

	if cfg.SourceDir == "" || cfg.DestinationDir == "" {
		h.respondError(w, http.StatusBadRequest, "source_dir e destination_dir são obrigatórios")
		return
	}

	if strings.ContainsRune(cfg.SourceDir, 0) || strings.ContainsRune(cfg.DestinationDir, 0) {
		h.respondError(w, http.StatusBadRequest, "caminhos de migração contêm caracteres nulos inválidos")
		return
	}

	if err := h.eng.StartJob(cfg); err != nil {
		h.respondError(w, http.StatusConflict, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "STARTED",
		"job_id":  cfg.JobID,
		"metrics": h.eng.GetMetrics(),
	})
}

// HandlePauseMigration pausa a migração em andamento.
func (h *Handler) HandlePauseMigration(w http.ResponseWriter, r *http.Request) {
	if err := h.eng.PauseJob(); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "PAUSED",
		"metrics": h.eng.GetMetrics(),
	})
}

// HandleResumeMigration retoma a migração pausada.
func (h *Handler) HandleResumeMigration(w http.ResponseWriter, r *http.Request) {
	if err := h.eng.ResumeJob(); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "RUNNING",
		"metrics": h.eng.GetMetrics(),
	})
}

// HandleStopMigration interrompe o job atual.
func (h *Handler) HandleStopMigration(w http.ResponseWriter, r *http.Request) {
	if err := h.eng.StopJob(); err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "STOPPED",
		"metrics": h.eng.GetMetrics(),
	})
}

// HandleGetStatus retorna o status e métricas correntes da migração.
func (h *Handler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, h.eng.GetMetrics())
}

// HandleRateLimit ajusta limites a quente sem reiniciar o processo.
func (h *Handler) HandleRateLimit(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req RateLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "parâmetros de limite inválidos")
		return
	}

	if req.MaxBandwidthMB < 0 || req.MaxIOPS < 0 {
		h.respondError(w, http.StatusBadRequest, "valores de taxa não podem ser negativos")
		return
	}

	h.eng.UpdateLimits(req.MaxBandwidthMB, req.MaxIOPS)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":             "LIMITS_UPDATED",
		"limit_bandwidth_mb": req.MaxBandwidthMB,
		"limit_iops":         req.MaxIOPS,
	})
}

// HandleSummaryReport retorna o relatório executivo sumarizado.
func (h *Handler) HandleSummaryReport(w http.ResponseWriter, r *http.Request) {
	summary := h.eng.GetExecutiveSummary()
	if summary == nil {
		h.respondError(w, http.StatusNotFound, "nenhum relatório disponível ou migração ainda não iniciada")
		return
	}
	h.respondJSON(w, http.StatusOK, summary)
}

// HandleExportReport exporta arquivos de auditoria para download (CSV ou JSONL).
func (h *Handler) HandleExportReport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	if format != "csv" && format != "jsonl" {
		h.respondError(w, http.StatusBadRequest, "formato inválido: apenas 'csv' e 'jsonl' são suportados")
		return
	}

	metrics := h.eng.GetMetrics()
	if metrics.JobID == "" {
		h.respondError(w, http.StatusNotFound, "nenhum trabalho registrado")
		return
	}

	// Sanitiza JobID contra Path Traversal
	cleanJobID := filepath.Base(metrics.JobID)
	if cleanJobID != metrics.JobID || strings.Contains(metrics.JobID, "..") || strings.ContainsAny(metrics.JobID, "/\\") {
		h.respondError(w, http.StatusBadRequest, "identificador de trabalho inválido")
		return
	}

	var fileName string
	var contentType string
	if format == "jsonl" {
		fileName = fmt.Sprintf("%s_events.jsonl", cleanJobID)
		contentType = "application/x-ndjson"
	} else {
		fileName = fmt.Sprintf("%s_report.csv", cleanJobID)
		contentType = "text/csv; charset=utf-8"
	}

	auditPath := filepath.Join(".", "audit_logs", fileName)
	file, err := os.Open(auditPath)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "arquivo de relatório não encontrado: "+err.Error())
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	_, _ = io.Copy(w, file)
}
