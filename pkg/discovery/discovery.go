package discovery

// GetDisks descobre e retorna a lista de todos os discos e volumes montados no host.
func GetDisks() ([]DiskVolume, error) {
	return discoverDisks()
}
