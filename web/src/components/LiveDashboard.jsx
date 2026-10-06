import React, { useState } from 'react';
import { 
  Activity, 
  Play, 
  Pause, 
  RotateCcw, 
  Square, 
  Zap, 
  Gauge, 
  Cpu, 
  Clock, 
  Timer, 
  FileCheck2, 
  FileX2, 
  FileWarning, 
  HardDrive, 
  CheckCircle2, 
  AlertTriangle,
  Flame,
  Layers,
  Sliders,
  Check
} from 'lucide-react';
import { 
  formatBytes, 
  formatSpeed, 
  formatDuration, 
  api 
} from '../services/api';

export function LiveDashboard({
  metrics,
  throughputHistory = [],
  onStart,
  onPause,
  onResume,
  onStop,
  onSwitchToConfig
}) {
  const [hotBandwidth, setHotBandwidth] = useState(metrics?.limit_bandwidth_mb || 0);
  const [hotIOPS, setHotIOPS] = useState(metrics?.limit_iops || 0);
  const [applyingLimits, setApplyingLimits] = useState(false);
  const [limitsFeedback, setLimitsFeedback] = useState(null);
  const [showStopConfirm, setShowStopConfirm] = useState(false);

  const status = metrics?.status || 'IDLE';
  const isRunning = status === 'RUNNING';
  const isPaused = status === 'PAUSED';
  const isCompleted = status === 'COMPLETED';
  const isIdle = status === 'IDLE';

  const progress = metrics?.progress_percent ? metrics.progress_percent.toFixed(1) : '0.0';
  const throughput = metrics?.current_throughput_mbs || 0;
  const iops = metrics?.current_iops || 0;

  const handleApplyHotLimits = async () => {
    try {
      setApplyingLimits(true);
      await api.updateRateLimit(hotBandwidth, hotIOPS);
      setLimitsFeedback('Limites atualizados com sucesso a quente!');
      setTimeout(() => setLimitsFeedback(null), 3000);
    } catch (err) {
      setLimitsFeedback(`Erro: ${err.message}`);
    } finally {
      setApplyingLimits(false);
    }
  };

  // Sparkline SVG helper
  const renderSparkline = () => {
    if (throughputHistory.length < 2) {
      return (
        <div className="h-16 flex items-center justify-center text-[11px] text-slate-500 font-mono">
          Aguardando amostras de telemetria...
        </div>
      );
    }

    const maxVal = Math.max(...throughputHistory, 10);
    const width = 300;
    const height = 60;
    const points = throughputHistory.map((val, idx) => {
      const x = (idx / (throughputHistory.length - 1)) * width;
      const y = height - (val / maxVal) * (height - 10) - 5;
      return `${x},${y}`;
    }).join(' ');

    return (
      <svg className="w-full h-16 overflow-visible" viewBox={`0 0 ${width} ${height}`}>
        <defs>
          <linearGradient id="speedGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#06b6d4" stopOpacity="0.4" />
            <stop offset="100%" stopColor="#06b6d4" stopOpacity="0.0" />
          </linearGradient>
        </defs>
        <polygon
          fill="url(#speedGrad)"
          points={`0,${height} ${points} ${width},${height}`}
        />
        <polyline
          fill="none"
          stroke="#06b6d4"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeLinejoin="round"
          points={points}
        />
      </svg>
    );
  };

  return (
    <div className="space-y-6">
      
      {/* Top Banner de Status e Botões de Controle */}
      <div className="bg-slate-900/70 rounded-2xl border border-slate-800 p-6 backdrop-blur-sm shadow-xl flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
        <div>
          <div className="flex items-center gap-3">
            <h2 className="text-xl font-black text-slate-100 tracking-tight flex items-center gap-2">
              <Activity className="w-6 h-6 text-cyan-400" />
              Console Operacional em Tempo Real
            </h2>
            {metrics?.job_id && (
              <span className="px-2.5 py-0.5 rounded-lg text-xs font-mono font-bold bg-slate-800 text-slate-300 border border-slate-700">
                Job: {metrics.job_id}
              </span>
            )}
          </div>
          <p className="text-xs text-slate-400 mt-1">
            Telemetria instantânea, verificação de integridade xxHash64 e controles de vazão.
          </p>
        </div>

        {/* Botões do Ciclo de Vida */}
        <div className="flex flex-wrap items-center gap-2.5">
          {isRunning && (
            <button
              onClick={onPause}
              className="px-4 py-2 rounded-xl bg-amber-500/20 hover:bg-amber-500/30 text-amber-300 text-xs font-bold border border-amber-500/40 flex items-center gap-2 transition"
            >
              <Pause className="w-4 h-4 fill-current" />
              Pausar Job
            </button>
          )}

          {isPaused && (
            <button
              onClick={onResume}
              className="px-4 py-2 rounded-xl bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-300 text-xs font-bold border border-emerald-500/40 flex items-center gap-2 transition"
            >
              <Play className="w-4 h-4 fill-current" />
              Retomar Migração
            </button>
          )}

          {(isRunning || isPaused) && (
            <>
              {showStopConfirm ? (
                <div className="flex items-center gap-2 p-1 bg-rose-950/80 rounded-xl border border-rose-500/40">
                  <span className="text-[11px] text-rose-300 px-2 font-semibold">Confirmar cancelamento?</span>
                  <button
                    onClick={() => {
                      setShowStopConfirm(false);
                      onStop();
                    }}
                    className="px-3 py-1 bg-rose-600 hover:bg-rose-500 text-white rounded-lg text-xs font-bold transition"
                  >
                    Sim, Parar
                  </button>
                  <button
                    onClick={() => setShowStopConfirm(false)}
                    className="px-2 py-1 bg-slate-800 text-slate-300 rounded-lg text-xs hover:bg-slate-700 transition"
                  >
                    Voltar
                  </button>
                </div>
              ) : (
                <button
                  onClick={() => setShowStopConfirm(true)}
                  className="px-4 py-2 rounded-xl bg-rose-500/20 hover:bg-rose-500/30 text-rose-300 text-xs font-bold border border-rose-500/40 flex items-center gap-2 transition"
                >
                  <Square className="w-3.5 h-3.5 fill-current" />
                  Abortar Migração
                </button>
              )}
            </>
          )}

          {(isIdle || isCompleted) && (
            <button
              onClick={onSwitchToConfig}
              className="px-5 py-2.5 rounded-xl bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-500 hover:to-cyan-500 text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-blue-500/25 transition"
            >
              <Play className="w-4 h-4 fill-current" />
              Nova Migração / Configuração
            </button>
          )}
        </div>
      </div>

      {/* Barra de Progresso Master */}
      <div className="bg-slate-900/70 rounded-2xl border border-slate-800 p-6 backdrop-blur-sm shadow-xl space-y-4">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
          <div>
            <div className="flex items-center gap-2">
              <span className="text-3xl font-extrabold font-mono tracking-tight bg-gradient-to-r from-cyan-400 via-blue-400 to-indigo-400 bg-clip-text text-transparent">
                {progress}%
              </span>
              <span className="text-xs text-slate-400 font-semibold uppercase tracking-wider">
                Concluído
              </span>
            </div>
            <p className="text-xs text-slate-400 font-mono mt-0.5">
              {formatBytes(metrics?.bytes_transferred || 0)} transferidos de {formatBytes(metrics?.total_bytes_discovered || 0)}
            </p>
          </div>

          <div className="flex items-center gap-4 text-xs font-mono">
            <div className="text-right">
              <span className="text-slate-500 block text-[10px] uppercase">Tempo Decorrido</span>
              <span className="text-slate-200 font-bold flex items-center gap-1 justify-end">
                <Timer className="w-3.5 h-3.5 text-blue-400" />
                {formatDuration(metrics?.elapsed_seconds || 0)}
              </span>
            </div>

            <div className="text-right pl-4 border-l border-slate-800">
              <span className="text-slate-500 block text-[10px] uppercase">Estimativa (ETA)</span>
              <span className="text-cyan-400 font-bold flex items-center gap-1 justify-end">
                <Clock className="w-3.5 h-3.5 text-cyan-400" />
                {metrics?.eta_seconds > 0 ? formatDuration(metrics.eta_seconds) : isCompleted ? 'Finalizado' : '--:--'}
              </span>
            </div>
          </div>
        </div>

        {/* Progress Bar Container */}
        <div className="w-full bg-slate-950 rounded-full h-4 overflow-hidden border border-slate-800/80 p-0.5 relative">
          <div
            className={`h-full rounded-full transition-all duration-300 relative ${
              isCompleted
                ? 'bg-gradient-to-r from-emerald-500 to-teal-400'
                : 'bg-gradient-to-r from-blue-600 via-cyan-500 to-teal-400'
            }`}
            style={{ width: `${Math.min(parseFloat(progress), 100)}%` }}
          >
            {isRunning && (
              <div className="absolute inset-0 bg-white/20 animate-pulse rounded-full" />
            )}
          </div>
        </div>

        {/* Arquivo Sendo Processado Agora */}
        <div className="flex items-center gap-2 p-2.5 rounded-xl bg-slate-950/60 border border-slate-800/80 text-xs font-mono">
          <span className="text-slate-500 font-bold flex-shrink-0">ARQUIVO ATUAL:</span>
          {metrics?.current_file ? (
            <span className="text-cyan-300 truncate" title={metrics.current_file}>
              {metrics.current_file}
            </span>
          ) : (
            <span className="text-slate-600 italic">
              {isRunning ? 'Escaneando diretórios...' : 'Aguardando início...'}
            </span>
          )}
        </div>
      </div>

      {/* Cards de Métricas e Gráficos */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        
        {/* Velocidade / Throughput */}
        <div className="p-5 rounded-2xl bg-slate-900/70 border border-slate-800 relative overflow-hidden flex flex-col justify-between">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
            <span className="flex items-center gap-1.5 font-semibold">
              <Gauge className="w-4 h-4 text-cyan-400" />
              Velocidade Instantânea
            </span>
            <span className="text-[10px] font-mono text-slate-500">Média 500ms</span>
          </div>

          <div className="my-1">
            <span className="text-3xl font-extrabold font-mono text-cyan-400">
              {formatSpeed(throughput)}
            </span>
          </div>

          <div className="mt-2">
            {renderSparkline()}
          </div>
        </div>

        {/* IOPS */}
        <div className="p-5 rounded-2xl bg-slate-900/70 border border-slate-800 flex flex-col justify-between">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
            <span className="flex items-center gap-1.5 font-semibold">
              <Flame className="w-4 h-4 text-amber-400" />
              Taxa de IOPS
            </span>
            <span className="text-[10px] font-mono text-slate-500">Arquivos/s</span>
          </div>

          <div className="my-auto py-2">
            <span className="text-3xl font-extrabold font-mono text-amber-400">
              {iops.toLocaleString()}
            </span>
            <span className="text-xs text-slate-500 font-mono ml-1">op/s</span>
          </div>

          <div className="text-[11px] text-slate-500 font-mono pt-3 border-t border-slate-800/80">
            Workers Concorrentes: <strong className="text-slate-300">{metrics?.active_workers || 0}</strong> ativos
          </div>
        </div>

        {/* Contagem de Arquivos */}
        <div className="p-5 rounded-2xl bg-slate-900/70 border border-slate-800 flex flex-col justify-between">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
            <span className="flex items-center gap-1.5 font-semibold">
              <FileCheck2 className="w-4 h-4 text-emerald-400" />
              Arquivos Copiados
            </span>
            <span className="text-[10px] font-mono text-slate-500">Total</span>
          </div>

          <div className="my-auto py-2">
            <span className="text-3xl font-extrabold font-mono text-emerald-400">
              {(metrics?.files_copied || 0).toLocaleString()}
            </span>
            <span className="text-xs text-slate-500 font-mono ml-1">
              / {(metrics?.total_files_discovered || 0).toLocaleString()}
            </span>
          </div>

          <div className="flex justify-between text-[11px] text-slate-500 font-mono pt-3 border-t border-slate-800/80">
            <span>Pulados: <strong className="text-slate-300">{metrics?.files_skipped || 0}</strong></span>
            <span>Falhas: <strong className="text-rose-400">{metrics?.files_failed || 0}</strong></span>
          </div>
        </div>

        {/* Status de Integridade xxHash64 */}
        <div className="p-5 rounded-2xl bg-slate-900/70 border border-slate-800 flex flex-col justify-between">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
            <span className="flex items-center gap-1.5 font-semibold">
              <CheckCircle2 className="w-4 h-4 text-cyan-400" />
              Integridade do Pipeline
            </span>
            <span className="text-[10px] font-mono text-emerald-400">xxHash64</span>
          </div>

          <div className="my-auto py-2">
            <span className="text-2xl font-bold font-mono text-cyan-300">
              {metrics?.files_failed === 0 ? '100% ÍNTEGRO' : `${metrics?.files_failed || 0} FALHAS`}
            </span>
          </div>

          <div className="text-[11px] text-slate-400 font-mono pt-3 border-t border-slate-800/80">
            Verificação dual de hash ponta a ponta
          </div>
        </div>

      </div>

      {/* Painel de Ajuste Dinâmico a Quente (Hot Rate-Limiting) */}
      <div className="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-sm">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-4">
          <div>
            <h3 className="text-sm font-bold text-slate-200 flex items-center gap-2">
              <Sliders className="w-4 h-4 text-cyan-400" />
              Ajuste Dinâmico a Quente (Hot Throttling em Tempo Real)
            </h3>
            <p className="text-[11px] text-slate-400">
              Altere a velocidade e o impacto de I/O em tempo de execução sem interromper o trabalho em andamento.
            </p>
          </div>

          {limitsFeedback && (
            <span className="text-xs font-mono px-3 py-1 rounded-lg bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
              {limitsFeedback}
            </span>
          )}
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-center">
          
          <div className="space-y-1">
            <div className="flex justify-between text-xs font-mono">
              <span className="text-slate-400">Limite Banda (MB/s):</span>
              <strong className="text-cyan-400">{hotBandwidth == 0 ? 'Ilimitado' : `${hotBandwidth} MB/s`}</strong>
            </div>
            <input
              type="range"
              min="0"
              max="1000"
              step="10"
              value={hotBandwidth}
              onChange={(e) => setHotBandwidth(e.target.value)}
              className="w-full accent-cyan-400"
            />
          </div>

          <div className="space-y-1">
            <div className="flex justify-between text-xs font-mono">
              <span className="text-slate-400">Limite IOPS:</span>
              <strong className="text-amber-400">{hotIOPS == 0 ? 'Ilimitado' : `${hotIOPS} IOPS`}</strong>
            </div>
            <input
              type="range"
              min="0"
              max="5000"
              step="50"
              value={hotIOPS}
              onChange={(e) => setHotIOPS(e.target.value)}
              className="w-full accent-amber-400"
            />
          </div>

          <div>
            <button
              onClick={handleApplyHotLimits}
              disabled={applyingLimits}
              className="w-full py-2.5 px-4 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-bold border border-slate-700 flex items-center justify-center gap-2 transition disabled:opacity-50"
            >
              <Zap className="w-3.5 h-3.5 text-cyan-400" />
              {applyingLimits ? 'Aplicando...' : 'Aplicar Limites a Quente'}
            </button>
          </div>

        </div>
      </div>

    </div>
  );
}
