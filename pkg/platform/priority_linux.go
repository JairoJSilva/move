//go:build linux

package platform

import (
	"fmt"
	"golang.org/x/sys/unix"
)

// Constantes do subsistema de I/O priority (ioprio) do kernel Linux.
const (
	IOPRIO_WHO_PROCESS = 1
	IOPRIO_WHO_PGRP    = 2
	IOPRIO_WHO_USER    = 3

	IOPRIO_CLASS_SHIFT = 13
	IOPRIO_CLASS_NONE  = 0
	IOPRIO_CLASS_RT    = 1
	IOPRIO_CLASS_BE    = 2
	IOPRIO_CLASS_IDLE  = 3
)

// IOPrioValue calcula o valor inteiro composto de (class, data) aceito por SYS_IOPRIO_SET.
func IOPrioValue(class, data int) int {
	return (class << IOPRIO_CLASS_SHIFT) | (data & 0x1fff)
}

// SetIOPriority invoca a syscall nativa SYS_IOPRIO_SET no Linux.
func SetIOPriority(which, who, class, data int) error {
	prio := IOPrioValue(class, data)
	r1, _, err := unix.Syscall(unix.SYS_IOPRIO_SET, uintptr(which), uintptr(who), uintptr(prio))
	if r1 != 0 {
		return err
	}
	return nil
}

// SetLowIOPriority configura a prioridade de E/S do processo atual no Linux.
// Tenta primeiro a classe IOPRIO_CLASS_IDLE (o processo só consome disco quando o sistema estiver ocioso).
// Caso o ambiente restrinja IOPRIO_CLASS_IDLE (ex: falta de CAP_SYS_ADMIN ou política de container),
// realiza fallback automático para IOPRIO_CLASS_BE (Best-Effort) classe 7 (menor prioridade ordinária).
func SetLowIOPriority() error {
	// 1. Tentar classe IDLE (prioridade 7)
	errIdle := SetIOPriority(IOPRIO_WHO_PROCESS, 0, IOPRIO_CLASS_IDLE, 7)
	if errIdle == nil {
		return nil
	}

	// 2. Fallback para Best-Effort classe 7 (menor prioridade sem privilégio de root)
	errBE := SetIOPriority(IOPRIO_WHO_PROCESS, 0, IOPRIO_CLASS_BE, 7)
	if errBE == nil {
		return nil
	}

	return fmt.Errorf("falha ao definir prioridade de I/O no Linux (IDLE: %v, BE: %v)", errIdle, errBE)
}

// SetBackgroundPriority aplica o modo de baixa prioridade de I/O em Linux (IDLE / BE classe 7).
func SetBackgroundPriority() error {
	return SetLowIOPriority()
}

// ResetBackgroundPriority restaura a prioridade de I/O padrão do Linux (Best-Effort classe 4).
func ResetBackgroundPriority() error {
	return SetIOPriority(IOPRIO_WHO_PROCESS, 0, IOPRIO_CLASS_BE, 4)
}
