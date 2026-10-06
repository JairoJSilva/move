package ratelimit

import (
	"context"
	"math"
	"sync"
	"time"
)

// Limiter implementa um Dual Token Bucket para controle independente de vazão em MB/s e IOPS.
// Suporta reconfiguração a quente (hot-reloading) de forma thread-safe.
type Limiter struct {
	mu sync.Mutex

	// Configurações
	bandwidthMB float64 // <= 0 indica sem limite
	iopsLimit   int64   // <= 0 indica sem limite

	// Token bucket de bytes
	byteRate        float64 // bytes por segundo
	byteCapacity    float64 // capacidade máxima do balde de bytes
	byteTokens      float64 // tokens atuais de bytes
	lastByteRefresh time.Time

	// Token bucket de IOPS
	iopsRate        float64 // operações por segundo
	iopsCapacity    float64 // capacidade máxima do balde de IOPS
	iopsTokens      float64 // tokens atuais de IOPS
	lastIopsRefresh time.Time
}

// NewLimiter cria uma nova instância do Dual Token Bucket Limiter.
// Valores <= 0 para bandwidthMB ou iops indicam tráfego ilimitado.
func NewLimiter(bandwidthMB float64, iops int64) *Limiter {
	l := &Limiter{}
	l.UpdateLimits(bandwidthMB, iops)
	return l
}

// UpdateLimits atualiza as taxas de MB/s e IOPS dinamicamente em tempo de execução.
func (l *Limiter) UpdateLimits(bandwidthMB float64, iops int64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Atualiza tokens existentes com base no tempo decorrido antes de trocar a taxa
	l.refreshByteTokensLocked(now)
	l.refreshIopsTokensLocked(now)

	l.bandwidthMB = bandwidthMB
	l.iopsLimit = iops

	if bandwidthMB > 0 {
		l.byteRate = bandwidthMB * 1024 * 1024
		// Capacidade de burst configurada para 1 segundo de dados, mínimo 2MB
		l.byteCapacity = math.Max(l.byteRate, 2*1024*1024)
		if l.byteTokens > l.byteCapacity {
			l.byteTokens = l.byteCapacity
		}
	} else {
		l.byteRate = 0
		l.byteCapacity = 0
		l.byteTokens = 0
	}

	if iops > 0 {
		l.iopsRate = float64(iops)
		// Capacidade de burst configurada para 1 segundo de IOPS, mínimo 10 ops
		l.iopsCapacity = math.Max(l.iopsRate, 10)
		if l.iopsTokens > l.iopsCapacity {
			l.iopsTokens = l.iopsCapacity
		}
	} else {
		l.iopsRate = 0
		l.iopsCapacity = 0
		l.iopsTokens = 0
	}

	if l.lastByteRefresh.IsZero() {
		l.lastByteRefresh = now
		l.byteTokens = l.byteCapacity
	}
	if l.lastIopsRefresh.IsZero() {
		l.lastIopsRefresh = now
		l.iopsTokens = l.iopsCapacity
	}
}

// GetLimits retorna os limites atuais configurados.
func (l *Limiter) GetLimits() (bandwidthMB float64, iops int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.bandwidthMB, l.iopsLimit
}

func (l *Limiter) refreshByteTokensLocked(now time.Time) {
	if l.byteRate <= 0 {
		return
	}
	if l.lastByteRefresh.IsZero() {
		l.lastByteRefresh = now
		l.byteTokens = l.byteCapacity
		return
	}
	elapsed := now.Sub(l.lastByteRefresh).Seconds()
	if elapsed > 0 {
		l.byteTokens = math.Min(l.byteCapacity, l.byteTokens+(elapsed*l.byteRate))
		l.lastByteRefresh = now
	}
}

func (l *Limiter) refreshIopsTokensLocked(now time.Time) {
	if l.iopsRate <= 0 {
		return
	}
	if l.lastIopsRefresh.IsZero() {
		l.lastIopsRefresh = now
		l.iopsTokens = l.iopsCapacity
		return
	}
	elapsed := now.Sub(l.lastIopsRefresh).Seconds()
	if elapsed > 0 {
		l.iopsTokens = math.Min(l.iopsCapacity, l.iopsTokens+(elapsed*l.iopsRate))
		l.lastIopsRefresh = now
	}
}

// Wait aguarda até que haja tokens suficientes tanto de bytes quanto de IOPS para autorizar a operação.
// Respeita o contexto fornecido (permitindo cancelamento imediato).
func (l *Limiter) Wait(ctx context.Context, bytesCount int64, ioOps int64) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		l.mu.Lock()
		now := time.Now()
		l.refreshByteTokensLocked(now)
		l.refreshIopsTokensLocked(now)

		var waitBytesDur time.Duration
		var waitIopsDur time.Duration

		reqBytes := float64(bytesCount)
		reqOps := float64(ioOps)

		if l.byteRate > 0 && reqBytes > 0 {
			if l.byteTokens < reqBytes {
				deficit := reqBytes - l.byteTokens
				waitSecs := deficit / l.byteRate
				waitBytesDur = time.Duration(waitSecs * float64(time.Second))
			}
		}

		if l.iopsRate > 0 && reqOps > 0 {
			if l.iopsTokens < reqOps {
				deficit := reqOps - l.iopsTokens
				waitSecs := deficit / l.iopsRate
				waitIopsDur = time.Duration(waitSecs * float64(time.Second))
			}
		}

		maxWait := waitBytesDur
		if waitIopsDur > maxWait {
			maxWait = waitIopsDur
		}

		// Se não precisa esperar, consome os tokens e sai
		if maxWait <= 0 {
			if l.byteRate > 0 && reqBytes > 0 {
				l.byteTokens -= reqBytes
			}
			if l.iopsRate > 0 && reqOps > 0 {
				l.iopsTokens -= reqOps
			}
			l.mu.Unlock()
			return nil
		}

		l.mu.Unlock()

		// Sleep cooperativo respeitando contexto
		timer := time.NewTimer(maxWait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Loop para nova verificação com tokens atualizados
		}
	}
}
