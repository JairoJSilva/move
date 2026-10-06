//go:build !linux && !windows

package platform

// SetLowIOPriority implementação fallback no-op para outros sistemas operacionais.
func SetLowIOPriority() error {
	return nil
}

// SetBackgroundPriority implementação fallback no-op para outros sistemas operacionais.
func SetBackgroundPriority() error {
	return nil
}

// ResetBackgroundPriority implementação fallback no-op para outros sistemas operacionais.
func ResetBackgroundPriority() error {
	return nil
}
