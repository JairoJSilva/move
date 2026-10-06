import React, { useState, useEffect } from 'react';
import { 
  X, 
  Folder, 
  FolderPlus, 
  File, 
  ArrowUp, 
  Check, 
  Search, 
  ShieldCheck, 
  AlertCircle,
  HardDrive,
  RefreshCw
} from 'lucide-react';
import { api, formatBytes, formatDateTime } from '../services/api';

export function FileBrowserModal({ 
  isOpen, 
  onClose, 
  initialPath = '', 
  targetType = 'source', // 'source' ou 'dest'
  onSelectPath 
}) {
  const [currentPath, setCurrentPath] = useState(initialPath || '/');
  const [pathInput, setPathInput] = useState(initialPath || '/');
  const [entries, setEntries] = useState([]);
  const [isReadable, setIsReadable] = useState(false);
  const [isWritable, setIsWritable] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [searchFilter, setSearchFilter] = useState('');

  const loadDirectory = async (pathToGo) => {
    try {
      setLoading(true);
      setError(null);
      const operation = targetType === 'dest' ? 'write' : 'read';
      const data = await api.browse(pathToGo, operation);
      setCurrentPath(data.current_path);
      setPathInput(data.current_path);
      setEntries(data.entries || []);
      setIsReadable(data.is_readable);
      setIsWritable(data.is_writable);
    } catch (err) {
      setError(err.message || 'Falha ao listar pasta');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      loadDirectory(initialPath || '/');
    }
  }, [isOpen, initialPath]);

  if (!isOpen) return null;

  const handleGoUp = () => {
    if (!currentPath || currentPath === '/' || currentPath === '.') return;
    const parts = currentPath.split(/[/\\]/).filter(Boolean);
    parts.pop();
    const upPath = currentPath.startsWith('/') ? '/' + parts.join('/') : parts.join('/') || '/';
    loadDirectory(upPath);
  };

  const handleSelectCurrent = () => {
    onSelectPath(currentPath);
    onClose();
  };

  const filteredEntries = entries.filter((e) =>
    e.name.toLowerCase().includes(searchFilter.toLowerCase())
  );

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-in fade-in duration-200">
      <div className="w-full max-w-4xl bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl flex flex-col max-h-[85vh] overflow-hidden">
        
        {/* Modal Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-800 bg-slate-950/60">
          <div className="flex items-center gap-2.5">
            <div className={`p-2 rounded-lg ${
              targetType === 'dest' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-blue-500/20 text-blue-400'
            }`}>
              <Folder className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-slate-100">
                Navegador de Pastas • {targetType === 'dest' ? 'Definir Destino' : 'Definir Origem'}
              </h3>
              <p className="text-xs text-slate-400">
                Selecione ou navegue até o diretório desejado no sistema de arquivos do servidor.
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

        {/* Path Input & Controls */}
        <div className="p-4 border-b border-slate-800 bg-slate-900/40 space-y-3">
          <div className="flex gap-2">
            <button
              onClick={handleGoUp}
              disabled={currentPath === '/' || loading}
              className="px-3 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition disabled:opacity-40"
              title="Subir um nível (pasta pai)"
            >
              <ArrowUp className="w-4 h-4" />
              Subir
            </button>

            <form
              onSubmit={(e) => {
                e.preventDefault();
                loadDirectory(pathInput);
              }}
              className="flex-1 flex gap-2"
            >
              <input
                type="text"
                value={pathInput}
                onChange={(e) => setPathInput(e.target.value)}
                placeholder="/caminho/para/pasta"
                className="flex-1 bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs font-mono text-slate-200 focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500"
              />
              <button
                type="submit"
                disabled={loading}
                className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold flex items-center gap-1 transition"
              >
                <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
                Ir
              </button>
            </form>
          </div>

          {/* Status de Permissões e Filtro de Busca */}
          <div className="flex flex-wrap items-center justify-between gap-3 text-xs">
            <div className="flex items-center gap-2">
              <span className={`px-2 py-0.5 rounded font-mono flex items-center gap-1 ${
                isReadable ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30' : 'bg-rose-500/15 text-rose-400 border border-rose-500/30'
              }`}>
                <ShieldCheck className="w-3 h-3" />
                Leitura: {isReadable ? 'Permitida' : 'Negada'}
              </span>

              <span className={`px-2 py-0.5 rounded font-mono flex items-center gap-1 ${
                isWritable ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30' : 'bg-slate-800 text-slate-400 border border-slate-700'
              }`}>
                <ShieldCheck className="w-3 h-3" />
                Escrita: {isWritable ? 'Permitida' : 'Sem Permissão'}
              </span>
            </div>

            <div className="relative min-w-[200px]">
              <Search className="w-3.5 h-3.5 text-slate-500 absolute left-2.5 top-1/2 -translate-y-1/2" />
              <input
                type="text"
                value={searchFilter}
                onChange={(e) => setSearchFilter(e.target.value)}
                placeholder="Filtrar arquivos..."
                className="w-full bg-slate-950 border border-slate-800 rounded-lg pl-8 pr-2.5 py-1 text-xs text-slate-200 focus:outline-none focus:border-slate-700"
              />
            </div>
          </div>
        </div>

        {/* Directory Listing Body */}
        <div className="flex-1 overflow-y-auto p-4 space-y-1">
          {error && (
            <div className="p-3 mb-2 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {loading ? (
            <div className="py-12 text-center text-slate-500 flex flex-col items-center gap-2">
              <RefreshCw className="w-6 h-6 animate-spin text-cyan-400" />
              <span className="text-xs">Carregando diretório...</span>
            </div>
          ) : filteredEntries.length === 0 ? (
            <div className="py-12 text-center text-slate-500 text-xs">
              Nenhum item encontrado nesta pasta.
            </div>
          ) : (
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="border-b border-slate-800 text-slate-500 font-mono text-[11px]">
                  <th className="pb-2 pl-2">Nome</th>
                  <th className="pb-2">Tamanho</th>
                  <th className="pb-2">Modificação</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/40 font-mono">
                {filteredEntries.map((entry) => (
                  <tr
                    key={entry.path}
                    onClick={() => {
                      if (entry.is_dir) {
                        loadDirectory(entry.path);
                      }
                    }}
                    className={`group cursor-pointer transition ${
                      entry.is_dir 
                        ? 'hover:bg-slate-800/70 text-slate-200' 
                        : 'text-slate-400 hover:bg-slate-850/40'
                    }`}
                  >
                    <td className="py-2 pl-2 flex items-center gap-2">
                      {entry.is_dir ? (
                        <Folder className="w-4 h-4 text-cyan-400 group-hover:scale-110 transition flex-shrink-0" />
                      ) : (
                        <File className="w-4 h-4 text-slate-500 flex-shrink-0" />
                      )}
                      <span className={`truncate max-w-[400px] ${entry.is_dir ? 'font-semibold text-cyan-300 group-hover:underline' : ''}`}>
                        {entry.name}
                      </span>
                    </td>
                    <td className="py-2 text-slate-400">
                      {entry.is_dir ? '--' : formatBytes(entry.size)}
                    </td>
                    <td className="py-2 text-slate-500 text-[11px]">
                      {formatDateTime(entry.mod_time)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Modal Footer */}
        <div className="p-4 border-t border-slate-800 bg-slate-950/80 flex items-center justify-between gap-4">
          <div className="text-xs font-mono text-slate-400 truncate max-w-[500px]">
            Pasta selecionada: <span className="text-slate-200 font-bold">{currentPath}</span>
          </div>

          <div className="flex gap-2">
            <button
              onClick={onClose}
              className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold transition"
            >
              Cancelar
            </button>
            <button
              onClick={handleSelectCurrent}
              className={`px-5 py-2 rounded-xl text-white text-xs font-bold flex items-center gap-1.5 shadow-lg transition ${
                targetType === 'dest'
                  ? 'bg-emerald-600 hover:bg-emerald-500 shadow-emerald-600/20'
                  : 'bg-blue-600 hover:bg-blue-500 shadow-blue-600/20'
              }`}
            >
              <Check className="w-4 h-4" />
              Confirmar esta Pasta
            </button>
          </div>
        </div>

      </div>
    </div>
  );
}
