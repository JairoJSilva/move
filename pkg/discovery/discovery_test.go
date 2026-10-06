package discovery

import (
	"testing"
)

func TestGetDisks(t *testing.T) {
	disks, err := GetDisks()
	if err != nil {
		t.Fatalf("erro ao obter discos: %v", err)
	}

	if len(disks) == 0 {
		t.Logf("aviso: nenhum disco retornado no ambiente de teste")
	} else {
		t.Logf("descobertos %d discos/volumes", len(disks))
		for _, d := range disks {
			t.Logf("Disco: ID=%s, Mount=%s, FS=%s, Total=%d MB, Livre=%d MB, Rotacional=%v, ReadOnly=%v",
				d.ID, d.MountPoint, d.FileSystem, d.TotalBytes/(1024*1024), d.FreeBytes/(1024*1024), d.IsRotational, d.IsReadOnly)
		}
	}
}
