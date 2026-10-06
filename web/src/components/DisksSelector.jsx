import React, { useState, useEffect } from 'react';
import { 
  HardDrive, 
  Cpu, 
  Disc, 
  RefreshCw, 
  FolderInput, 
  FolderOutput, 
  Check, 
  AlertTriangle,
  Lock,
  Unlock,
  Layers
} from 'lucide-react';
import { api, formatBytes } from '../services/api';

export function DisksSelector({ 
  sourceDir, 
  setSourceDir, 
  destDir, 
  setDestDir,
  onOpenBrowser
}) {
  const [disks, setDisks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const loadDisks = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getDisks();
      setDisks(data.disks || []);
    } catch (err) {
      setError(err.message || 'Falha ao obter volumes detectados');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDisks();
  }, []);

  return (
    <div className="bg-slate-900/60 rounded-2xl border border-slate-800 p-6 backdrop-blur-sm shadow-xl">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-lg font-bold text-slate-100 flex items-center gap-2">
              <HardDrive className="w-5 h-5 text-cyan-400" />
              Discos & Volumes Detectados no Host
            </h2>
            <span className="px-2 py-0.5 rounded-full text-xs font-mono font-semibold bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
              {disks.length} {disks.length === 1 ? 'Volume' : 'Volumes'}
            </span>
          </div>
          <p className="text-xs text-slate-400 mt-1">
            Reconhecimento nativo de armazenamento (Linux, Windows e Storages). Clique para atribuir diretamente como Origem ou Destino.
          </p>
        </div>

        <button
          onClick={loadDisks}
          disabled={loading}
          className="flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition disabled:opacity-50"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
          Atualizar Discos
        </button>
      </div>

      {error && (
        <div className="mb-6 p-4 rounded-xl bg-rose-500/10 border border-rose-500/30 flex items-center gap-3 text-rose-300 text-sm">
          <AlertTriangle className="w-5 h-5 flex-shrink-0 text-rose-400" />
          <span>{error}</span>
        </div>
      )}

      {loading && disks.length === 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-44 rounded-xl bg-slate-800/40 animate-pulse border border-slate-800/60 p-4"></div>
          ))}
        </div>
      ) : disks.length === 0 ? (
        <div className="p-8 text-center rounded-xl bg-slate-950/40 border border-dashed border-slate-800 text-slate-400">
          <HardDrive className="w-10 h-10 mx-auto text-slate-600 mb-2" />
          <p className="text-sm font-medium">Nenhum volume físico identificado automaticamente.</p>
          <p className="text-xs text-slate-500 mt-1">Você pode especificar o caminho absoluto manualmente abaixo.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {disks.map((disk) => {
            const usedPct = disk.total_bytes > 0 
              ? ((disk.used_bytes / disk.total_bytes) * 100).toFixed(1) 
              : 0;

            const isSource = sourceDir && sourceDir.startsWith(disk.mount_point);
            const isDest = destDir && destDir.startsWith(disk.mount_point);

            return (
              <div
                key={disk.id || disk.mount_point}
                className={`group relative rounded-xl p-4 transition-all duration-200 border ${
                  isSource
                    ? 'bg-blue-950/30 border-blue-500/60 ring-1 ring-blue-500/40 shadow-lg shadow-blue-950/50'
                    : isDest
                    ? 'bg-emerald-950/30 border-emerald-500/60 ring-1 ring-emerald-500/40 shadow-lg shadow-emerald-950/50'
                    : 'bg-slate-950/60 hover:bg-slate-900 border-slate-800/80 hover:border-slate-700'
                }`}
              >
                {/* Header do Card */}
                <div className="flex items-start justify-between gap-2 mb-3">
                  <div className="flex items-center gap-2.5">
                    <div className={`p-2.5 rounded-lg ${
                      disk.is_rotational 
                        ? 'bg-slate-800 text-slate-300' 
                        : 'bg-cyan-500/10 text-cyan-400 border border-cyan-500/20'
                    }`}>
                      {disk.is_rotational ? (
                        <Disc className="w-5 h-5" />
                      ) : (
                        <Cpu className="w-5 h-5" />
                      )}
                    </div>
                    <div>
                      <div className="flex items-center gap-1.5">
                        <span className="font-bold text-sm text-slate-200 font-mono">
                          {disk.mount_point}
                        </span>
                        {disk.label && (
                          <span className="text-xs text-slate-400 truncate max-w-[100px]" title={disk.label}>
                            ({disk.label})
                          </span>
                        )}
                      </div>
                      <p className="text-[11px] text-slate-400 font-mono">
                        {disk.id} • {disk.filesystem.toUpperCase()}
                      </p>
                    </div>
                  </div>

                  {/* Badges de Mídia e Permissão */}
                  <div className="flex flex-col items-end gap-1">
                    <span className={`px-2 py-0.5 rounded text-[10px] font-semibold tracking-wider uppercase font-mono ${
                      disk.is_rotational 
                        ? 'bg-slate-800 text-slate-400' 
                        : 'bg-cyan-950 text-cyan-300 border border-cyan-800/50'
                    }`}>
                      {disk.is_rotational ? 'HDD' : 'SSD / NVMe'}
                    </span>
                    {disk.is_read_only ? (
                      <span className="flex items-center gap-1 text-[10px] text-amber-400 font-mono">
                        <Lock className="w-3 h-3" /> Read-Only
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-[10px] text-emerald-400 font-mono">
                        <Unlock className="w-3 h-3" /> R/W
                      </span>
                    )}
                  </div>
                </div>

                {/* Barra de Progresso de Capacidade */}
                <div className="space-y-1.5 mb-4">
                  <div className="flex justify-between text-xs font-mono">
                    <span className="text-slate-400">
                      Usado: <strong className="text-slate-200">{formatBytes(disk.used_bytes)}</strong> ({usedPct}%)
                    </span>
                    <span className="text-slate-400">
                      Livre: <strong className="text-slate-200">{formatBytes(disk.free_bytes)}</strong>
                    </span>
                  </div>

                  <div className="w-full bg-slate-800/80 rounded-full h-2 overflow-hidden border border-slate-700/50">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${
                        usedPct > 90
                          ? 'bg-gradient-to-r from-amber-500 to-rose-500'
                          : usedPct > 75
                          ? 'bg-gradient-to-r from-blue-500 to-amber-500'
                          : 'bg-gradient-to-r from-cyan-500 to-blue-500'
                      }`}
                      style={{ width: `${Math.min(usedPct, 100)}%` }}
                    />
                  </div>

                  <div className="text-[11px] text-right text-slate-500 font-mono">
                    Capacidade Total: {formatBytes(disk.total_bytes)}
                  </div>
                </div>

                {/* Botões de Seleção Rápida */}
                <div className="grid grid-cols-2 gap-2 pt-2 border-t border-slate-800/60">
                  <button
                    onClick={() => setSourceDir(disk.mount_point)}
                    className={`flex items-center justify-center gap-1.5 py-1.5 px-2 rounded-lg text-xs font-medium transition ${
                      isSource
                        ? 'bg-blue-600 text-white font-semibold'
                        : 'bg-slate-800/80 hover:bg-blue-600/20 text-slate-300 hover:text-blue-300 border border-slate-700/60 hover:border-blue-500/40'
                    }`}
                  >
                    <FolderInput className="w-3.5 h-3.5" />
                    {isSource ? 'Origem Atual' : 'Usar Origem'}
                  </button>

                  <button
                    onClick={() => setDestDir(disk.mount_point)}
                    disabled={disk.is_read_only}
                    className={`flex items-center justify-center gap-1.5 py-1.5 px-2 rounded-lg text-xs font-medium transition ${
                      disk.is_read_only
                        ? 'opacity-40 cursor-not-allowed bg-slate-800 text-slate-500'
                        : isDest
                        ? 'bg-emerald-600 text-white font-semibold'
                        : 'bg-slate-800/80 hover:bg-emerald-600/20 text-slate-300 hover:text-emerald-300 border border-slate-700/60 hover:border-emerald-500/40'
                    }`}
                    title={disk.is_read_only ? 'Volume montado como somente-leitura' : 'Definir como destino'}
                  >
                    <FolderOutput className="w-3.5 h-3.5" />
                    {isDest ? 'Destino Atual' : 'Usar Destino'}
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
