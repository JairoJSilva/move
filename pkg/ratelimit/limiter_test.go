package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestLimiter_Unthrottled(t *testing.T) {
	limiter := NewLimiter(0, 0)
	ctx := context.Background()

	start := time.Now()
	err := limiter.Wait(ctx, 10*1024*1024, 100)
	if err != nil {
		t.Fatalf("esperado nil, obteve %v", err)
	}

	if time.Since(start) > 50*time.Millisecond {
		t.Fatalf("operação não limitada demorou demais: %v", time.Since(start))
	}
}

func TestLimiter_UpdateLimits(t *testing.T) {
	limiter := NewLimiter(10, 100)
	bw, iops := limiter.GetLimits()
	if bw != 10 || iops != 100 {
		t.Fatalf("esperado (10, 100), obteve (%v, %v)", bw, iops)
	}

	limiter.UpdateLimits(50, 500)
	bw, iops = limiter.GetLimits()
	if bw != 50 || iops != 500 {
		t.Fatalf("esperado (50, 500), obteve (%v, %v)", bw, iops)
	}
}
