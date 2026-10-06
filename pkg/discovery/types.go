package discovery

// DiskVolume representa informações de um disco ou volume de armazenamento detectado.
type DiskVolume struct {
	ID           string `json:"id"`            // Ex: "/dev/sda1" ou "C:"
	MountPoint   string `json:"mount_point"`   // Ex: "/" ou "C:\"
	Label        string `json:"label"`         // Rótulo descritivo do volume
	FileSystem   string `json:"filesystem"`    // Ex: "ext4", "xfs", "NTFS", "btrfs"
	TotalBytes   uint64 `json:"total_bytes"`   // Capacidade total em bytes
	FreeBytes    uint64 `json:"free_bytes"`    // Espaço livre utilizável
	UsedBytes    uint64 `json:"used_bytes"`    // Espaço ocupado
	IsRotational bool   `json:"is_rotational"` // True para HDD mecânico, False para SSD/NVMe
	IsReadOnly   bool   `json:"is_read_only"`  // Indica se a montagem é somente-leitura
}
