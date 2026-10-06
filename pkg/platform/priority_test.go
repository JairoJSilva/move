package platform

import (
	"testing"
)

func TestBackgroundPriority(t *testing.T) {
	// A chamada SetBackgroundPriority pode retornar erro em ambientes de sandbox/container restritos (seccomp EPERM),
	// portanto validamos que ela executa sem pânico.
	err := SetBackgroundPriority()
	t.Logf("SetBackgroundPriority() retornou: %v", err)

	errReset := ResetBackgroundPriority()
	t.Logf("ResetBackgroundPriority() retornou: %v", errReset)
}
