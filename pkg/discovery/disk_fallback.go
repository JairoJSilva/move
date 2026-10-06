//go:build !linux && !windows

package discovery

import (
	"os"
)

func discoverDisks() ([]DiskVolume, error) {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	return []DiskVolume{
		{
			ID:           "default",
			MountPoint:   wd,
			Label:        "Diretório Atual (" + wd + ")",
			FileSystem:   "generic",
			TotalBytes:   1000 * 1024 * 1024 * 1024, // 1TB simulado
			FreeBytes:    500 * 1024 * 1024 * 1024,
			UsedBytes:    500 * 1024 * 1024 * 1024,
			IsRotational: false,
			IsReadOnly:   false,
		},
	}, nil
}
