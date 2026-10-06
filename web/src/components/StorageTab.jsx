import React, { useState, useEffect } from 'react';
import { 
  HardDrive, 
  Cpu, 
  Disc, 
  RefreshCw, 
  FolderInput, 
  FolderOutput, 
  Search, 
  Check, 
  AlertTriangle, 
  Lock, 
  Unlock, 
  Sliders, 
  ArrowRight,
  Database,
  Layers,
  Server
} from 'lucide-react';
import { api, formatBytes } from '../services/api';

export function StorageTab({
  sourceDir,
  setSourceDir,
  destDir,
  setDestDir,
  onOpenBrowser,
  setActiveTab
}) {
  const [disks, setDisks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [filterType, setFilterType] = useState('ALL'); // ALL, SSD, HDD

  const loadDisks = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getDisks();
      setDisks(data.disks || []);
    } catch (err) {
      setError(err.message || 'Falha ao obter volumes detectados no host');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDisks();
  }, []);

  const totalBytes = disks.reduce((acc, d) => acc + (d.total_bytes || 0), 0);
  const usedBytes = disks.reduce((acc, d) => acc + (d.used_bytes || 0), 0);
  const freeBytes = disks.reduce((acc, d) => acc + (d.free_bytes || 0), 0);

  const filteredDisks = disks.filter((d) => {
    if (filterType === 'SSD') return !d.is_rotational;
    if (filterType === 'HDD') return d.is_rotational;
    return true;
  });

  return (
    <div className="space-y-6 animate-in fade-in duration-200">
      
      {/* 1. Header do Módulo de Armazenamento */}
      <div className="bg-slate-900/60 rounded-2xl border border-slate-800 p-6 backdrop-blur-sm shadow-xl">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2.5">
              <div className="p-2.5 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
                <HardDrive className="w-6 h-6" />
              </div>
              <div>
                <h2 className="text-xl font-extrabold text-white tracking-tight flex items-center gap-2">
                  Gerenciamento de Armazenamento & Volumes
                </h2>
                <p className="text-xs text-slate-400 mt-0.5">
                  Topologia de discos, sistemas de arquivos e pontos de montagem detectados no host operacional.
                </p>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={loadDisks}
              disabled={loading}
              className="flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition disabled:opacity-50 shadow"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
              Atualizar Discos
            </button>
            <button
              onClick={() => setActiveTab('config')}
              className="flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-extrabold bg-blue-600 hover:bg-blue-500 text-white transition shadow-lg shadow-blue-600/25"
            >
              <span>Ir para Migração</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Resumo Agregado */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mt-6 pt-6 border-t border-slate-800">
          <div className="p-3.5 rounded-xl bg-slate-950/60 border border-slate-800/80">
            <span className="text-[11px] text-slate-500 uppercase tracking-wider font-semibold block">
              Volumes Físicos Detectados
            </span>
            <span className="text-lg font-bold text-slate-100 font-mono mt-0.5 block">
              {disks.length} Pontos de Montagem
            </span>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-950/60 border border-slate-800/80">
            <span className="text-[11px] text-slate-500 uppercase tracking-wider font-semibold block">
              Capacidade Agregada Total
            </span>
            <span className="text-lg font-bold text-cyan-400 font-mono mt-0.5 block">
              {formatBytes(totalBytes)}
            </span>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-950/60 border border-slate-800/80">
            <span className="text-[11px] text-slate-500 uppercase tracking-wider font-semibold block">
              Espaço Livre Total Disponível
            </span>
            <span className="text-lg font-bold text-emerald-400 font-mono mt-0.5 block">
              {formatBytes(freeBytes)} ({totalBytes > 0 ? ((freeBytes / totalBytes) * 100).toFixed(0) : 0}% livre)
            </span>
          </div>
        </div>
      </div>

      {/* Erro de leitura caso ocorra */}
      {error && (
        <div className="p-4 rounded-xl bg-rose-500/10 border border-rose-500/30 flex items-center gap-3 text-rose-300 text-sm">
          <AlertTriangle className="w-5 h-5 flex-shrink-0 text-rose-400" />
          <span>{error}</span>
        </div>
      )}

      {/* Filtros de Mídia */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {[
            { id: 'ALL', label: `Todos (${disks.length})` },
            { id: 'SSD', label: `SSD / NVMe (${disks.filter(d => !d.is_rotational).length})` },
            { id: 'HDD', label: `HDD Rotacional (${disks.filter(d => d.is_rotational).length})` },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setFilterType(tab.id)}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition ${
                filterType === tab.id
                  ? 'bg-slate-800 text-cyan-400 border border-cyan-500/30 shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 bg-slate-950/60 border border-slate-800'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Grid de Discos & Volumes */}
      {loading && disks.length === 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-56 rounded-2xl bg-slate-900/40 animate-pulse border border-slate-800 p-5"></div>
          ))}
        </div>
      ) : filteredDisks.length === 0 ? (
        <div className="p-12 text-center rounded-2xl bg-slate-900/40 border border-dashed border-slate-800 text-slate-400">
          <HardDrive className="w-12 h-12 mx-auto text-slate-600 mb-3" />
          <p className="text-sm font-semibold text-slate-300">Nenhum volume corresponde ao filtro selecionado.</p>
          <p className="text-xs text-slate-500 mt-1">Clique em "Todos" para exibir todos os pontos de montagem.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {filteredDisks.map((disk) => {
            const usedPct = disk.total_bytes > 0 
              ? ((disk.used_bytes / disk.total_bytes) * 100).toFixed(1) 
              : 0;

            const isSource = sourceDir && sourceDir.startsWith(disk.mount_point);
            const isDest = destDir && destDir.startsWith(disk.mount_point);

            return (
              <div
                key={disk.id || disk.mount_point}
                className={`group relative rounded-2xl p-5 transition-all duration-200 border flex flex-col justify-between ${
                  isSource
                    ? 'bg-blue-950/30 border-blue-500/70 ring-2 ring-blue-500/40 shadow-xl shadow-blue-950/60'
                    : isDest
                    ? 'bg-emerald-950/30 border-emerald-500/70 ring-2 ring-emerald-500/40 shadow-xl shadow-emerald-950/60'
                    : 'bg-slate-900/60 hover:bg-slate-900/90 border-slate-800/90 hover:border-slate-700 shadow-lg'
                }`}
              >
                <div>
                  {/* Topo do Card */}
                  <div className="flex items-start justify-between gap-3 mb-4">
                    <div className="flex items-center gap-3">
                      <div className={`p-3 rounded-xl ${
                        disk.is_rotational 
                          ? 'bg-slate-800 text-slate-300 border border-slate-700' 
                          : 'bg-cyan-500/10 text-cyan-400 border border-cyan-500/25'
                      }`}>
                        {disk.is_rotational ? (
                          <Disc className="w-5 h-5" />
                        ) : (
                          <Cpu className="w-5 h-5" />
                        )}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-mono font-black text-base text-slate-100">
                            {disk.mount_point}
                          </span>
                        </div>
                        <p className="text-[11px] text-slate-400 font-mono mt-0.5">
                          {disk.id} • <span className="uppercase font-semibold text-slate-300">{disk.filesystem}</span>
                        </p>
                      </div>
                    </div>

                    {/* Badges de Mídia e Permissão */}
                    <div className="flex flex-col items-end gap-1.5">
                      <span className={`px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase tracking-wider ${
                        disk.is_rotational 
                          ? 'bg-slate-800 text-slate-400 border border-slate-700' 
                          : 'bg-cyan-950/80 text-cyan-300 border border-cyan-800/60'
                      }`}>
                        {disk.is_rotational ? 'HDD' : 'SSD / NVMe'}
                      </span>
                      {disk.is_read_only ? (
                        <span className="flex items-center gap-1 text-[10px] text-amber-400 font-mono bg-amber-950/40 px-1.5 py-0.5 rounded border border-amber-500/30">
                          <Lock className="w-3 h-3" /> Read-Only
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 text-[10px] text-emerald-400 font-mono bg-emerald-950/40 px-1.5 py-0.5 rounded border border-emerald-500/30">
                          <Unlock className="w-3 h-3" /> R/W
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Barra de Progresso de Capacidade */}
                  <div className="space-y-2 mb-5">
                    <div className="flex justify-between text-xs font-mono">
                      <span className="text-slate-400">
                        Usado: <strong className="text-slate-200">{formatBytes(disk.used_bytes)}</strong> ({usedPct}%)
                      </span>
                      <span className="text-slate-400">
                        Livre: <strong className="text-emerald-400">{formatBytes(disk.free_bytes)}</strong>
                      </span>
                    </div>

                    <div className="w-full bg-slate-950 rounded-full h-2.5 overflow-hidden border border-slate-800">
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

                    <div className="flex justify-between text-[11px] text-slate-500 font-mono">
                      <span>Capacidade Total: {formatBytes(disk.total_bytes)}</span>
                      {disk.label && <span className="truncate max-w-[120px]">{disk.label}</span>}
                    </div>
                  </div>
                </div>

                {/* Botões de Ação Direta */}
                <div className="space-y-2 pt-3 border-t border-slate-800">
                  <div className="grid grid-cols-2 gap-2">
                    <button
                      type="button"
                      onClick={() => setSourceDir(disk.mount_point)}
                      className={`flex items-center justify-center gap-1.5 py-2 px-2.5 rounded-xl text-xs font-semibold transition ${
                        isSource
                          ? 'bg-blue-600 text-white shadow-md shadow-blue-600/30'
                          : 'bg-slate-800/80 hover:bg-blue-600/20 text-slate-200 hover:text-blue-300 border border-slate-700/80 hover:border-blue-500/50'
                      }`}
                    >
                      <FolderInput className="w-3.5 h-3.5" />
                      {isSource ? 'Origem Definida' : 'Definir Origem'}
                    </button>

                    <button
                      type="button"
                      onClick={() => setDestDir(disk.mount_point)}
                      disabled={disk.is_read_only}
                      className={`flex items-center justify-center gap-1.5 py-2 px-2.5 rounded-xl text-xs font-semibold transition ${
                        disk.is_read_only
                          ? 'opacity-40 cursor-not-allowed bg-slate-800 text-slate-500'
                          : isDest
                          ? 'bg-emerald-600 text-white shadow-md shadow-emerald-600/30'
                          : 'bg-slate-800/80 hover:bg-emerald-600/20 text-slate-200 hover:text-emerald-300 border border-slate-700/80 hover:border-emerald-500/50'
                      }`}
                      title={disk.is_read_only ? 'Volume montado como somente-leitura' : 'Definir como destino'}
                    >
                      <FolderOutput className="w-3.5 h-3.5" />
                      {isDest ? 'Destino Definido' : 'Definir Destino'}
                    </button>
                  </div>

                  {/* Botão Navegar Arquivos no Volume */}
                  <button
                    type="button"
                    onClick={() => onOpenBrowser('source', disk.mount_point)}
                    className="w-full flex items-center justify-center gap-1.5 py-1.5 px-3 rounded-lg text-xs font-medium bg-slate-950/70 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-800 hover:border-slate-700 transition"
                  >
                    <Search className="w-3.5 h-3.5 text-cyan-400" />
                    <span>Navegar Arquivos neste Volume</span>
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Barra de Rodapé para Ir para Nova Migração */}
      {(sourceDir || destDir) && (
        <div className="p-4 rounded-2xl bg-gradient-to-r from-blue-950/40 via-indigo-950/30 to-slate-900 border border-blue-500/30 flex flex-col sm:flex-row items-center justify-between gap-4 shadow-xl">
          <div className="text-xs font-mono space-y-1">
            <div className="text-slate-300">
              <strong className="text-blue-400">Origem:</strong> {sourceDir || '(vazio)'}
            </div>
            <div className="text-slate-300">
              <strong className="text-emerald-400">Destino:</strong> {destDir || '(vazio)'}
            </div>
          </div>
          <button
            onClick={() => setActiveTab('config')}
            className="px-5 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-blue-600/30 transition"
          >
            <Sliders className="w-4 h-4" />
            Configurar Migração com estes Caminhos
          </button>
        </div>
      )}

    </div>
  );
}
