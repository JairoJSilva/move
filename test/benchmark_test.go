package test

import (
	"context"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/cespare/xxhash/v2"
	"migrations-engine/pkg/engine"
	"migrations-engine/pkg/ratelimit"
)

// Benchmark xxHash64 Throughput em diferentes tamanhos de buffers (64KB, 1MB, 4MB)
func BenchmarkHash_xxHash64_64KB(b *testing.B) {
	size := 64 * 1024
	data := make([]byte, size)
	_, _ = rand.Read(data)

	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = xxhash.Sum64(data)
	}
}

func BenchmarkHash_xxHash64_1MB(b *testing.B) {
	size := 1024 * 1024
	data := make([]byte, size)
	_, _ = rand.Read(data)

	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = xxhash.Sum64(data)
	}
}

func BenchmarkHash_xxHash64_4MB(b *testing.B) {
	size := 4 * 1024 * 1024
	data := make([]byte, size)
	_, _ = rand.Read(data)

	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = xxhash.Sum64(data)
	}
}

// Benchmark Streaming File Copy com Verificação Inline de xxHash64 e Buffer Pool (1MB)
func BenchmarkStreamingCopy_1MB(b *testing.B) {
	tempDir := b.TempDir()
	srcPath := filepath.Join(tempDir, "src_bench.bin")
	dstPath := filepath.Join(tempDir, "dst_bench.bin")

	size := int64(1024 * 1024)
	data := make([]byte, size)
	_, _ = rand.Read(data)
	if err := os.WriteFile(srcPath, data, 0644); err != nil {
		b.Fatal(err)
	}

	fi, err := os.Stat(srcPath)
	if err != nil {
		b.Fatal(err)
	}

	limiter := ratelimit.NewLimiter(0, 0)
	ctx := context.Background()

	b.SetBytes(size)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res := engine.CopyFileStream(ctx, srcPath, dstPath, fi, limiter)
		if res.Error != nil {
			b.Fatalf("erro na cópia de benchmark: %v", res.Error)
		}
	}
}

// Benchmark Streaming File Copy para 8MB
func BenchmarkStreamingCopy_8MB(b *testing.B) {
	tempDir := b.TempDir()
	srcPath := filepath.Join(tempDir, "src_bench_8mb.bin")
	dstPath := filepath.Join(tempDir, "dst_bench_8mb.bin")

	size := int64(8 * 1024 * 1024)
	data := make([]byte, size)
	_, _ = rand.Read(data)
	if err := os.WriteFile(srcPath, data, 0644); err != nil {
		b.Fatal(err)
	}

	fi, err := os.Stat(srcPath)
	if err != nil {
		b.Fatal(err)
	}

	limiter := ratelimit.NewLimiter(0, 0)
	ctx := context.Background()

	b.SetBytes(size)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res := engine.CopyFileStream(ctx, srcPath, dstPath, fi, limiter)
		if res.Error != nil {
			b.Fatalf("erro na cópia de benchmark: %v", res.Error)
		}
	}
}

// Benchmark Token Bucket Limiter Wait Overhead (sem throttling)
func BenchmarkRateLimiter_UnthrottledWait(b *testing.B) {
	limiter := ratelimit.NewLimiter(0, 0)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = limiter.Wait(ctx, 1024, 1)
	}
}
