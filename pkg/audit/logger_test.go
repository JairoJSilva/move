package audit

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogger_JSONL_And_CSV(t *testing.T) {
	tempDir := t.TempDir()
	jobID := "audit-test-01"

	logger, err := NewLogger(jobID, tempDir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	errMsg := "permission denied"
	events := []AuditEvent{
		{
			EventID:        "evt-001",
			JobID:          jobID,
			Timestamp:      time.Now().Truncate(time.Millisecond),
			Action:         ActionCopied,
			SourcePath:     "/source/file1.txt",
			DestPath:       "/dest/file1.txt",
			SizeBytes:      1024,
			DurationMs:     50,
			ThroughputMBs:  20.48,
			SourceHashXX64: "abcdef0123456789",
			DestHashXX64:   "abcdef0123456789",
			SourceMTime:    time.Now().Add(-time.Hour).Truncate(time.Millisecond),
			DestMTime:      time.Now().Add(-time.Hour).Truncate(time.Millisecond),
			Status:         StatusSuccess,
		},
		{
			EventID:        "evt-002",
			JobID:          jobID,
			Timestamp:      time.Now().Truncate(time.Millisecond),
			Action:         ActionSkipped,
			SourcePath:     "/source/file2.txt",
			DestPath:       "/dest/file2.txt",
			SizeBytes:      2048,
			Status:         StatusSkipped,
		},
		{
			EventID:      "evt-003",
			JobID:        jobID,
			Timestamp:    time.Now().Truncate(time.Millisecond),
			Action:       ActionFailed,
			SourcePath:   "/source/file3.txt",
			DestPath:     "/dest/file3.txt",
			SizeBytes:    4096,
			Status:       StatusError,
			ErrorMessage: &errMsg,
		},
	}

	for _, evt := range events {
		if err := logger.LogEvent(evt); err != nil {
			t.Fatalf("failed to log event: %v", err)
		}
	}

	// Verify Checkpoint logic
	if !logger.IsCompleted("/source/file1.txt") {
		t.Errorf("expected /source/file1.txt to be completed checkpoint")
	}
	if logger.IsCompleted("/source/file2.txt") {
		t.Errorf("expected /source/file2.txt not to be completed checkpoint")
	}

	// Verify Events Buffer
	buffered := logger.GetEventsBuffer()
	if len(buffered) != 3 {
		t.Fatalf("expected 3 buffered events, got %d", len(buffered))
	}

	if err := logger.Close(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	// Check JSONL file
	jsonlPath := filepath.Join(tempDir, jobID+"_events.jsonl")
	jsonlData, err := os.ReadFile(jsonlPath)
	if err != nil {
		t.Fatalf("failed to read JSONL: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(jsonlData)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines in JSONL, got %d", len(lines))
	}

	for i, line := range lines {
		var decoded AuditEvent
		if err := json.Unmarshal([]byte(line), &decoded); err != nil {
			t.Fatalf("line %d is not valid JSON: %v", i, err)
		}
		if decoded.EventID != events[i].EventID {
			t.Errorf("expected EventID %s, got %s", events[i].EventID, decoded.EventID)
		}
		if decoded.Status != events[i].Status {
			t.Errorf("expected Status %s, got %s", events[i].Status, decoded.Status)
		}
	}

	// Check CSV file
	csvPath := filepath.Join(tempDir, jobID+"_report.csv")
	csvFile, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("failed to open CSV: %v", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// 1 header + 3 rows = 4 records
	if len(records) != 4 {
		t.Fatalf("expected 4 CSV records, got %d", len(records))
	}

	headers := records[0]
	if headers[0] != "event_id" || headers[1] != "job_id" || headers[4] != "status" {
		t.Errorf("unexpected headers: %v", headers)
	}

	// Verify row 1
	row1 := records[1]
	if row1[0] != "evt-001" || row1[4] != "SUCCESS" || row1[10] != "abcdef0123456789" {
		t.Errorf("unexpected row1 content: %v", row1)
	}

	// Verify row 3 (error message)
	row3 := records[3]
	if row3[0] != "evt-003" || row3[4] != "ERROR" || row3[14] != errMsg {
		t.Errorf("unexpected row3 error message: %v", row3)
	}
}

func TestExecutiveSummary_Calculation(t *testing.T) {
	start := time.Now().Add(-10 * time.Second)
	end := time.Now()

	errMsg := "io timeout"
	events := []AuditEvent{
		{
			Action:        ActionCopied,
			Status:        StatusSuccess,
			SizeBytes:     10 * 1024 * 1024, // 10MB
			ThroughputMBs: 50.0,
		},
		{
			Action:        ActionCopied,
			Status:        StatusSuccess,
			SizeBytes:     20 * 1024 * 1024, // 20MB
			ThroughputMBs: 100.0,
		},
		{
			Action: ActionSkipped,
			Status: StatusSkipped,
		},
		{
			Action:       ActionFailed,
			Status:       StatusError,
			ErrorMessage: &errMsg,
		},
	}

	summary := GenerateSummary("summary-job-01", start, end, events)

	if summary.JobID != "summary-job-01" {
		t.Errorf("expected job ID summary-job-01, got %s", summary.JobID)
	}
	if summary.TotalFilesExamined != 4 {
		t.Errorf("expected 4 files examined, got %d", summary.TotalFilesExamined)
	}
	if summary.TotalFilesCopied != 2 {
		t.Errorf("expected 2 files copied, got %d", summary.TotalFilesCopied)
	}
	if summary.TotalFilesSkipped != 1 {
		t.Errorf("expected 1 file skipped, got %d", summary.TotalFilesSkipped)
	}
	if summary.TotalFilesFailed != 1 {
		t.Errorf("expected 1 file failed, got %d", summary.TotalFilesFailed)
	}
	if summary.TotalBytesTransferred != 30*1024*1024 {
		t.Errorf("expected 30MB transferred, got %d", summary.TotalBytesTransferred)
	}
	if summary.PeakThroughputMBs != 100.0 {
		t.Errorf("expected peak throughput 100.0, got %f", summary.PeakThroughputMBs)
	}
	if len(summary.Errors) != 1 {
		t.Errorf("expected 1 error logged, got %d", len(summary.Errors))
	}
}

func TestLogger_PathTraversalRejection(t *testing.T) {
	tempDir := t.TempDir()

	maliciousJobIDs := []string{
		"../../etc/passwd",
		"..\\windows\\system32",
		"job/with/slash",
		"job\\with\\backslash",
		"",
		"job\x00nullbyte",
	}

	for _, badID := range maliciousJobIDs {
		logger, err := NewLogger(badID, tempDir)
		if err == nil {
			_ = logger.Close()
			t.Errorf("NewLogger com jobID malicioso %q deveria ter falhado com erro de validação", badID)
		}
	}
}
