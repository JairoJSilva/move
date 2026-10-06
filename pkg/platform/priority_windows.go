//go:build windows

package platform

import (
	"fmt"
	"syscall"
)

// Constantes da API Win32 para controle de prioridade de processo e subsistema de E/S.
const (
	// PROCESS_MODE_BACKGROUND_BEGIN reduz a prioridade de CPU para IDLE,
	// I/O de disco para VERY LOW e a prioridade de memória de trabalho.
	PROCESS_MODE_BACKGROUND_BEGIN = 0x00100000
	// PROCESS_MODE_BACKGROUND_END restaura os modos de prioridade padrão.
	PROCESS_MODE_BACKGROUND_END   = 0x00200000

	IDLE_PRIORITY_CLASS         = 0x00000040
	BELOW_NORMAL_PRIORITY_CLASS = 0x00004000
	NORMAL_PRIORITY_CLASS       = 0x00000020
)

var (
	modkernel32           = syscall.NewLazyDLL("kernel32.dll")
	procSetPriorityClass  = modkernel32.NewProc("SetPriorityClass")
	procGetCurrentProcess = modkernel32.NewProc("GetCurrentProcess")
)

// SetProcessBackgroundMode ativa o modo de segundo plano do Windows (PROCESS_MODE_BACKGROUND_BEGIN).
// Reduz o impacto de CPU e I/O no sistema para o nível mínimo operacional.
func SetProcessBackgroundMode() error {
	hProcess, _, _ := procGetCurrentProcess.Call()
	ret, _, err := procSetPriorityClass.Call(hProcess, uintptr(PROCESS_MODE_BACKGROUND_BEGIN))
	if ret == 0 {
		// Se PROCESS_MODE_BACKGROUND_BEGIN falhar (ex: restrição de privilégio), tenta IDLE_PRIORITY_CLASS
		ret2, _, err2 := procSetPriorityClass.Call(hProcess, uintptr(IDLE_PRIORITY_CLASS))
		if ret2 == 0 {
			return fmt.Errorf("falha ao ativar modo background no Windows: %v (fallback IDLE: %v)", err, err2)
		}
	}
	return nil
}

// ResetProcessBackgroundMode finaliza o modo de segundo plano restaurando a prioridade normal.
func ResetProcessBackgroundMode() error {
	hProcess, _, _ := procGetCurrentProcess.Call()
	ret, _, err := procSetPriorityClass.Call(hProcess, uintptr(PROCESS_MODE_BACKGROUND_END))
	if ret == 0 {
		// Restaura para normal diretamente
		ret2, _, err2 := procSetPriorityClass.Call(hProcess, uintptr(NORMAL_PRIORITY_CLASS))
		if ret2 == 0 {
			return fmt.Errorf("falha ao restaurar prioridade normal no Windows: %v (fallback NORMAL: %v)", err, err2)
		}
	}
	return nil
}

// SetLowIOPriority ativa o modo de baixa prioridade / background no Windows.
func SetLowIOPriority() error {
	return SetProcessBackgroundMode()
}

// SetBackgroundPriority ativa o modo de segundo plano (PROCESS_MODE_BACKGROUND_BEGIN) no Windows.
func SetBackgroundPriority() error {
	return SetProcessBackgroundMode()
}

// ResetBackgroundPriority restaura a prioridade normal no Windows.
func ResetBackgroundPriority() error {
	return ResetProcessBackgroundMode()
}
