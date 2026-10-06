package audit

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Logger gerencia a escrita síncrona/estruturada de eventos de auditoria em JSONL e CSV.
type Logger struct {
	mu           sync.Mutex
	jobID        string
	outputDir    string
	jsonlFile    *os.File
	csvFile      *os.File
	csvWriter    *csv.Writer
	eventsBuffer []AuditEvent
	checkpoints  map[string]bool // SourcePath -> true se completado com sucesso
}

// NewLogger inicializa os arquivos de auditoria em outputDir para o job especificado.
func NewLogger(jobID, outputDir string) (*Logger, error) {
	if jobID == "" || strings.ContainsAny(jobID, "/\\") || strings.Contains(jobID, "..") || strings.ContainsRune(jobID, 0) {
		return nil, fmt.Errorf("job_id inválido para auditoria: tentativa de path traversal detectada")
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("falha ao criar pasta de auditoria: %w", err)
	}

	jsonlPath := filepath.Join(outputDir, fmt.Sprintf("%s_events.jsonl", jobID))

	// Carrega checkpoints pré-existentes caso o arquivo JSONL já exista (suporte a Resume pós-falha)
	checkpoints := make(map[string]bool)
	if data, err := os.ReadFile(jsonlPath); err == nil && len(data) > 0 {
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			var evt AuditEvent
			if err := json.Unmarshal(line, &evt); err == nil {
				if evt.Status == StatusSuccess && evt.SourcePath != "" {
					checkpoints[evt.SourcePath] = true
				}
			}
		}
	}

	jsonlFile, err := os.OpenFile(jsonlPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir arquivo jsonl: %w", err)
	}

	csvPath := filepath.Join(outputDir, fmt.Sprintf("%s_report.csv", jobID))
	csvExists := false
	if fi, err := os.Stat(csvPath); err == nil && fi.Size() > 0 {
		csvExists = true
	}

	csvFile, err := os.OpenFile(csvPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		jsonlFile.Close()
		return nil, fmt.Errorf("falha ao abrir arquivo csv: %w", err)
	}

	csvWriter := csv.NewWriter(csvFile)

	// Escreve cabeçalho CSV caso seja arquivo recém-criado
	if !csvExists {
		headers := []string{
			"event_id", "job_id", "timestamp", "action", "status",
			"source_path", "dest_path", "size_bytes", "duration_ms",
			"throughput_mbs", "source_hash_xx64", "dest_hash_xx64",
			"source_mtime", "dest_mtime", "error_message",
		}
		if err := csvWriter.Write(headers); err != nil {
			jsonlFile.Close()
			csvFile.Close()
			return nil, err
		}
		csvWriter.Flush()
	}

	l := &Logger{
		jobID:        jobID,
		outputDir:    outputDir,
		jsonlFile:    jsonlFile,
		csvFile:      csvFile,
		csvWriter:    csvWriter,
		checkpoints:  checkpoints,
		eventsBuffer: make([]AuditEvent, 0, 100),
	}

	return l, nil
}

// LogEvent grava um evento de auditoria nos formatos JSON Lines e CSV.
func (l *Logger) LogEvent(evt AuditEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 1. Gravar em JSONL
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	if _, err := l.jsonlFile.Write(append(data, '\n')); err != nil {
		return err
	}

	// 2. Gravar em CSV
	errMsg := ""
	if evt.ErrorMessage != nil {
		errMsg = *evt.ErrorMessage
	}

	record := []string{
		evt.EventID,
		evt.JobID,
		evt.Timestamp.Format(time.RFC3339),
		string(evt.Action),
		string(evt.Status),
		evt.SourcePath,
		evt.DestPath,
		strconv.FormatInt(evt.SizeBytes, 10),
		strconv.FormatInt(evt.DurationMs, 10),
		fmt.Sprintf("%.2f", evt.ThroughputMBs),
		evt.SourceHashXX64,
		evt.DestHashXX64,
		evt.SourceMTime.Format(time.RFC3339),
		evt.DestMTime.Format(time.RFC3339),
		errMsg,
	}

	if err := l.csvWriter.Write(record); err != nil {
		return err
	}
	l.csvWriter.Flush()

	// 3. Atualizar checkpoints se completado com sucesso
	if evt.Status == StatusSuccess {
		l.checkpoints[evt.SourcePath] = true
	}

	l.eventsBuffer = append(l.eventsBuffer, evt)
	return nil
}

// IsCompleted verifica se o arquivo já foi transferido com sucesso em execução anterior (Checkpointing).
func (l *Logger) IsCompleted(sourcePath string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.checkpoints[sourcePath]
}

// GetEventsBuffer retorna cópia dos eventos registrados em memória.
func (l *Logger) GetEventsBuffer() []AuditEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	copied := make([]AuditEvent, len(l.eventsBuffer))
	copy(copied, l.eventsBuffer)
	return copied
}

// Close fecha os descritores de arquivos abertos.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.csvWriter != nil {
		l.csvWriter.Flush()
	}
	var err1, err2 error
	if l.jsonlFile != nil {
		err1 = l.jsonlFile.Close()
	}
	if l.csvFile != nil {
		err2 = l.csvFile.Close()
	}

	if err1 != nil {
		return err1
	}
	return err2
}
