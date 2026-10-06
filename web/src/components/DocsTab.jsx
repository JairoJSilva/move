import React, { useState, useMemo } from 'react';
import { 
  BookOpen, 
  Search, 
  Copy, 
  Check, 
  FileText, 
  ChevronRight, 
  ShieldCheck, 
  Download,
  Terminal,
  ExternalLink,
  Layers,
  Sparkles,
  Bookmark
} from 'lucide-react';
import { DOCS_VOLUMES } from '../data/docsData';

export function DocsTab() {
  const [selectedId, setSelectedId] = useState('01');
  const [searchQuery, setSearchQuery] = useState('');
  const [copied, setCopied] = useState(false);
  const [viewRaw, setViewRaw] = useState(false);

  const selectedDoc = useMemo(() => {
    return DOCS_VOLUMES.find((d) => d.id === selectedId) || DOCS_VOLUMES[0];
  }, [selectedId]);

  const filteredVolumes = useMemo(() => {
    if (!searchQuery.trim()) return DOCS_VOLUMES;
    const q = searchQuery.toLowerCase();
    return DOCS_VOLUMES.filter(
      (d) =>
        d.title.toLowerCase().includes(q) ||
        d.subtitle.toLowerCase().includes(q) ||
        d.content.toLowerCase().includes(q)
    );
  }, [searchQuery]);

  const handleCopy = () => {
    if (!selectedDoc) return;
    navigator.clipboard.writeText(selectedDoc.content);
    setCopied(true);
    setTimeout(() => setCopied(false), 2500);
  };

  const handleDownload = () => {
    if (!selectedDoc) return;
    const blob = new Blob([selectedDoc.content], { type: 'text/markdown;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = selectedDoc.fileName;
    link.click();
    URL.revokeObjectURL(url);
  };

  // Renderizador simplificado de Markdown nativo para visual corporativo
  const renderMarkdownFormatted = (content) => {
    const lines = content.split('\n');
    const elements = [];
    let inCodeBlock = false;
    let codeBuffer = [];
    let codeLang = '';

    lines.forEach((line, idx) => {
      if (line.startsWith('```')) {
        if (inCodeBlock) {
          elements.push(
            <div key={`code-${idx}`} className="my-4 rounded-xl bg-slate-950 border border-slate-800 p-4 font-mono text-xs overflow-x-auto text-cyan-300 shadow-inner">
              <pre>{codeBuffer.join('\n')}</pre>
            </div>
          );
          codeBuffer = [];
          inCodeBlock = false;
        } else {
          inCodeBlock = true;
          codeLang = line.replace('```', '').trim();
        }
        return;
      }

      if (inCodeBlock) {
        codeBuffer.push(line);
        return;
      }

      // Títulos
      if (line.startsWith('# ')) {
        elements.push(
          <h1 key={idx} className="text-xl md:text-2xl font-black text-white mt-6 mb-3 tracking-tight border-b border-slate-800 pb-2 flex items-center gap-2">
            <Bookmark className="w-5 h-5 text-amber-400 flex-shrink-0" />
            {line.replace('# ', '')}
          </h1>
        );
      } else if (line.startsWith('## ')) {
        elements.push(
          <h2 key={idx} className="text-lg md:text-xl font-bold text-cyan-300 mt-6 mb-2 tracking-tight">
            {line.replace('## ', '')}
          </h2>
        );
      } else if (line.startsWith('### ')) {
        elements.push(
          <h3 key={idx} className="text-base font-bold text-slate-200 mt-4 mb-2">
            {line.replace('### ', '')}
          </h3>
        );
      } else if (line.startsWith('#### ')) {
        elements.push(
          <h4 key={idx} className="text-sm font-semibold text-slate-300 mt-3 mb-1 font-mono">
            {line.replace('#### ', '')}
          </h4>
        );
      } else if (line.startsWith('> ')) {
        elements.push(
          <blockquote key={idx} className="my-3 p-3.5 rounded-xl bg-slate-950/80 border-l-4 border-amber-500 text-xs text-slate-300 italic shadow-sm">
            {line.replace('> ', '')}
          </blockquote>
        );
      } else if (line.startsWith('- ') || line.startsWith('* ')) {
        elements.push(
          <li key={idx} className="text-xs text-slate-300 ml-4 my-1 list-disc leading-relaxed">
            {line.substring(2)}
          </li>
        );
      } else if (/^\d+\.\s/.test(line)) {
        elements.push(
          <p key={idx} className="text-xs text-slate-300 ml-2 my-1 leading-relaxed">
            {line}
          </p>
        );
      } else if (line.trim() === '---') {
        elements.push(<hr key={idx} className="my-6 border-slate-800" />);
      } else if (line.trim().length > 0) {
        elements.push(
          <p key={idx} className="text-xs md:text-sm text-slate-300 my-2 leading-relaxed">
            {line}
          </p>
        );
      }
    });

    return elements;
  };

  return (
    <div className="space-y-6 animate-in fade-in duration-200">
      
      {/* Header do Módulo: Manual do Usuário */}
      <div className="bg-slate-900/60 rounded-2xl border border-slate-800 p-6 backdrop-blur-sm shadow-xl flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-2xl bg-amber-500/10 text-amber-400 border border-amber-500/20 shadow-md">
            <BookOpen className="w-6 h-6" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-xl font-extrabold text-white tracking-tight">
                Manual do Usuário — MoveOps
              </h2>
              <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-amber-500/15 text-amber-300 border border-amber-500/30">
                GUIA COMPLETO
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-0.5">
              Guia operacional passo a passo para migração segura de dados em produção (9 capítulos didáticos).
            </p>
          </div>
        </div>

        {/* Busca rápida */}
        <div className="relative w-full md:w-80">
          <Search className="w-4 h-4 absolute left-3 top-2.5 text-slate-500" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Pesquisar nos capítulos do manual..."
            className="w-full bg-slate-950 border border-slate-800 rounded-xl pl-9 pr-3.5 py-2 text-xs text-slate-200 focus:outline-none focus:border-amber-500 shadow-inner"
          />
        </div>
      </div>

      {/* Conteúdo: Sidebar com 9 Capítulos + Leitor Formatado */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        
        {/* Sidebar com a Lista dos Capítulos (4 colunas em lg) */}
        <div className="lg:col-span-4 space-y-2">
          <div className="flex items-center justify-between px-1 mb-2">
            <span className="text-[11px] font-mono text-slate-500 uppercase tracking-wider font-semibold">
              Capítulos do Manual ({filteredVolumes.length})
            </span>
            <span className="text-[10px] font-mono text-slate-500">
              Passo a Passo
            </span>
          </div>

          <div className="space-y-2 max-h-[750px] overflow-y-auto pr-1">
            {filteredVolumes.map((vol) => {
              const isSelected = vol.id === selectedId;
              return (
                <button
                  key={vol.id}
                  onClick={() => setSelectedId(vol.id)}
                  className={`w-full text-left p-3.5 rounded-xl border transition-all flex flex-col justify-between ${
                    isSelected
                      ? 'bg-amber-950/25 border-amber-500/70 ring-1 ring-amber-500/40 shadow-lg shadow-amber-950/40 text-amber-300'
                      : 'bg-slate-900/60 hover:bg-slate-900 border-slate-800/90 text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <span className="font-bold text-xs text-slate-100 flex items-center gap-1.5">
                      <FileText className={`w-3.5 h-3.5 flex-shrink-0 ${isSelected ? 'text-amber-400' : 'text-slate-500'}`} />
                      {vol.title}
                    </span>
                    <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 flex-shrink-0">
                      Cap {vol.id}
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-400 mt-1 line-clamp-2 leading-snug">
                    {vol.subtitle}
                  </p>
                </button>
              );
            })}
          </div>
        </div>

        {/* Leitor Principal do Manual (8 colunas em lg) */}
        <div className="lg:col-span-8 bg-slate-900/60 rounded-2xl border border-slate-800 backdrop-blur-sm shadow-xl p-6 md:p-8">
          
          {/* Topo do Leitor com Ações */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 border-b border-slate-800 mb-6">
            <div>
              <span className="text-[10px] font-mono font-bold uppercase tracking-wider text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/20">
                Manual do Usuário • Capítulo {selectedDoc.id}
              </span>
              <h3 className="text-lg font-bold text-white mt-1">
                {selectedDoc.title}
              </h3>
            </div>

            <div className="flex items-center gap-2 self-start sm:self-center">
              <button
                onClick={() => setViewRaw(!viewRaw)}
                className={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition ${
                  viewRaw
                    ? 'bg-amber-600 text-white border-amber-500'
                    : 'bg-slate-800 hover:bg-slate-700 text-slate-300 border-slate-700'
                }`}
              >
                {viewRaw ? 'Visualizar Formatado' : 'Ver Markdown Puro'}
              </button>

              <button
                onClick={handleCopy}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition"
                title="Copiar conteúdo"
              >
                {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                <span>{copied ? 'Copiado' : 'Copiar'}</span>
              </button>

              <button
                onClick={handleDownload}
                className="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition"
                title="Baixar arquivo .md deste capítulo"
              >
                <Download className="w-4 h-4" />
              </button>
            </div>
          </div>

          {/* Visualização: Formatado ou Markdown Puro */}
          {viewRaw ? (
            <div className="bg-slate-950 p-4 rounded-xl border border-slate-800 font-mono text-xs text-slate-300 overflow-x-auto whitespace-pre-wrap max-h-[700px] overflow-y-auto">
              {selectedDoc.content}
            </div>
          ) : (
            <div className="prose prose-invert max-w-none max-h-[750px] overflow-y-auto pr-2">
              {renderMarkdownFormatted(selectedDoc.content)}
            </div>
          )}

        </div>

      </div>

    </div>
  );
}
