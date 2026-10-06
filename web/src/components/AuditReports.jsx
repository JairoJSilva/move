import React, { useState, useEffect } from 'react';
import { 
  FileText, 
  Download, 
  RefreshCw, 
  CheckCircle2, 
  AlertTriangle, 
  FileDown, 
  Clock, 
  HardDrive, 
  Gauge, 
  Flame, 
  Layers, 
  ShieldCheck,
  FileX
} from 'lucide-react';
import { 
  api, 
  formatBytes, 
  formatSpeed, 
  formatDuration, 
  formatDateTime 
} from '../services/api';

export function AuditReports({ currentJobId }) {
  const [summary, setSummary] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const loadSummary = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getSummaryReport();
      setSummary(data);
    } catch (err) {
      setError(err.message || 'Falha ao carregar relatório');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSummary();
  }, [currentJobId]);

  return (
    <div className="space-y-6">
      
      {/* Top Banner de Relatórios */}
      <div className="bg-slate-900/70 rounded-2xl border border-slate-800 p-6 backdrop-blur-sm shadow-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
              <FileText className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-slate-100 flex items-center gap-2">
                Auditoria & Relatório Executivo
              </h2>
              <p className="text-xs text-slate-400">
                Sumário executivo de migração, trilha de conformidade e exportação em formatos abertos.
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={loadSummary}
            disabled={loading}
            className="flex items-center gap-2 px-3 py-2 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition disabled:opacity-50"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
            Recarregar Relatório
          </button>

          {/* Botões de Exportação */}
          <a
            href={api.getExportUrl('csv')}
            download
            className="flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-600/20 transition"
          >
            <Download className="w-4 h-4" />
            Exportar CSV
          </a>

          <a
            href={api.getExportUrl('jsonl')}
            download
            className="flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold bg-emerald-600 hover:bg-emerald-500 text-white shadow-lg shadow-emerald-600/20 transition"
          >
            <FileDown className="w-4 h-4" />
            Exportar JSON
          </a>
        </div>
      </div>

      {error && (
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-center gap-3 text-amber-300 text-xs">
          <AlertTriangle className="w-5 h-5 flex-shrink-0 text-amber-400" />
          <span>{error}</span>
        </div>
      )}

      {!summary ? (
        <div className="p-12 text-center rounded-2xl bg-slate-900/40 border border-dashed border-slate-800 text-slate-400 flex flex-col items-center gap-3">
          <FileText className="w-12 h-12 text-slate-600" />
          <h3 className="text-base font-bold text-slate-300">Nenhum relatório consolidado no momento</h3>
          <p className="text-xs text-slate-500 max-w-md">
            Inicie um job de migração na aba de configuração para que os registros estruturados de auditoria e os hashes de integridade sejam contabilizados.
          </p>
        </div>
      ) : (
        <div className="space-y-6">
          
          {/* Header do Job Consolidado */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            
            <div className="p-5 rounded-2xl bg-slate-900/70 border border-slate-800">
              <div className="text-xs text-slate-400 mb-1 flex items-center gap-1.5 font-semibold">
                <Clock className="w-4 h-4 text-cyan-400" />
                Duração Total
              </div>
              <div className="text-2xl font-black font-mono text-cyan-400">
                {formatDuration(summary.duration_seconds)}
              </div>
              <div className="text-[11px] text-slate-500 font-mono mt-1">
                Início: {formatDateTime(summary.start_time)}
              </div>
            </div>

            <div className="p-5 rounded-2xl bg-slate-900/70 border border-slate-800">
              <div className="text-xs text-slate-400 mb-1 flex items-center gap-1.5 font-semibold">
                <HardDrive className="w-4 h-4 text-blue-400" />
                Volume Transferido
              </div>
              <div className="text-2xl font-black font-mono text-blue-400">
                {formatBytes(summary.total_bytes_transferred)}
              </div>
              <div className="text-[11px] text-slate-500 font-mono mt-1">
                {summary.total_files_copied?.toLocaleString()} arquivos gravados
              </div>
            </div>

            <div className="p-5 rounded-2xl bg-slate-900/70 border border-slate-800">
              <div className="text-xs text-slate-400 mb-1 flex items-center gap-1.5 font-semibold">
                <Gauge className="w-4 h-4 text-emerald-400" />
                Throughput Médio / Pico
              </div>
              <div className="text-2xl font-black font-mono text-emerald-400">
                {formatSpeed(summary.average_throughput_mbs)}
              </div>
              <div className="text-[11px] text-slate-500 font-mono mt-1">
                Pico máximo: {formatSpeed(summary.peak_throughput_mbs)}
              </div>
            </div>

            <div className="p-5 rounded-2xl bg-slate-900/70 border border-slate-800">
              <div className="text-xs text-slate-400 mb-1 flex items-center gap-1.5 font-semibold">
                <ShieldCheck className="w-4 h-4 text-cyan-400" />
                Índice de Integridade
              </div>
              <div className="text-2xl font-black font-mono text-cyan-300">
                {summary.total_files_failed === 0 ? '100.0%' : 'Com Alertas'}
              </div>
              <div className="text-[11px] text-slate-500 font-mono mt-1">
                {summary.total_files_failed === 0 ? 'Validação xxHash64 perfeita' : `${summary.total_files_failed} arquivos com erro`}
              </div>
            </div>

          </div>

          {/* Sumário Detalhado em Tabela */}
          <div className="p-6 rounded-2xl bg-slate-900/70 border border-slate-800 space-y-4">
            <h3 className="text-sm font-bold text-slate-200 flex items-center gap-2">
              <Layers className="w-4 h-4 text-cyan-400" />
              Métricas Consolidadas do Job: <span className="font-mono text-cyan-300">{summary.job_id}</span>
            </h3>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 pt-2 border-t border-slate-800">
              <div className="p-3 rounded-xl bg-slate-950/60 font-mono">
                <span className="text-[11px] text-slate-500 block">Arquivos Examinados</span>
                <span className="text-lg font-bold text-slate-200">{summary.total_files_examined?.toLocaleString()}</span>
              </div>
              <div className="p-3 rounded-xl bg-slate-950/60 font-mono">
                <span className="text-[11px] text-slate-500 block">Arquivos Copiados</span>
                <span className="text-lg font-bold text-emerald-400">{summary.total_files_copied?.toLocaleString()}</span>
              </div>
              <div className="p-3 rounded-xl bg-slate-950/60 font-mono">
                <span className="text-[11px] text-slate-500 block">Arquivos Pulados</span>
                <span className="text-lg font-bold text-slate-400">{summary.total_files_skipped?.toLocaleString()}</span>
              </div>
              <div className="p-3 rounded-xl bg-slate-950/60 font-mono">
                <span className="text-[11px] text-slate-500 block">Arquivos com Falha</span>
                <span className={`text-lg font-bold ${summary.total_files_failed > 0 ? 'text-rose-400' : 'text-slate-400'}`}>
                  {summary.total_files_failed?.toLocaleString()}
                </span>
              </div>
            </div>
          </div>

          {/* Lista de Falhas ou Erros Registrados (se houver) */}
          {summary.errors && summary.errors.length > 0 && (
            <div className="p-6 rounded-2xl bg-rose-950/20 border border-rose-500/30 space-y-3">
              <h3 className="text-sm font-bold text-rose-300 flex items-center gap-2">
                <FileX className="w-4 h-4 text-rose-400" />
                Detalhamento de Arquivos com Falha ({summary.errors.length})
              </h3>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs font-mono">
                  <thead>
                    <tr className="border-b border-rose-500/30 text-rose-300">
                      <th className="py-2">Origem</th>
                      <th className="py-2">Destino</th>
                      <th className="py-2">Mensagem de Erro</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-rose-500/20 text-rose-200">
                    {summary.errors.map((err, i) => (
                      <tr key={i}>
                        <td className="py-2 pr-3">{err.source_path}</td>
                        <td className="py-2 pr-3">{err.dest_path}</td>
                        <td className="py-2 text-rose-400">{err.error_message || 'Erro desconhecido'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

        </div>
      )}

    </div>
  );
}
