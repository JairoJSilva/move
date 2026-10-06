import React from 'react';
import { 
  LayoutDashboard,
  HardDrive, 
  Sliders,
  Activity, 
  FileText, 
  BookOpen,
  Radio, 
  Layers, 
  Sun, 
  Moon, 
  CheckCircle2, 
  PauseCircle, 
  XCircle, 
  ShieldCheck
} from 'lucide-react';

export function Navbar({ 
  activeTab, 
  setActiveTab, 
  wsStatus, 
  metrics, 
  darkMode, 
  setDarkMode 
}) {
  const getStatusBadge = () => {
    const status = metrics?.status || 'IDLE';
    switch (status) {
      case 'RUNNING':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 animate-pulse">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span>
            EM EXECUÇÃO
          </span>
        );
      case 'PAUSED':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-amber-500/15 text-amber-400 border border-amber-500/30">
            <PauseCircle className="w-3.5 h-3.5" />
            PAUSADO
          </span>
        );
      case 'COMPLETED':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-cyan-500/15 text-cyan-400 border border-cyan-500/30">
            <CheckCircle2 className="w-3.5 h-3.5" />
            CONCLUÍDO
          </span>
        );
      case 'CANCELLED':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-rose-500/15 text-rose-400 border border-rose-500/30">
            <XCircle className="w-3.5 h-3.5" />
            CANCELADO
          </span>
        );
      case 'FAILED':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-red-500/15 text-red-400 border border-red-500/30">
            <XCircle className="w-3.5 h-3.5" />
            FALHA
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-slate-800 text-slate-400 border border-slate-700">
            <span className="w-2 h-2 rounded-full bg-slate-500"></span>
            OCIOSO (IDLE)
          </span>
        );
    }
  };

  const getPhaseBadge = () => {
    if (!metrics?.current_phase || metrics.status === 'IDLE') return null;
    let label = 'Fase 1: Baseline';
    if (metrics.current_phase === 'PHASE_2_DELTA') label = 'Fase 2: Delta Sync';
    if (metrics.current_phase === 'PHASE_3_CUTOVER') label = 'Fase 3: Cutover';

    return (
      <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded text-xs font-medium bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
        <Layers className="w-3 h-3" />
        {label}
      </span>
    );
  };

  const navItems = [
    { id: 'overview', label: 'Visão Geral', icon: LayoutDashboard },
    { id: 'storage', label: 'Armazenamento', icon: HardDrive },
    { id: 'config', label: 'Nova Migração', icon: Sliders },
    { id: 'telemetry', label: 'Telemetria & Ao Vivo', icon: Activity, hasLiveBadge: metrics?.status === 'RUNNING' },
    { id: 'audit', label: 'Auditoria & Relatórios', icon: FileText },
    { id: 'docs', label: 'Manual do Usuário', icon: BookOpen },
  ];

  return (
    <header className="sticky top-0 z-40 w-full border-b border-slate-800/80 bg-slate-950/90 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-20 gap-4">
          
          {/* Logo & Marca Oficial */}
          <div 
            onClick={() => setActiveTab('overview')}
            className="flex items-center gap-3.5 cursor-pointer group flex-shrink-0 py-1"
            title="Ir para Visão Geral"
          >
            <img 
              src="/logo-inteira.jpeg" 
              alt="MoveOps" 
              className="h-12 md:h-14 w-auto max-w-[220px] object-contain rounded-xl shadow-md border border-slate-700/60 p-0.5 group-hover:scale-105 group-hover:border-cyan-500/50 transition duration-200"
              onError={(e) => {
                e.target.src = '/logo-n-fundo.png';
              }}
            />
            <div className="hidden sm:block">
              <div className="flex items-center gap-2">
                <span className="font-extrabold text-lg tracking-tight bg-gradient-to-r from-white via-slate-100 to-slate-300 bg-clip-text text-transparent group-hover:from-cyan-300 group-hover:to-white transition">
                  MoveOps
                </span>
                <span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 tracking-wider">
                  ENTERPRISE
                </span>
              </div>
              <p className="text-[10px] text-slate-400 flex items-center gap-1 font-mono">
                <ShieldCheck className="w-3 h-3 text-emerald-400" />
                Zero-Downtime Engine
              </p>
            </div>
          </div>

          {/* Abas Centrais de Navegação Segmentada */}
          <nav className="hidden lg:flex items-center gap-1 bg-slate-900/90 p-1 rounded-xl border border-slate-800/90">
            {navItems.map((item) => {
              const Icon = item.icon;
              const isActive = activeTab === item.id;
              return (
                <button
                  key={item.id}
                  onClick={() => setActiveTab(item.id)}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all relative ${
                    isActive
                      ? 'bg-gradient-to-r from-blue-600 to-cyan-600 text-white shadow-md shadow-blue-500/25 ring-1 ring-cyan-400/30'
                      : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
                  }`}
                >
                  <Icon className="w-3.5 h-3.5" />
                  <span>{item.label}</span>
                  {item.hasLiveBadge && (
                    <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping ml-0.5"></span>
                  )}
                </button>
              );
            })}
          </nav>

          {/* Status & Controles Laterais */}
          <div className="flex items-center gap-2.5 flex-shrink-0">
            {getPhaseBadge()}
            {getStatusBadge()}

            {/* Status do WebSocket */}
            <div 
              title={`Conexão WebSocket: ${wsStatus}`}
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-900 border border-slate-800"
            >
              <Radio className={`w-3 h-3 ${
                wsStatus === 'CONNECTED' 
                  ? 'text-emerald-400 animate-pulse' 
                  : wsStatus === 'CONNECTING' 
                  ? 'text-amber-400' 
                  : 'text-rose-400'
              }`} />
              <span className="hidden sm:inline text-slate-400 text-[11px]">
                {wsStatus === 'CONNECTED' ? 'LIVE' : wsStatus}
              </span>
            </div>

            {/* Alternador de Tema */}
            <button
              onClick={() => setDarkMode(!darkMode)}
              className="p-2 rounded-lg bg-slate-900 text-slate-400 hover:text-slate-200 border border-slate-800 hover:border-slate-700 transition"
              title="Alternar tema escuro/claro"
            >
              {darkMode ? <Sun className="w-4 h-4 text-amber-400" /> : <Moon className="w-4 h-4 text-indigo-400" />}
            </button>
          </div>

        </div>

        {/* Barra de Navegação Mobile/Telas menores que lg */}
        <div className="lg:hidden flex items-center justify-start gap-1 pb-2 overflow-x-auto border-t border-slate-800/40 pt-2">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={`flex items-center gap-1 px-2.5 py-1 rounded-md text-[11px] font-semibold whitespace-nowrap transition-all ${
                  isActive
                    ? 'bg-blue-600 text-white'
                    : 'text-slate-400 hover:text-slate-200 bg-slate-900/60'
                }`}
              >
                <Icon className="w-3 h-3" />
                <span>{item.label}</span>
                {item.hasLiveBadge && (
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping"></span>
                )}
              </button>
            );
          })}
        </div>

      </div>
    </header>
  );
}
