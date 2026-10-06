// API Service e Comunicação em Tempo Real

// Detecta URL base de forma resiliente:
// Se servido diretamente pelo Go (ex: porta 8082), usa o caminho relativo '/api/v1'
// Se aberto em porta dev ou standalone, faz fallback automático para 'http://localhost:8082/api/v1'
export function getApiBase() {
  if (typeof window === 'undefined') return 'http://localhost:8082/api/v1';
  if (window.location.port === '8082' || !window.location.port) {
    return '/api/v1';
  }
  if (window.location.port === '3000' || window.location.port === '5173') {
    return '/api/v1'; // Usa o proxy do Vite
  }
  return 'http://localhost:8082/api/v1';
}

export function getWsUrl() {
  if (typeof window === 'undefined') return 'ws://localhost:8082/api/v1/ws';
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  if (window.location.port === '8082' || !window.location.port) {
    return `${protocol}//${window.location.host}/api/v1/ws`;
  }
  if (window.location.port === '3000' || window.location.port === '5173') {
    return `${protocol}//${window.location.host}/api/v1/ws`; // Proxy ws do Vite
  }
  return `${protocol}//localhost:8082/api/v1/ws`;
}

export const api = {
  // 1. Descoberta e Navegação
  async getDisks() {
    const res = await fetch(`${getApiBase()}/disks`);
    if (!res.ok) throw new Error(`Falha ao obter discos: ${res.statusText}`);
    return res.json();
  },

  async browse(path = '', operation = 'read') {
    const res = await fetch(`${getApiBase()}/browse`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path, operation }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || `Falha ao explorar caminho: ${res.statusText}`);
    }
    return res.json();
  },

  // 2. Estimativa Prévia (Dry Run Preview)
  async preview(payload) {
    const res = await fetch(`${getApiBase()}/migration/preview`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || `Erro ao calcular estimativa: ${res.statusText}`);
    }
    return res.json();
  },

  // 3. Ciclo de Vida da Migração
  async startMigration(config) {
    const res = await fetch(`${getApiBase()}/migration/start`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || `Falha ao iniciar migração: ${res.statusText}`);
    }
    return res.json();
  },

  async pauseMigration() {
    const res = await fetch(`${getApiBase()}/migration/pause`, { method: 'POST' });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || 'Falha ao pausar migração');
    }
    return res.json();
  },

  async resumeMigration() {
    const res = await fetch(`${getApiBase()}/migration/resume`, { method: 'POST' });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || 'Falha ao retomar migração');
    }
    return res.json();
  },

  async stopMigration() {
    const res = await fetch(`${getApiBase()}/migration/stop`, { method: 'POST' });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || 'Falha ao parar migração');
    }
    return res.json();
  },

  async getStatus() {
    const res = await fetch(`${getApiBase()}/migration/status`);
    if (!res.ok) throw new Error('Falha ao obter status do motor');
    return res.json();
  },

  // 4. Rate Limiting Dinâmico a Quente
  async updateRateLimit(maxBandwidthMB, maxIOPS) {
    const res = await fetch(`${getApiBase()}/migration/rate-limit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        max_bandwidth_mb: parseFloat(maxBandwidthMB) || 0,
        max_iops: parseInt(maxIOPS, 10) || 0,
      }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || 'Falha ao atualizar limites dinâmicos');
    }
    return res.json();
  },

  // 5. Relatórios e Auditoria
  async getSummaryReport() {
    const res = await fetch(`${getApiBase()}/reports/summary`);
    if (!res.ok) {
      if (res.status === 404) return null;
      throw new Error('Falha ao obter sumário executivo');
    }
    return res.json();
  },

  getExportUrl(format = 'csv') {
    return `${getApiBase()}/reports/export?format=${encodeURIComponent(format)}`;
  },

  // 6. Canal de WebSocket com reconexão
  createWebSocketClient({ onMetrics, onFileEvent, onStatusChange }) {
    let ws = null;
    let reconnectTimer = null;
    let shouldReconnect = true;

    const connect = () => {
      try {
        const wsUrl = getWsUrl();

        onStatusChange?.('CONNECTING');
        ws = new WebSocket(wsUrl);

        ws.onopen = () => {
          onStatusChange?.('CONNECTED');
          if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
          }
        };

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === 'METRICS_UPDATE') {
              onMetrics?.(data.payload);
            } else if (data.type === 'FILE_EVENT') {
              onFileEvent?.(data.payload);
            }
          } catch (e) {
            console.error('Erro ao processar mensagem WS:', e);
          }
        };

        ws.onclose = () => {
          onStatusChange?.('DISCONNECTED');
          if (shouldReconnect) {
            reconnectTimer = setTimeout(connect, 2500);
          }
        };

        ws.onerror = (err) => {
          console.warn('Erro de conexão WebSocket:', err);
          ws?.close();
        };
      } catch (e) {
        console.error('Falha ao inicializar WebSocket:', e);
        if (shouldReconnect) {
          reconnectTimer = setTimeout(connect, 3000);
        }
      }
    };

    connect();

    return {
      disconnect() {
        shouldReconnect = false;
        if (reconnectTimer) clearTimeout(reconnectTimer);
        if (ws) ws.close();
      },
    };
  },
};

// Funções de auxílio e formatação
export function formatBytes(bytes, decimals = 2) {
  if (bytes === 0 || bytes === undefined || bytes === null) return '0 B';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  if (i < 0) return '0 B';
  const safeI = Math.min(i, sizes.length - 1);
  return `${parseFloat((bytes / Math.pow(k, safeI)).toFixed(dm))} ${sizes[safeI]}`;
}

export function formatSpeed(mbPerSec) {
  if (!mbPerSec || mbPerSec <= 0) return '0.00 MB/s';
  if (mbPerSec >= 1024) {
    return `${(mbPerSec / 1024).toFixed(2)} GB/s`;
  }
  return `${mbPerSec.toFixed(2)} MB/s`;
}

export function formatDuration(seconds) {
  if (!seconds || seconds <= 0) return '00:00';
  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);

  if (hrs > 0) {
    return `${hrs}h ${mins < 10 ? '0' : ''}${mins}m ${secs < 10 ? '0' : ''}${secs}s`;
  }
  return `${mins < 10 ? '0' : ''}${mins}m ${secs < 10 ? '0' : ''}${secs}s`;
}

export function formatDateTime(isoString) {
  if (!isoString) return '--';
  try {
    const d = new Date(isoString);
    if (isNaN(d.getTime())) return isoString;
    return d.toLocaleString('pt-BR', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  } catch {
    return isoString;
  }
}
