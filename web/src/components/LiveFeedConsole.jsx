import React, { useState, useEffect, useRef } from 'react';
import { 
  Terminal, 
  Search, 
  Trash2, 
  ArrowDownCircle, 
  CheckCircle2, 
  XCircle, 
  MinusCircle, 
  ShieldCheck, 
  Hash, 
  Maximize2
} from 'lucide-react';
import { formatBytes, formatSpeed } from '../services/api';

export function LiveFeedConsole({ events = [], onClearEvents }) {
  const [filterText, setFilterText] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL'); // ALL, COPIED, SKIPPED, FAILED
  const [autoScroll, setAutoScroll] = useState(true);
  const terminalBottomRef = useRef(null);

  useEffect(() => {
    if (autoScroll && terminalBottomRef.current) {
      terminalBottomRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [events, autoScroll]);

  const filteredEvents = events.filter((evt) => {
    if (statusFilter !== 'ALL' && evt.action !== statusFilter) {
      return false;
    }
    if (!filterText) return true;
    const search = filterText.toLowerCase();
    return (
      (evt.source_path && evt.source_path.toLowerCase().includes(search)) ||
      (evt.dest_path && evt.dest_path.toLowerCase().includes(search)) ||
      (evt.source_hash_xx64 && evt.source_hash_xx64.toLowerCase().includes(search)) ||
      (evt.error_message && evt.error_message.toLowerCase().includes(search))
    );
  });

  const formatTime = (ts) => {
    if (!ts) return '--:--:--';
    try {
      const d = new Date(ts);
      return d.toLocaleTimeString('pt-BR');
    } catch {
      return ts;
    }
  };

  return (
    <div className="bg-slate-950/90 rounded-2xl border border-slate-800 flex flex-col h-[520px] overflow-hidden shadow-2xl">
      
      {/* Console Top Bar */}
      <div className="p-3.5 border-b border-slate-800/80 bg-slate-900/80 flex flex-wrap items-center justify-between gap-3">
        
        {/* Esquerda: Identificação */}
        <div className="flex items-center gap-2.5">
          <div className="flex gap-1.5">
            <span className="w-3 h-3 rounded-full bg-rose-500/80 inline-block"></span>
            <span className="w-3 h-3 rounded-full bg-amber-500/80 inline-block"></span>
            <span className="w-3 h-3 rounded-full bg-emerald-500/80 inline-block"></span>
          </div>
          <div className="h-4 w-px bg-slate-700 mx-1"></div>
          <span className="text-xs font-mono font-bold text-slate-200 flex items-center gap-1.5">
            <Terminal className="w-4 h-4 text-cyan-400" />
            Live Feed de Integridade xxHash64
          </span>
          <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-slate-800 text-slate-400 border border-slate-700">
            {filteredEvents.length} eventos
          </span>
        </div>

        {/* Direita: Filtros e Controles */}
        <div className="flex items-center gap-2 flex-wrap">
          
          {/* Status Filter */}
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="bg-slate-950 border border-slate-700 rounded-lg px-2 py-1 text-xs font-mono text-slate-300 focus:outline-none"
          >
            <option value="ALL">Todos os Eventos</option>
            <option value="COPIED">Apenas Copiados</option>
            <option value="SKIPPED">Apenas Pulados</option>
            <option value="FAILED">Apenas Falhas / Erros</option>
          </select>

          {/* Search Input */}
          <div className="relative">
            <Search className="w-3.5 h-3.5 text-slate-500 absolute left-2 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={filterText}
              onChange={(e) => setFilterText(e.target.value)}
              placeholder="Buscar arquivo ou hash..."
              className="bg-slate-950 border border-slate-700 rounded-lg pl-7 pr-2 py-1 text-xs font-mono text-slate-300 w-36 sm:w-48 focus:outline-none focus:border-cyan-500"
            />
          </div>

          {/* Toggle AutoScroll */}
          <button
            onClick={() => setAutoScroll(!autoScroll)}
            className={`p-1.5 rounded-lg border text-xs font-mono transition flex items-center gap-1 ${
              autoScroll
                ? 'bg-cyan-500/20 text-cyan-300 border-cyan-500/40'
                : 'bg-slate-800 text-slate-400 border-slate-700'
            }`}
            title="Fixar no final (Auto-Scroll)"
          >
            <ArrowDownCircle className="w-3.5 h-3.5" />
            <span className="hidden sm:inline text-[11px]">{autoScroll ? 'Scroll ON' : 'Scroll OFF'}</span>
          </button>

          {/* Clear Feed */}
          <button
            onClick={onClearEvents}
            className="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-slate-200 border border-slate-700 transition"
            title="Limpar histórico na tela"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>

        </div>
      </div>

      {/* Console Stream Feed */}
      <div className="flex-1 overflow-y-auto p-4 space-y-2 font-mono text-xs select-text">
        {filteredEvents.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-slate-600 gap-2">
            <Terminal className="w-8 h-8 opacity-40" />
            <span>Nenhum evento registrado no console ainda.</span>
            <span className="text-[11px] text-slate-700">Eventos de cópia e verificação de integridade aparecerão aqui em tempo real.</span>
          </div>
        ) : (
          filteredEvents.map((evt, idx) => {
            const isSuccess = evt.action === 'COPIED';
            const isSkipped = evt.action === 'SKIPPED';
            const isFailed = evt.action === 'FAILED';

            return (
              <div
                key={evt.event_id || idx}
                className={`p-2.5 rounded-xl border transition-all ${
                  isSuccess
                    ? 'bg-slate-900/60 hover:bg-slate-900 border-slate-800/80 hover:border-emerald-500/40'
                    : isFailed
                    ? 'bg-rose-950/20 hover:bg-rose-950/40 border-rose-500/30'
                    : 'bg-slate-900/40 hover:bg-slate-900/60 border-slate-800/60 text-slate-400'
                }`}
              >
                {/* Linha 1: Status, Horário, Caminho e Tamanho */}
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="flex items-center gap-2 truncate">
                    <span className="text-slate-500 text-[11px]">[{formatTime(evt.timestamp)}]</span>

                    {isSuccess && (
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 flex items-center gap-1">
                        <CheckCircle2 className="w-3 h-3" /> COPIADO
                      </span>
                    )}

                    {isSkipped && (
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-slate-800 text-slate-400 border border-slate-700 flex items-center gap-1">
                        <MinusCircle className="w-3 h-3" /> PULADO
                      </span>
                    )}

                    {isFailed && (
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-rose-500/20 text-rose-400 border border-rose-500/30 flex items-center gap-1">
                        <XCircle className="w-3 h-3" /> FALHA
                      </span>
                    )}

                    <span className="text-slate-200 font-semibold truncate max-w-md sm:max-w-xl" title={evt.source_path}>
                      {evt.source_path}
                    </span>
                  </div>

                  <div className="flex items-center gap-3 text-slate-400 text-[11px] flex-shrink-0">
                    {evt.size_bytes > 0 && <span>{formatBytes(evt.size_bytes)}</span>}
                    {evt.duration_ms > 0 && <span>{evt.duration_ms}ms</span>}
                    {evt.throughput_mbs > 0 && (
                      <span className="text-cyan-400 font-bold">{formatSpeed(evt.throughput_mbs)}</span>
                    )}
                  </div>
                </div>

                {/* Linha 2: Verificação xxHash64 ou Mensagem de Erro / Motivo */}
                {isSuccess && evt.source_hash_xx64 && (
                  <div className="mt-1.5 pt-1.5 border-t border-slate-800/60 flex flex-wrap items-center justify-between gap-2 text-[11px]">
                    <div className="flex items-center gap-2 text-slate-400 truncate">
                      <Hash className="w-3 h-3 text-cyan-400 flex-shrink-0" />
                      <span className="text-slate-500">Hash Origem:</span>
                      <span className="text-cyan-300 font-bold">{evt.source_hash_xx64}</span>
                      <span className="text-slate-600">⇄</span>
                      <span className="text-slate-500">Destino:</span>
                      <span className="text-cyan-300 font-bold">{evt.dest_hash_xx64 || evt.source_hash_xx64}</span>
                    </div>

                    <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-bold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 flex-shrink-0">
                      <ShieldCheck className="w-3 h-3" /> xxHash64 MATCH
                    </span>
                  </div>
                )}

                {isSkipped && evt.error_message && (
                  <div className="mt-1 text-[11px] text-slate-500">
                    Motivo: <span className="text-slate-400">{evt.error_message}</span>
                  </div>
                )}

                {isFailed && evt.error_message && (
                  <div className="mt-1 text-[11px] text-rose-400 font-medium">
                    Erro: {evt.error_message}
                  </div>
                )}
              </div>
            );
          })
        )}
        <div ref={terminalBottomRef} />
      </div>

    </div>
  );
}
