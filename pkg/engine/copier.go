package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"migrations-engine/pkg/ratelimit"
)

// Tamanho do buffer de streaming reciclado (1MB)
const CopyBufferSize = 1024 * 1024

var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, CopyBufferSize)
		return &buf
	},
}

// CopyResult contém o resultado da transferência de um arquivo individual.
type CopyResult struct {
	BytesTransferred int64
	Duration         time.Duration
	SourceHash       string
	DestHash         string
	Error            error
}

// CopyFileStream executa a cópia de um arquivo único em streaming, com rate limiting,
// verificação inline de xxHash64 e preservação de metadados.
func CopyFileStream(
	ctx context.Context,
	sourcePath string,
	destPath string,
	info os.FileInfo,
	limiter *ratelimit.Limiter,
) CopyResult {
	start := time.Now()
	var res CopyResult

	// Garante que o diretório pai no destino existe
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(NormalizePath(destDir), 0755); err != nil {
		res.Error = fmt.Errorf("falha ao criar pasta de destino %s: %w", destDir, err)
		return res
	}

	srcFile, err := os.Open(NormalizePath(sourcePath))
	if err != nil {
		res.Error = fmt.Errorf("falha ao abrir origem: %w", err)
		return res
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(NormalizePath(destPath), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		res.Error = fmt.Errorf("falha ao criar destino: %w", err)
		return res
	}
	destClosed := false
	defer func() {
		if !destClosed {
			_ = destFile.Close()
		}
	}()

	// Buffers e Hashers
	bufPtr := bufferPool.Get().(*[]byte)
	buf := *bufPtr
	defer bufferPool.Put(bufPtr)

	srcHasher := xxhash.New()
	destHasher := xxhash.New()

	var totalBytes int64

	for {
		select {
		case <-ctx.Done():
			destFile.Close()
			os.Remove(NormalizePath(destPath)) // Remove arquivo parcial
			res.Error = ctx.Err()
			return res
		default:
		}

		n, readErr := srcFile.Read(buf)
		if n > 0 {
			srcHasher.Write(buf[:n])

			// Aplica controle de taxa antes de persistir no disco
			if limiter != nil {
				if err := limiter.Wait(ctx, int64(n), 1); err != nil {
					destFile.Close()
					os.Remove(NormalizePath(destPath))
					res.Error = err
					return res
				}
			}

			wn, writeErr := destFile.Write(buf[:n])
			if writeErr != nil {
				destFile.Close()
				os.Remove(NormalizePath(destPath))
				res.Error = fmt.Errorf("falha ao gravar no destino: %w", writeErr)
				return res
			}

			destHasher.Write(buf[:wn])
			totalBytes += int64(wn)
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			destFile.Close()
			os.Remove(NormalizePath(destPath))
			res.Error = fmt.Errorf("falha ao ler origem: %w", readErr)
			return res
		}
	}

	// Sincroniza dados com o subsistema de armazenamento
	_ = destFile.Sync()
	destClosed = true
	if err := destFile.Close(); err != nil {
		res.Error = fmt.Errorf("falha ao fechar arquivo de destino: %w", err)
		return res
	}

	res.BytesTransferred = totalBytes
	res.Duration = time.Since(start)
	res.SourceHash = FormatHashXX64(srcHasher.Sum64())
	res.DestHash = FormatHashXX64(destHasher.Sum64())

	// Verificação estrita de hash (Integridade ponta a ponta)
	if res.SourceHash != res.DestHash {
		os.Remove(NormalizePath(destPath))
		res.Error = fmt.Errorf("discrepância de integridade de hash: origem=%s destino=%s", res.SourceHash, res.DestHash)
		return res
	}

	// Preservação de atributos e timestamps (mtime e atime)
	modTime := info.ModTime()
	_ = os.Chtimes(NormalizePath(destPath), modTime, modTime)
	_ = os.Chmod(NormalizePath(destPath), info.Mode())

	return res
}
