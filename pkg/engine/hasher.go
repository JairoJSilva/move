package engine

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/cespare/xxhash/v2"
)

var hashBufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 256*1024)
		return &buf
	},
}

// ComputeFileHashXX64 calcula a soma de verificação xxHash64 de um arquivo do disco.
func ComputeFileHashXX64(filePath string) (string, error) {
	f, err := os.Open(NormalizePath(filePath))
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := xxhash.New()
	bufPtr := hashBufferPool.Get().(*[]byte)
	buf := *bufPtr
	defer hashBufferPool.Put(bufPtr)

	if _, err := io.CopyBuffer(hasher, f, buf); err != nil {
		return "", err
	}

	return fmt.Sprintf("%016x", hasher.Sum64()), nil
}

// FormatHashXX64 formata o valor uint64 em representação hexadecimal padronizada.
func FormatHashXX64(val uint64) string {
	return fmt.Sprintf("%016x", val)
}
