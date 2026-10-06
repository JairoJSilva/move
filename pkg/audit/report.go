package audit

import (
	"time"
)

// GenerateSummary calcula as métricas agregadas a partir dos eventos gravados no job.
func GenerateSummary(jobID string, startTime, endTime time.Time, events []AuditEvent) ExecutiveSummary {
	var summary ExecutiveSummary
	summary.JobID = jobID
	summary.StartTime = startTime
	summary.EndTime = endTime

	if endTime.IsZero() {
		endTime = time.Now()
		summary.EndTime = endTime
	}

	summary.DurationSeconds = endTime.Sub(startTime).Seconds()
	if summary.DurationSeconds <= 0 {
		summary.DurationSeconds = 1.0
	}

	var peakMBs float64
	var totalThroughputSum float64
	var activeThroughputEvents int64

	for _, evt := range events {
		summary.TotalFilesExamined++

		switch evt.Action {
		case ActionCopied:
			summary.TotalFilesCopied++
			summary.TotalBytesTransferred += evt.SizeBytes
			if evt.ThroughputMBs > peakMBs {
				peakMBs = evt.ThroughputMBs
			}
			totalThroughputSum += evt.ThroughputMBs
			activeThroughputEvents++
		case ActionSkipped:
			summary.TotalFilesSkipped++
		case ActionFailed:
			summary.TotalFilesFailed++
			summary.Errors = append(summary.Errors, evt)
		}
	}

	summary.PeakThroughputMBs = peakMBs

	// Throughput médio real = Bytes transferidos / Segundos
	summary.AverageThroughputMBs = (float64(summary.TotalBytesTransferred) / (1024 * 1024)) / summary.DurationSeconds
	if summary.AverageThroughputMBs < 0 {
		summary.AverageThroughputMBs = 0
	}

	summary.AverageIOPS = float64(summary.TotalFilesExamined) / summary.DurationSeconds

	return summary
}
