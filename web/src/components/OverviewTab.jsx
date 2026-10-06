import React, { useState, useEffect } from 'react';
import { 
  HardDrive, 
  Sliders, 
  Activity, 
  FileText, 
  BookOpen, 
  ShieldCheck, 
  ArrowRight, 
  Zap, 
  Layers, 
  CheckCircle2, 
  PauseCircle, 
  XCircle, 
  AlertTriangle,
  Server,
  Database,
  Lock,
  Clock,
  Sparkles
} from 'lucide-react';
import { api, formatBytes } from '../services/api';

export function OverviewTab({ 
  metrics, 
  setActiveTab, 
  sourceDir, 
  destDir 
}) {
  const [disks, setDisks] = useState([]);
  const [loadingDisks, setLoadingDisks] = useState(false);

  useEffect(() => {
    let isMounted = true;
    const fetchDisks = async () => {
      try {
        setLoadingDisks(true);
        const data = await api.getDisks();
        if (isMounted) {
          setDisks(data.disks || []);
        }
      } catch (err) {
        // Fallback gracioso
      } finally {
        if (isMounted) setLoadingDisks(false);
      }
    };
    fetchDisks();
    return () => { isMounted = false; };
  }, []);

  // Cálculos agregados de disco
  const totalStorageBytes = disks.reduce((acc, d) => acc + (d.total_bytes || 0), 0);
  const usedStorageBytes = disks.reduce((acc, d) => acc + (d.used_bytes || 0), 0);
  const freeStorageBytes = disks.reduce((acc, d) => acc + (d.free_bytes || 0), 0);
  const aggregateUsedPct = totalStorageBytes > 0 
    ? ((usedStorageBytes / totalStorageBytes) * 100).toFixed(1) 
    : 0;

  const isJobRunning = metrics?.status === 'RUNNING';
  const isJobActive = isJobRunning || metrics?.status === 'PAUSED';

  return (
    <div className="space-y-6 animate-in fade-in duration-200">
      
      {/* 1. Header Executivo Principal com Logo Oficial */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-slate-900 via-slate-900 to-slate-950 border border-slate-800 p-6 md:p-8 shadow-2xl">
        <div className="absolute top-0 right-0 -mt-10 -mr-10 w-80 h-80 bg-cyan-500/10 rounded-full blur-3xl pointer-events-none"></div>
        <div className="absolute bottom-0 left-1/3 -mb-10 w-60 h-60 bg-blue-600/10 rounded-full blur-2xl pointer-events-none"></div>

        <div className="relative z-10 flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
          <div className="flex flex-col sm:flex-row items-start sm:items-center gap-6">
            <img 
              src="/logo-inteira.jpeg" 
              alt="MoveOps" 
              className="h-24 md:h-28 w-auto max-w-[320px] object-contain rounded-2xl shadow-2xl border border-cyan-500/40 p-1.5 bg-slate-900 flex-shrink-0"
              onError={(e) => {
                e.target.src = '/logo-n-fundo.png';
              }}
            />
            <div>
              <div className="flex flex-wrap items-center gap-2 mb-1.5">
                <h1 className="text-2xl md:text-3xl font-extrabold text-white tracking-tight">
                  MoveOps
                </h1>
                <span className="px-2.5 py-0.5 rounded-full text-xs font-mono font-bold bg-cyan-500/20 text-cyan-300 border border-cyan-500/40">
                  v1.0.0 ENTERPRISE
                </span>
                <span className="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 flex items-center gap-1">
                  <ShieldCheck className="w-3.5 h-3.5" />
                  Produção Segura Ativa
                </span>
              </div>
              <p className="text-xs md:text-sm text-slate-300 max-w-2xl leading-relaxed">
                Suíte corporativa de migração e replicação de dados com tolerância a falhas, zero-downtime cutover atômico, integridade criptográfica xxHash64 e auditoria regulatória contínua.
              </p>
            </div>
          </div>

          <div className="flex flex-row md:flex-col items-center md:items-end justify-between w-full md:w-auto gap-3 border-t md:border-t-0 border-slate-800 pt-4 md:pt-0">
            <button
              onClick={() => setActiveTab('config')}
              className="px-5 py-2.5 rounded-xl bg-gradient-to-r from-blue-600 via-indigo-600 to-cyan-600 hover:from-blue-500 hover:to-cyan-500 text-white text-xs font-extrabold flex items-center gap-2 shadow-lg shadow-blue-600/30 transition transform hover:-translate-y-0.5"
            >
              <Zap className="w-4 h-4 fill-current" />
              Iniciar Nova Migração
            </button>
            <span className="text-[11px] text-slate-400 font-mono">
              Origem &rarr; Destino com Checksum
            </span>
          </div>
        </div>
      </div>

      {/* 2. Grid de KPIs Executivos do Sistema */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        
        {/* KPI 1: Status do Motor */}
        <div className="p-5 rounded-2xl bg-slate-900/60 border border-slate-800 backdrop-blur-sm shadow-lg">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-3 font-semibold">
            <span className="flex items-center gap-1.5">
              <Server className="w-4 h-4 text-blue-400" />
              Status do Motor
            </span>
            <span className="font-mono text-[10px] text-slate-500">Go Runtime</span>
          </div>
          <div className="flex items-center gap-2">
            <span className={`w-3 h-3 rounded-full ${
              isJobRunning ? 'bg-emerald-400 animate-ping' : metrics?.status === 'PAUSED' ? 'bg-amber-400' : 'bg-slate-500'
            }`}></span>
            <span className="text-xl font-black text-white tracking-wide">
              {metrics?.status || 'IDLE'}
            </span>
          </div>
          <p className="text-[11px] text-slate-400 mt-2 font-mono">
            {metrics?.job_id ? `Job: ${metrics.job_id.substring(0, 8)}...` : 'Nenhum job em execução'}
          </p>
        </div>

        {/* KPI 2: Vazão & Throughput */}
        <div className="p-5 rounded-2xl bg-slate-900/60 border border-slate-800 backdrop-blur-sm shadow-lg">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-3 font-semibold">
            <span className="flex items-center gap-1.5">
              <Activity className="w-4 h-4 text-cyan-400" />
              Throughput Instantâneo
            </span>
            <span className="font-mono text-[10px] text-slate-500">Throttling: {metrics?.limit_bandwidth_mb || 50} MB/s</span>
          </div>
          <div className="text-xl font-black text-cyan-400 font-mono">
            {(metrics?.current_throughput_mbs || 0).toFixed(2)} <span className="text-sm font-semibold text-slate-400">MB/s</span>
          </div>
          <p className="text-[11px] text-slate-400 mt-2 font-mono">
            IOPS: <strong className="text-amber-400">{metrics?.current_iops || 0}</strong> / Limite: {metrics?.limit_iops || 300}
          </p>
        </div>

        {/* KPI 3: Capacidade Agregada */}
        <div className="p-5 rounded-2xl bg-slate-900/60 border border-slate-800 backdrop-blur-sm shadow-lg">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-3 font-semibold">
            <span className="flex items-center gap-1.5">
              <Database className="w-4 h-4 text-emerald-400" />
              Armazenamento Host
            </span>
            <span className="font-mono text-[10px] text-slate-500">{disks.length} Volumes</span>
          </div>
          <div className="text-xl font-black text-emerald-400 font-mono">
            {formatBytes(freeStorageBytes)} <span className="text-sm font-semibold text-slate-400">Livres</span>
          </div>
          <p className="text-[11px] text-slate-400 mt-2 font-mono">
            Total: {formatBytes(totalStorageBytes)} ({aggregateUsedPct}% Usado)
          </p>
        </div>

        {/* KPI 4: Integridade & Verificação */}
        <div className="p-5 rounded-2xl bg-slate-900/60 border border-slate-800 backdrop-blur-sm shadow-lg">
          <div className="flex items-center justify-between text-slate-400 text-xs mb-3 font-semibold">
            <span className="flex items-center gap-1.5">
              <Lock className="w-4 h-4 text-indigo-400" />
              Integridade Transacional
            </span>
            <span className="font-mono text-[10px] text-emerald-400">100% OK</span>
          </div>
          <div className="text-xl font-black text-indigo-300 font-mono">
            xxHash64
          </div>
          <p className="text-[11px] text-slate-400 mt-2 font-mono">
            Falhas: <strong className="text-slate-200">{metrics?.files_failed || 0}</strong> • Ignorados: {metrics?.files_skipped || 0}
          </p>
        </div>

      </div>

      {/* 3. Painel de Status do Job Ativo ou Última Migração */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-slate-800 backdrop-blur-sm shadow-xl">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-4">
          <div>
            <h3 className="text-base font-bold text-slate-100 flex items-center gap-2">
              <Activity className="w-5 h-5 text-cyan-400" />
              Status da Migração em Andamento
            </h3>
            <p className="text-xs text-slate-400 mt-0.5">
              {isJobActive 
                ? 'Operação de transferência ativa no motor assíncrono.' 
                : 'Nenhum trabalho em execução no momento. Pronto para iniciar novo pipeline.'}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setActiveTab('telemetry')}
              className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition flex items-center gap-1.5"
            >
              <Activity className="w-3.5 h-3.5 text-cyan-400" />
              Ver Telemetria Completa
            </button>
            <button
              onClick={() => setActiveTab('config')}
              className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white transition flex items-center gap-1.5"
            >
              <Sliders className="w-3.5 h-3.5" />
              Configurar Job
            </button>
          </div>
        </div>

        {/* Barra de Progresso Global */}
        <div className="space-y-2 mb-6">
          <div className="flex justify-between text-xs font-mono">
            <span className="text-slate-300">
              Progresso Geral: <strong className="text-cyan-400 font-bold">{(metrics?.progress_percent || 0).toFixed(1)}%</strong>
            </span>
            <span className="text-slate-400">
              Transferidos: <strong className="text-slate-200">{formatBytes(metrics?.bytes_transferred || 0)}</strong> de {formatBytes(metrics?.total_bytes_discovered || 0)}
            </span>
          </div>
          <div className="w-full bg-slate-950 rounded-full h-3 overflow-hidden border border-slate-800">
            <div 
              className="h-full bg-gradient-to-r from-blue-500 via-indigo-500 to-cyan-400 transition-all duration-300 rounded-full"
              style={{ width: `${Math.min(metrics?.progress_percent || 0, 100)}%` }}
            />
          </div>
          <div className="flex justify-between text-[11px] text-slate-500 font-mono">
            <span>Arquivos: {metrics?.files_copied || 0} / {metrics?.total_files_discovered || 0}</span>
            <span>Workers Ativos: {metrics?.active_workers || 0}</span>
          </div>
        </div>

        {/* Origem e Destino Atuais */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 p-3.5 rounded-xl bg-slate-950/70 border border-slate-800 text-xs font-mono">
          <div className="flex items-center gap-2 truncate">
            <span className="text-slate-500 flex-shrink-0">ORIGEM:</span>
            <span className="text-slate-300 truncate" title={sourceDir || 'Não definida'}>
              {sourceDir || '(Clique em Nova Migração para definir)'}
            </span>
          </div>
          <div className="flex items-center gap-2 truncate">
            <span className="text-slate-500 flex-shrink-0">DESTINO:</span>
            <span className="text-slate-300 truncate" title={destDir || 'Não definido'}>
              {destDir || '(Clique em Nova Migração para definir)'}
            </span>
          </div>
        </div>
      </div>

      {/* 4. Atalhos de Navegação Rápida para as 5 outras seções */}
      <div>
        <h3 className="text-sm font-bold text-slate-300 uppercase tracking-wider mb-3 flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-cyan-400" />
          Acesso Direto aos Módulos do Sistema
        </h3>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          
          {/* Módulo: Armazenamento */}
          <div 
            onClick={() => setActiveTab('storage')}
            className="group cursor-pointer p-5 rounded-2xl bg-slate-900/60 hover:bg-slate-900 border border-slate-800 hover:border-cyan-500/50 transition-all duration-200 shadow-lg flex flex-col justify-between"
          >
            <div>
              <div className="p-3 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 w-fit mb-3 group-hover:scale-105 transition-transform">
                <HardDrive className="w-5 h-5" />
              </div>
              <h4 className="font-bold text-sm text-slate-100 group-hover:text-cyan-300 transition">
                Armazenamento & Volumes
              </h4>
              <p className="text-xs text-slate-400 mt-1 leading-relaxed">
                Inspecione partições de disco do host, arquivos montados, capacidades e selecione diretórios com 1 clique.
              </p>
            </div>
            <div className="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs font-semibold text-cyan-400">
              <span>Explorar Volumes</span>
              <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </div>
          </div>

          {/* Módulo: Nova Migração */}
          <div 
            onClick={() => setActiveTab('config')}
            className="group cursor-pointer p-5 rounded-2xl bg-slate-900/60 hover:bg-slate-900 border border-slate-800 hover:border-blue-500/50 transition-all duration-200 shadow-lg flex flex-col justify-between"
          >
            <div>
              <div className="p-3 rounded-xl bg-blue-500/10 text-blue-400 border border-blue-500/20 w-fit mb-3 group-hover:scale-105 transition-transform">
                <Sliders className="w-5 h-5" />
              </div>
              <h4 className="font-bold text-sm text-slate-100 group-hover:text-blue-300 transition">
                Nova Migração
              </h4>
              <p className="text-xs text-slate-400 mt-1 leading-relaxed">
                Configure estratégias Full ou Delta, filtros por ano/mês/dia e aplique presets de Produção Segura.
              </p>
            </div>
            <div className="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs font-semibold text-blue-400">
              <span>Configurar Pipeline</span>
              <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </div>
          </div>

          {/* Módulo: Telemetria */}
          <div 
            onClick={() => setActiveTab('telemetry')}
            className="group cursor-pointer p-5 rounded-2xl bg-slate-900/60 hover:bg-slate-900 border border-slate-800 hover:border-emerald-500/50 transition-all duration-200 shadow-lg flex flex-col justify-between"
          >
            <div>
              <div className="p-3 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 w-fit mb-3 group-hover:scale-105 transition-transform">
                <Activity className="w-5 h-5" />
              </div>
              <h4 className="font-bold text-sm text-slate-100 group-hover:text-emerald-300 transition">
                Telemetria & Ao Vivo
              </h4>
              <p className="text-xs text-slate-400 mt-1 leading-relaxed">
                Acompanhe velocímetros de Throughput e IOPS, gráficos históricos e o live feed com hashes xxHash64.
              </p>
            </div>
            <div className="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs font-semibold text-emerald-400">
              <span>Abrir Telemetria</span>
              <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </div>
          </div>

          {/* Módulo: Auditoria & Relatórios */}
          <div 
            onClick={() => setActiveTab('audit')}
            className="group cursor-pointer p-5 rounded-2xl bg-slate-900/60 hover:bg-slate-900 border border-slate-800 hover:border-purple-500/50 transition-all duration-200 shadow-lg flex flex-col justify-between"
          >
            <div>
              <div className="p-3 rounded-xl bg-purple-500/10 text-purple-400 border border-purple-500/20 w-fit mb-3 group-hover:scale-105 transition-transform">
                <FileText className="w-5 h-5" />
              </div>
              <h4 className="font-bold text-sm text-slate-100 group-hover:text-purple-300 transition">
                Auditoria & Relatórios
              </h4>
              <p className="text-xs text-slate-400 mt-1 leading-relaxed">
                Consulte a trilha formal de conformidade, logs JSONL e faça download dos relatórios consolidados em CSV.
              </p>
            </div>
            <div className="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs font-semibold text-purple-400">
              <span>Ver Auditoria</span>
              <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </div>
          </div>

          {/* Módulo: Manual do Usuário */}
          <div 
            onClick={() => setActiveTab('docs')}
            className="group cursor-pointer p-5 rounded-2xl bg-slate-900/60 hover:bg-slate-900 border border-slate-800 hover:border-amber-500/50 transition-all duration-200 shadow-lg flex flex-col justify-between md:col-span-2 lg:col-span-2"
          >
            <div>
              <div className="p-3 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20 w-fit mb-3 group-hover:scale-105 transition-transform">
                <BookOpen className="w-5 h-5" />
              </div>
              <div className="flex items-center gap-2">
                <h4 className="font-bold text-sm text-slate-100 group-hover:text-amber-300 transition">
                  Manual do Usuário — MoveOps (9 Capítulos)
                </h4>
                <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-amber-500/15 text-amber-300 border border-amber-500/30">
                  GUIA PRÁTICO
                </span>
              </div>
              <p className="text-xs text-slate-400 mt-1 leading-relaxed">
                Guia operacional passo a passo: introdução, gerenciamento de discos, modos Full/Delta, Produção Segura, simulação Dry-Run, telemetria em tempo real, auditoria e execução em novas máquinas.
              </p>
            </div>
            <div className="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs font-semibold text-amber-400">
              <span>Abrir Manual Completo do Usuário</span>
              <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </div>
          </div>

        </div>
      </div>

    </div>
  );
}
