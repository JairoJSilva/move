//go:build windows

package discovery

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func discoverDisks() ([]DiskVolume, error) {
	bitmask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil, err
	}

	var volumes []DiskVolume

	for i := 0; i < 26; i++ {
		if (bitmask & (1 << i)) == 0 {
			continue
		}

		driveLetter := string(rune('A' + i))
		driveRoot := driveLetter + ":\\"
		rootPtr, err := windows.UTF16PtrFromString(driveRoot)
		if err != nil {
			continue
		}

		driveType := windows.GetDriveType(rootPtr)
		// Ignora CD-ROM ou drives sem ponto de montagem
		if driveType == windows.DRIVE_CDROM || driveType == windows.DRIVE_NO_ROOT_DIR {
			continue
		}

		var freeBytesAvailable, totalBytes, totalFreeBytes uint64
		err = windows.GetDiskFreeSpaceEx(
			rootPtr,
			&freeBytesAvailable,
			&totalBytes,
			&totalFreeBytes,
		)
		if err != nil {
			// Pode falhar se for unidade sem mídia inserida
			continue
		}

		var volNameBuf [260]uint16
		var fsNameBuf [260]uint16
		var serialNum uint32
		var maxComponentLen uint32
		var flags uint32

		err = windows.GetVolumeInformation(
			rootPtr,
			&volNameBuf[0],
			uint32(len(volNameBuf)),
			&serialNum,
			&maxComponentLen,
			&flags,
			&fsNameBuf[0],
			uint32(len(fsNameBuf)),
		)

		volName := syscall.UTF16ToString(volNameBuf[:])
		fsName := "NTFS"
		if err == nil {
			fsName = syscall.UTF16ToString(fsNameBuf[:])
		}

		if volName == "" {
			volName = "Disco Local (" + driveLetter + ":)"
		} else {
			volName = volName + " (" + driveLetter + ":)"
		}

		const FILE_READ_ONLY_VOLUME = 0x00080000
		isReadOnly := (flags & FILE_READ_ONLY_VOLUME) != 0

		usedBytes := uint64(0)
		if totalBytes >= totalFreeBytes {
			usedBytes = totalBytes - totalFreeBytes
		}

		// Heurística de tipo de unidade: drives removíveis ou tipo fixo padrão
		isRotational := false
		if driveType == windows.DRIVE_FIXED {
			// Por padrão marcamos como disco padrão
			isRotational = false
		}

		vol := DiskVolume{
			ID:           driveLetter + ":",
			MountPoint:   driveRoot,
			Label:        volName,
			FileSystem:   fsName,
			TotalBytes:   totalBytes,
			FreeBytes:    freeBytesAvailable,
			UsedBytes:    usedBytes,
			IsRotational: isRotational,
			IsReadOnly:   isReadOnly,
		}

		volumes = append(volumes, vol)
	}

	return volumes, nil
}
