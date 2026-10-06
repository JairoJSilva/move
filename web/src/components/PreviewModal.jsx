import React from 'react';
import { 
  X, 
  CheckCircle, 
  Clock, 
  FileCheck, 
  FileX, 
  HardDrive, 
  Play, 
  Layers, 
  ArrowRight
} from 'lucide-react';
import { formatBytes, formatDuration } from '../services/api';

export function PreviewModal({ 
  isOpen, 
  onClose, 
  previewData, 
  maxBandwidthMB,
  onStartMigration 
}) {
  if (!isOpen || !previewData) return null;

  // Calcula estimativa aproximada de tempo se houver limite de banda
  let estimatedSeconds = 0;
  if (maxBandwidthMB > 0 && previewData.eligible_bytes_total > 0) {
    estimatedSeconds = Math.round(
      previewData.eligible_bytes_total / (maxBandwidthMB * 1024 * 1024)
    );
  }

  const eligiblePct = previewData.total_files_discovered > 0
    ? ((previewData.eligible_files_count / previewData.total_files_discovered) * 100).toFixed(1)
    : 0;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/85 backdrop-blur-md animate-in fade-in duration-200">
      <div className="w-full max-w-2xl bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl overflow-hidden flex flex-col">
        
        {/* Header */}
        <div className="p-5 border-b border-slate-800 bg-slate-950/60 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-cyan-500/20 text-cyan-400 border border-cyan-500/30">
              <FileCheck className="w-6 h-6" />
            </div>
            <div>
              <h3 className="text-base font-bold text-slate-100">
                Resultado da Estimativa Prévia (Dry Run Preview)
              </h3>
              <p className="text-xs text-slate-400">
                Simulação realizada sem modificar arquivos no destino.
              </p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-slate-800 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content Body */}
        <div className="p-6 space-y-6">
          
          {/* Grid de Métricas Principais */}
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
            
            <div className="p-4 rounded-xl bg-slate-950/70 border border-slate-800">
              <div className="flex items-center gap-2 text-xs text-slate-400 mb-1">
                <FileCheck className="w-4 h-4 text-emerald-400" />
                Arquivos Elegíveis
              </div>
              <div className="text-2xl font-bold font-mono text-emerald-400">
                {previewData.eligible_files_count?.toLocaleString()}
              </div>
              <div className="text-[11px] text-slate-500 font-mono mt-1">
                {eligiblePct}% do total escaneado
              </div>
            </div>

            <div className="p-4 rounded-xl bg-slate-950/70 border border-slate-800">
              <div className="flex items-center gap-2 text-xs text-slate-400 mb-1">
                <HardDrive className="w-4 h-4 text-cyan-400" />
                Volume a Transferir
              </div>
              <div className="text-2xl font-bold font-mono text-cyan-400">
                {formatBytes(previewData.eligible_bytes_total)}
              </div>
              <div className="text-[11px] text-slate-500 font-mono mt-1">
                de {formatBytes(previewData.total_bytes_discovered)}
              </div>
            </div>

            <div className="p-4 rounded-xl bg-slate-950/70 border border-slate-800 col-span-2 sm:col-span-1">
              <div className="flex items-center gap-2 text-xs text-slate-400 mb-1">
                <FileX className="w-4 h-4 text-amber-400" />
                Arquivos Ignorados
              </div>
              <div className="text-2xl font-bold font-mono text-amber-400">
                {previewData.skipped_files_count?.toLocaleString()}
              </div>
              <div className="text-[11px] text-slate-500 font-mono mt-1">
                Filtros ou já sincronizados
              </div>
            </div>

          </div>

          {/* Card Detalhado de Resumo */}
          <div className="p-4 rounded-xl bg-slate-800/40 border border-slate-700/60 space-y-2 text-xs font-mono">
            <div className="flex justify-between py-1 border-b border-slate-800">
              <span className="text-slate-400">Total de Arquivos Escaneados:</span>
              <strong className="text-slate-200">{previewData.total_files_discovered?.toLocaleString()}</strong>
            </div>
            <div className="flex justify-between py-1 border-b border-slate-800">
              <span className="text-slate-400">Total de Pastas Identificadas:</span>
              <strong className="text-slate-200">{previewData.total_dirs_discovered?.toLocaleString()}</strong>
            </div>
            <div className="flex justify-between py-1 border-b border-slate-800">
              <span className="text-slate-400">Volume Total dos Arquivos Ignorados:</span>
              <strong className="text-slate-200">{formatBytes(previewData.skipped_bytes_total)}</strong>
            </div>
            {maxBandwidthMB > 0 ? (
              <div className="flex justify-between py-1 text-cyan-300">
                <span className="flex items-center gap-1.5">
                  <Clock className="w-3.5 h-3.5" /> Tempo Estimado (a {maxBandwidthMB} MB/s):
                </span>
                <strong>{formatDuration(estimatedSeconds)}</strong>
              </div>
            ) : (
              <div className="flex justify-between py-1 text-slate-400">
                <span>Tempo Estimado:</span>
                <span className="italic">Velocidade máxima da mídia (sem limite de banda)</span>
              </div>
            )}
          </div>

        </div>

        {/* Footer */}
        <div className="p-4 border-t border-slate-800 bg-slate-950/70 flex items-center justify-between gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold transition"
          >
            Fechar e Ajustar Parâmetros
          </button>

          <button
            onClick={() => {
              onClose();
              onStartMigration();
            }}
            className="px-5 py-2.5 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-emerald-600/25 transition"
          >
            <Play className="w-4 h-4 fill-current" />
            Iniciar Migração Agora
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>

      </div>
    </div>
  );
}
