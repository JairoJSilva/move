//go:build linux

package discovery

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sys/unix"
)

var ignoredFs = map[string]bool{
	"proc":        true,
	"sysfs":       true,
	"devtmpfs":    true,
	"devpts":      true,
	"cgroup":      true,
	"cgroup2":     true,
	"pstore":      true,
	"bpf":         true,
	"autofs":      true,
	"mqueue":      true,
	"debugfs":     true,
	"tracefs":     true,
	"hugetlbfs":   true,
	"configfs":    true,
	"fusectl":     true,
	"efivarfs":    true,
	"binfmt_misc": true,
	"nsfs":        true,
	"securityfs":  true,
	"rpc_pipefs":  true,
	"none":        true,
	"pipefs":      true,
	"sockfs":      true,
}

var partitionRegex = regexp.MustCompile(`^(nvme\d+n\d+|mmcblk\d+|[a-z]+)`)

func discoverDisks() ([]DiskVolume, error) {
	// Tenta ler /proc/mounts, /proc/self/mounts ou /etc/mtab
	mountFiles := []string{"/proc/mounts", "/proc/self/mounts", "/etc/mtab"}
	var file *os.File

	for _, mf := range mountFiles {
		f, e := os.Open(mf)
		if e == nil {
			file = f
			break
		}
	}

	if file != nil {
		defer file.Close()
		volumes, parseErr := parseMountsFile(file)
		if parseErr == nil && len(volumes) > 0 {
			return volumes, nil
		}
	}

	// Fallback inteligente para ambientes confinados (Snap, Flatpak, Docker não-root)
	return discoverFallbackLinux()
}

func parseMountsFile(file *os.File) ([]DiskVolume, error) {
	scanner := bufio.NewScanner(file)
	seenMounts := make(map[string]bool)
	var volumes []DiskVolume

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]
		mountOptions := fields[3]

		if ignoredFs[fsType] {
			continue
		}

		if seenMounts[mountPoint] {
			continue
		}

		var stat unix.Statfs_t
		if err := unix.Statfs(mountPoint, &stat); err != nil {
			continue
		}

		totalBytes := stat.Blocks * uint64(stat.Bsize)
		if totalBytes == 0 {
			continue
		}

		freeBytes := stat.Bavail * uint64(stat.Bsize)
		usedBytes := uint64(0)
		if totalBytes >= (stat.Bfree * uint64(stat.Bsize)) {
			usedBytes = totalBytes - (stat.Bfree * uint64(stat.Bsize))
		}

		isReadOnly := false
		for _, opt := range strings.Split(mountOptions, ",") {
			if opt == "ro" {
				isReadOnly = true
				break
			}
		}

		isRotational := checkRotational(device)

		label := mountPoint
		if label == "/" {
			label = "Sistema Raiz (/)"
		} else {
			label = filepath.Base(mountPoint) + " (" + mountPoint + ")"
		}

		vol := DiskVolume{
			ID:           device,
			MountPoint:   mountPoint,
			Label:        label,
			FileSystem:   fsType,
			TotalBytes:   totalBytes,
			FreeBytes:    freeBytes,
			UsedBytes:    usedBytes,
			IsRotational: isRotational,
			IsReadOnly:   isReadOnly,
		}

		seenMounts[mountPoint] = true
		volumes = append(volumes, vol)
	}

	return volumes, scanner.Err()
}

func discoverFallbackLinux() ([]DiskVolume, error) {
	candidates := []string{"/", "/home", "/mnt", "/media", "/var"}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, home)
	}

	seenFsID := make(map[uint64]bool)
	seenPath := make(map[string]bool)
	var volumes []DiskVolume

	for _, p := range candidates {
		if seenPath[p] {
			continue
		}
		fi, err := os.Stat(p)
		if err != nil || !fi.IsDir() {
			continue
		}

		var stat unix.Statfs_t
		if err := unix.Statfs(p, &stat); err != nil {
			continue
		}

		// Identificador único de filesystem
		fsID := (uint64(stat.Fsid.Val[0]) << 32) | uint64(stat.Fsid.Val[1])
		if fsID != 0 && seenFsID[fsID] {
			continue
		}

		totalBytes := stat.Blocks * uint64(stat.Bsize)
		if totalBytes == 0 {
			continue
		}

		freeBytes := stat.Bavail * uint64(stat.Bsize)
		usedBytes := uint64(0)
		if totalBytes >= (stat.Bfree * uint64(stat.Bsize)) {
			usedBytes = totalBytes - (stat.Bfree * uint64(stat.Bsize))
		}

		label := p
		if p == "/" {
			label = "Sistema Raiz (/)"
		} else {
			label = filepath.Base(p) + " (" + p + ")"
		}

		vol := DiskVolume{
			ID:           p,
			MountPoint:   p,
			Label:        label,
			FileSystem:   "posix",
			TotalBytes:   totalBytes,
			FreeBytes:    freeBytes,
			UsedBytes:    usedBytes,
			IsRotational: false,
			IsReadOnly:   false,
		}

		if fsID != 0 {
			seenFsID[fsID] = true
		}
		seenPath[p] = true
		volumes = append(volumes, vol)
	}

	return volumes, nil
}

func checkRotational(device string) bool {
	if !strings.HasPrefix(device, "/dev/") {
		return false
	}
	devName := filepath.Base(device)
	matches := partitionRegex.FindStringSubmatch(devName)
	if len(matches) > 1 {
		devName = matches[1]
	}

	rotPath := filepath.Join("/sys/block", devName, "queue", "rotational")
	data, err := os.ReadFile(rotPath)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "1"
}
