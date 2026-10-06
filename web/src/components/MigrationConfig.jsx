import React, { useState } from 'react';
import { 
  FolderInput, 
  FolderOutput, 
  Search, 
  Sliders, 
  Calendar, 
  ShieldAlert, 
  Zap, 
  Cpu, 
  FileSpreadsheet, 
  Play, 
  Eye, 
  Filter, 
  Settings2,
  Check,
  AlertCircle,
  ShieldCheck,
  Network,
  HardDrive,
  Layers,
  Sun,
  Moon,
  Flame
} from 'lucide-react';
import { api } from '../services/api';

const PRODUCTION_PRESETS = [
  {
    id: 'prod_standard',
    name: 'Produção Padrão',
    badge: 'Recomendado',
    icon: ShieldCheck,
    bandwidth: 50,
    iops: 300,
    concurrency: 4,
    headline: '50 MB/s • 300 IOPS • 4 Workers',
    desc: 'Segurança balanceada',
    detail: 'Proteção ativa contra sobrecarga de switches e fila de disco sem impacto operacional.',
    cardBorder: 'border-emerald-500/50 hover:border-emerald-400',
    activeStyle: 'bg-emerald-950/40 border-emerald-500 text-emerald-300 ring-2 ring-emerald-500/40 shadow-lg shadow-emerald-950/50',
    badgeClass: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40',
    badgeRecommend: true,
  },
  {
    id: 'business_hours',
    name: 'Horário Comercial',
    badge: 'Conservador',
    icon: Sun,
    bandwidth: 25,
    iops: 150,
    concurrency: 2,
    headline: '25 MB/s • 150 IOPS • 2 Workers',
    desc: 'Mínimo impacto operacional',
    detail: 'Carga ultraleve projetada para não competir com tráfego de usuários ou ERPs.',
    cardBorder: 'border-amber-500/50 hover:border-amber-400',
    activeStyle: 'bg-amber-950/40 border-amber-500 text-amber-300 ring-2 ring-amber-500/40 shadow-lg shadow-amber-950/50',
    badgeClass: 'bg-amber-500/20 text-amber-300 border-amber-500/40',
    badgeRecommend: false,
  },
  {
    id: 'night_window',
    name: 'Janela Noturna',
    badge: 'Alta Vazão',
    icon: Moon,
    bandwidth: 150,
    iops: 1500,
    concurrency: 8,
    headline: '150 MB/s • 1500 IOPS • 8 Workers',
    desc: 'Fora de expediente',
    detail: 'Acelera a sincronização durante janelas noturnas aproveitando a ociosidade da infra.',
    cardBorder: 'border-blue-500/50 hover:border-blue-400',
    activeStyle: 'bg-blue-950/40 border-blue-500 text-blue-300 ring-2 ring-blue-500/40 shadow-lg shadow-blue-950/50',
    badgeClass: 'bg-blue-500/20 text-blue-300 border-blue-500/40',
    badgeRecommend: false,
  },
  {
    id: 'maintenance',
    name: 'Dedicado / Manutenção',
    badge: 'Atenção',
    icon: Flame,
    bandwidth: 0,
    iops: 0,
    concurrency: 16,
    headline: 'Ilimitado • Ilimitado • 16 Workers',
    desc: 'Vazão máxima sem limites',
    detail: 'Para janelas exclusivas de parada. Pode saturar switches e elevar fila de disco.',
    cardBorder: 'border-rose-500/50 hover:border-rose-400',
    activeStyle: 'bg-rose-950/40 border-rose-500 text-rose-300 ring-2 ring-rose-500/40 shadow-lg shadow-rose-950/50',
    badgeClass: 'bg-rose-500/20 text-rose-300 border-rose-500/40',
    badgeRecommend: false,
  },
];

export function MigrationConfig({
  sourceDir,
  setSourceDir,
  destDir,
  setDestDir,
  onOpenBrowser,
  onOpenPreview,
  onStartMigration,
  isJobRunning,
  bandwidthLimit,
  setBandwidthLimit,
  iopsLimit,
  setIopsLimit,
  concurrency,
  setConcurrency,
  enableAudit,
  setEnableAudit,
  auditDir,
  setAuditDir,
  filterMode,
  setFilterMode,
  yearFilter,
  setYearFilter,
  monthFilter,
  setMonthFilter,
  dayFilter,
  setDayFilter,
  startDateFilter,
  setStartDateFilter,
  endDateFilter,
  setEndDateFilter,
  olderThanDays,
  setOlderThanDays,
  includePatterns,
  setIncludePatterns,
  excludePatterns,
  setExcludePatterns,
  onSwitchToStorage,
}) {
  const [estimating, setEstimating] = useState(false);
  const [estimateError, setEstimateError] = useState(null);

  const currentBw = parseFloat(bandwidthLimit) || 0;
  const currentIops = parseInt(iopsLimit, 10) || 0;
  const currentConc = parseInt(concurrency, 10) || 4;

  const activePreset = PRODUCTION_PRESETS.find(
    (p) => p.bandwidth === currentBw && p.iops === currentIops && p.concurrency === currentConc
  );

  const isSafeProduction = currentBw > 0 || currentIops > 0;

  const buildFilterPayload = () => {
    const filters = {
      mode: filterMode,
    };

    if (filterMode === 'BY_YEAR') {
      filters.year = parseInt(yearFilter, 10);
    } else if (filterMode === 'BY_MONTH') {
      filters.year = parseInt(yearFilter, 10);
      filters.month = parseInt(monthFilter, 10);
    } else if (filterMode === 'BY_DAY' && dayFilter) {
      filters.target_day = new Date(dayFilter).toISOString();
    } else if (filterMode === 'DATE_RANGE') {
      if (startDateFilter) filters.start_date = new Date(startDateFilter).toISOString();
      if (endDateFilter) filters.end_date = new Date(endDateFilter).toISOString();
    } else if (filterMode === 'RETENTION') {
      filters.older_than_days = parseInt(olderThanDays, 10) || 30;
    }

    if (includePatterns.trim()) {
      filters.include_patterns = includePatterns
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean);
    }
    if (excludePatterns.trim()) {
      filters.exclude_patterns = excludePatterns
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean);
    }

    return filters;
  };

  const handlePreviewClick = async () => {
    if (!sourceDir) {
      setEstimateError('Por favor, informe a pasta de Origem.');
      return;
    }
    try {
      setEstimating(true);
      setEstimateError(null);
      const filters = buildFilterPayload();
      const summary = await api.preview({
        source_dir: sourceDir,
        destination_dir: destDir || '',
        filters,
      });
      onOpenPreview(summary);
    } catch (err) {
      setEstimateError(err.message || 'Falha ao executar cálculo de estimativa');
    } finally {
      setEstimating(false);
    }
  };

  const handleStartClick = () => {
    if (!sourceDir || !destDir) {
      setEstimateError('Origem e Destino são obrigatórios.');
      return;
    }
    setEstimateError(null);
    const filters = buildFilterPayload();
    onStartMigration({
      source_dir: sourceDir,
      destination_dir: destDir,
      filters,
      max_bandwidth_mb: parseFloat(bandwidthLimit) || 0,
      max_iops: parseInt(iopsLimit, 10) || 0,
      concurrency: parseInt(concurrency, 10) || 4,
      audit_dir: enableAudit ? auditDir : '',
    });
  };

  const currentYear = new Date().getFullYear();
  const yearsList = Array.from({ length: 15 }, (_, i) => currentYear - i);

  return (
    <div className="space-y-6">
      
      {/* Bloco 1: Seleção de Diretórios Origem e Destino */}
      <div className="bg-slate-900/60 rounded-2xl border border-slate-800 p-6 backdrop-blur-sm shadow-xl">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
          <h3 className="text-base font-bold text-slate-100 flex items-center gap-2">
            <Settings2 className="w-5 h-5 text-blue-400" />
            Mapeamento de Pastas (Origem & Destino)
          </h3>
          {onSwitchToStorage && (
            <button
              type="button"
              onClick={onSwitchToStorage}
              className="text-xs font-semibold text-cyan-400 hover:text-cyan-300 flex items-center gap-1.5 transition px-3 py-1.5 rounded-lg bg-cyan-950/40 hover:bg-cyan-900/40 border border-cyan-500/25 self-start sm:self-auto"
            >
              <HardDrive className="w-3.5 h-3.5" />
              <span>Inspecionar Discos & Volumes do Host</span>
            </button>
          )}
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Origem */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold text-slate-300 flex items-center justify-between">
              <span className="flex items-center gap-1.5">
                <FolderInput className="w-4 h-4 text-blue-400" />
                Diretório de Origem (Source)
              </span>
              <span className="text-[11px] text-slate-500 font-mono">Leitura nativa</span>
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={sourceDir}
                onChange={(e) => setSourceDir(e.target.value)}
                placeholder="/dados/antigos ou C:\Dados"
                disabled={isJobRunning}
                className="flex-1 bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2.5 text-xs font-mono text-slate-200 focus:outline-none focus:border-blue-500 disabled:opacity-50"
              />
              <button
                type="button"
                onClick={() => onOpenBrowser('source')}
                disabled={isJobRunning}
                className="px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition disabled:opacity-50"
              >
                <Search className="w-3.5 h-3.5" />
                Explorar
              </button>
            </div>
            <p className="text-[11px] text-slate-500">
              Caminho no servidor de onde os arquivos serão lidos.
            </p>
          </div>

          {/* Destino */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold text-slate-300 flex items-center justify-between">
              <span className="flex items-center gap-1.5">
                <FolderOutput className="w-4 h-4 text-emerald-400" />
                Diretório de Destino (Destination)
              </span>
              <span className="text-[11px] text-slate-500 font-mono">Escrita garantida</span>
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={destDir}
                onChange={(e) => setDestDir(e.target.value)}
                placeholder="/mnt/novo_storage ou D:\Storage"
                disabled={isJobRunning}
                className="flex-1 bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2.5 text-xs font-mono text-slate-200 focus:outline-none focus:border-emerald-500 disabled:opacity-50"
              />
              <button
                type="button"
                onClick={() => onOpenBrowser('dest')}
                disabled={isJobRunning}
                className="px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold flex items-center gap-1.5 transition disabled:opacity-50"
              >
                <Search className="w-3.5 h-3.5" />
                Explorar
              </button>
            </div>
            <p className="text-[11px] text-slate-500">
              Caminho para onde os dados serão copiados e verificados com hash.
            </p>
          </div>
        </div>
      </div>

      {/* Bloco 2: Modo de Migração & Filtro Temporal */}
      <div className="bg-slate-900/60 rounded-2xl border border-slate-800 p-6 backdrop-blur-sm shadow-xl">
        <h3 className="text-base font-bold text-slate-100 flex items-center gap-2 mb-4">
          <Filter className="w-5 h-5 text-indigo-400" />
          Estratégia de Sincronização & Filtro Temporal
        </h3>

        {/* Seleção de Modo */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 mb-6">
          
          <label className={`cursor-pointer p-4 rounded-xl border transition flex flex-col justify-between ${
            filterMode === 'FULL'
              ? 'bg-blue-950/40 border-blue-500 text-white ring-1 ring-blue-500/30'
              : 'bg-slate-950/60 border-slate-800 text-slate-400 hover:border-slate-700 hover:text-slate-200'
          }`}>
            <div className="flex items-start justify-between">
              <div>
                <input
                  type="radio"
                  name="filterMode"
                  value="FULL"
                  checked={filterMode === 'FULL'}
                  onChange={() => setFilterMode('FULL')}
                  className="hidden"
                />
                <span className="font-bold text-sm block text-slate-100">Completa (Full Sync)</span>
                <span className="text-xs text-slate-400 mt-1 block">
                  Copia todos os arquivos da pasta de origem.
                </span>
              </div>
              <div className={`w-5 h-5 rounded-full border flex items-center justify-center ${
                filterMode === 'FULL' ? 'border-blue-400 bg-blue-500 text-white' : 'border-slate-700'
              }`}>
                {filterMode === 'FULL' && <Check className="w-3 h-3" />}
              </div>
            </div>
          </label>

          <label className={`cursor-pointer p-4 rounded-xl border transition flex flex-col justify-between ${
            filterMode === 'DELTA_ONLY'
              ? 'bg-emerald-950/40 border-emerald-500 text-white ring-1 ring-emerald-500/30'
              : 'bg-slate-950/60 border-slate-800 text-slate-400 hover:border-slate-700 hover:text-slate-200'
          }`}>
            <div className="flex items-start justify-between">
              <div>
                <input
                  type="radio"
                  name="filterMode"
                  value="DELTA_ONLY"
                  checked={filterMode === 'DELTA_ONLY'}
                  onChange={() => setFilterMode('DELTA_ONLY')}
                  className="hidden"
                />
                <span className="font-bold text-sm block text-slate-100">Apenas Diferenças (Delta)</span>
                <span className="text-xs text-slate-400 mt-1 block">
                  Transfere somente arquivos modificados ou novos.
                </span>
              </div>
              <div className={`w-5 h-5 rounded-full border flex items-center justify-center ${
                filterMode === 'DELTA_ONLY' ? 'border-emerald-400 bg-emerald-500 text-white' : 'border-slate-700'
              }`}>
                {filterMode === 'DELTA_ONLY' && <Check className="w-3 h-3" />}
              </div>
            </div>
          </label>

          <label className={`cursor-pointer p-4 rounded-xl border transition flex flex-col justify-between ${
            ['DATE_RANGE', 'BY_YEAR', 'BY_MONTH', 'BY_DAY', 'RETENTION'].includes(filterMode)
              ? 'bg-purple-950/40 border-purple-500 text-white ring-1 ring-purple-500/30'
              : 'bg-slate-950/60 border-slate-800 text-slate-400 hover:border-slate-700 hover:text-slate-200'
          }`}>
            <div className="flex items-start justify-between">
              <div>
                <input
                  type="radio"
                  name="filterMode"
                  value="DATE_RANGE"
                  checked={['DATE_RANGE', 'BY_YEAR', 'BY_MONTH', 'BY_DAY', 'RETENTION'].includes(filterMode)}
                  onChange={() => setFilterMode('DATE_RANGE')}
                  className="hidden"
                />
                <span className="font-bold text-sm block text-slate-100">Filtro Temporal</span>
                <span className="text-xs text-slate-400 mt-1 block">
                  Por intervalo de datas, ano, mês ou retenção.
                </span>
              </div>
              <div className={`w-5 h-5 rounded-full border flex items-center justify-center ${
                ['DATE_RANGE', 'BY_YEAR', 'BY_MONTH', 'BY_DAY', 'RETENTION'].includes(filterMode)
                  ? 'border-purple-400 bg-purple-500 text-white'
                  : 'border-slate-700'
              }`}>
                {['DATE_RANGE', 'BY_YEAR', 'BY_MONTH', 'BY_DAY', 'RETENTION'].includes(filterMode) && <Check className="w-3 h-3" />}
              </div>
            </div>
          </label>

        </div>

        {/* Painel Específico do Filtro Temporal */}
        {['DATE_RANGE', 'BY_YEAR', 'BY_MONTH', 'BY_DAY', 'RETENTION'].includes(filterMode) && (
          <div className="p-4 rounded-xl bg-slate-950/80 border border-purple-500/30 space-y-4 animate-in fade-in duration-150">
            <div className="flex flex-wrap items-center gap-2 border-b border-slate-800 pb-3">
              <span className="text-xs font-semibold text-purple-300 mr-2 flex items-center gap-1.5">
                <Calendar className="w-4 h-4" /> Tipo de Seletor:
              </span>
              {[
                { id: 'DATE_RANGE', label: 'Intervalo de Datas [Início - Fim]' },
                { id: 'BY_YEAR', label: 'Por Ano Específico' },
                { id: 'BY_MONTH', label: 'Por Mês/Ano' },
                { id: 'BY_DAY', label: 'Por Dia Específico' },
                { id: 'RETENTION', label: 'Mais Antigos que X Dias' },
              ].map((sub) => (
                <button
                  key={sub.id}
                  type="button"
                  onClick={() => setFilterMode(sub.id)}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium transition ${
                    filterMode === sub.id
                      ? 'bg-purple-600 text-white shadow-md'
                      : 'bg-slate-900 text-slate-400 hover:text-slate-200 border border-slate-800'
                  }`}
                >
                  {sub.label}
                </button>
              ))}
            </div>

            {/* Inputs Dinâmicos do Filtro */}
            {filterMode === 'DATE_RANGE' && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-mono text-slate-400 mb-1">Data de Início</label>
                  <input
                    type="date"
                    value={startDateFilter}
                    onChange={(e) => setStartDateFilter(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-purple-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-mono text-slate-400 mb-1">Data de Fim</label>
                  <input
                    type="date"
                    value={endDateFilter}
                    onChange={(e) => setEndDateFilter(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-purple-500"
                  />
                </div>
              </div>
            )}

            {filterMode === 'BY_YEAR' && (
              <div className="max-w-xs">
                <label className="block text-xs font-mono text-slate-400 mb-1">Ano Desejado</label>
                <select
                  value={yearFilter}
                  onChange={(e) => setYearFilter(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-purple-500"
                >
                  {yearsList.map((y) => (
                    <option key={y} value={y}>{y}</option>
                  ))}
                </select>
              </div>
            )}

            {filterMode === 'BY_MONTH' && (
              <div className="grid grid-cols-2 gap-4 max-w-md">
                <div>
                  <label className="block text-xs font-mono text-slate-400 mb-1">Mês</label>
                  <select
                    value={monthFilter}
                    onChange={(e) => setMonthFilter(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-purple-500"
                  >
                    {[
                      { num: 1, name: '01 - Janeiro' },
                      { num: 2, name: '02 - Fevereiro' },
                      { num: 3, name: '03 - Março' },
                      { num: 4, name: '04 - Abril' },
                      { num: 5, name: '05 - Maio' },
                      { num: 6, name: '06 - Junho' },
                      { num: 7, name: '07 - Julho' },
                      { num: 8, name: '08 - Agosto' },
                      { num: 9, name: '09 - Setembro' },
                      { num: 10, name: '10 - Outubro' },
                      { num: 11, name: '11 - Novembro' },
                      { num: 12, name: '12 - Dezembro' },
                    ].map((m) => (
                      <option key={m.num} value={m.num}>{m.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-mono text-slate-400 mb-1">Ano</label>
                  <select
                    value={yearFilter}
                    onChange={(e) => setYearFilter(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-purple-500"
                  >
                    {yearsList.map((y) => (
                      <option key={y} value={y}>{y}</option>
                    ))}
                  </select>
                </div>
              </div>
            )}

            {filterMode === 'BY_DAY' && (
              <div className="max-w-xs">
                <label className="block text-xs font-mono text-slate-400 mb-1">Dia Específico</label>
                <input
                  type="date"
                  value={dayFilter}
                  onChange={(e) => setDayFilter(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-purple-500"
                />
              </div>
            )}

            {filterMode === 'RETENTION' && (
              <div className="max-w-xs">
                <label className="block text-xs font-mono text-slate-400 mb-1">
                  Arquivos com mais de (Dias):
                </label>
                <input
                  type="number"
                  min="1"
                  value={olderThanDays}
                  onChange={(e) => setOlderThanDays(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-purple-500 font-mono"
                />
              </div>
            )}
          </div>
        )}

        {/* Padrões de Inclusão e Exclusão */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4 pt-4 border-t border-slate-800">
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">
              Padrões para Incluir (Include Patterns)
            </label>
            <input
              type="text"
              value={includePatterns}
              onChange={(e) => setIncludePatterns(e.target.value)}
              placeholder="*.pdf, *.docx, *.xlsx (vazio = todos)"
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs font-mono text-slate-200 focus:outline-none focus:border-indigo-500"
            />
            <span className="text-[11px] text-slate-500">Separados por vírgula.</span>
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">
              Padrões para Ignorar (Exclude Patterns)
            </label>
            <input
              type="text"
              value={excludePatterns}
              onChange={(e) => setExcludePatterns(e.target.value)}
              placeholder="*.tmp, ~*, node_modules/**, .git/**"
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs font-mono text-slate-200 focus:outline-none focus:border-indigo-500"
            />
            <span className="text-[11px] text-slate-500">Arquivos temporários e pastas de build.</span>
          </div>
        </div>
      </div>

      {/* Bloco 3: Throttling & Produção Segura (Rate Limits e Concorrência) */}
      <div className="bg-slate-900/60 rounded-2xl border border-slate-800 p-6 backdrop-blur-sm shadow-xl">
        
        {/* Cabeçalho do Bloco */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-5">
          <div>
            <h3 className="text-base font-bold text-slate-100 flex items-center gap-2">
              <Zap className="w-5 h-5 text-amber-400" />
              Controle de Throttling (Produção Segura)
            </h3>
            <p className="text-xs text-slate-400 mt-0.5">
              Proteção ativa para evitar degradação de serviços em produção, hosts e infraestrutura de rede.
            </p>
          </div>
          <div>
            {isSafeProduction ? (
              <span className="px-3 py-1.5 rounded-full text-xs font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 flex items-center gap-1.5 shadow-sm">
                <ShieldCheck className="w-4 h-4 text-emerald-400" />
                Modo Produção Segura (Padrão Ativo)
              </span>
            ) : (
              <span className="px-3 py-1.5 rounded-full text-xs font-semibold bg-rose-500/15 text-rose-400 border border-rose-500/30 flex items-center gap-1.5 shadow-sm">
                <ShieldAlert className="w-4 h-4 text-rose-400" />
                Modo Ilimitado (Atenção ao Storage)
              </span>
            )}
          </div>
        </div>

        {/* Destaque Visual do Modo Produção Segura e Proteção Simultânea */}
        {isSafeProduction ? (
          <div className="mb-6 p-4 rounded-xl bg-gradient-to-r from-emerald-950/40 via-cyan-950/20 to-slate-950 border border-emerald-500/40 flex flex-col md:flex-row md:items-center justify-between gap-4 shadow-lg">
            <div className="flex items-start gap-3.5">
              <div className="p-2.5 rounded-xl bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 flex-shrink-0 mt-0.5">
                <ShieldCheck className="w-6 h-6" />
              </div>
              <div className="space-y-1">
                <div className="flex flex-wrap items-center gap-2">
                  <h4 className="text-sm font-bold text-emerald-300 tracking-wide uppercase">
                    Modo Produção Segura (Padrão Ativo)
                  </h4>
                  <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-emerald-500/20 text-emerald-300 border border-emerald-500/40">
                    PROTEÇÃO TRIPLA SIMULTÂNEA
                  </span>
                </div>
                <p className="text-xs text-slate-300 leading-relaxed">
                  Garante a proteção simultânea do <strong className="text-white">Host</strong> (latência baixa e filas de disco protegidas), da <strong className="text-white">Rede</strong> (sem saturação de canais SAN/NAS, switches ou links WAN) e da <strong className="text-white">Cópia</strong> (cadência balanceada com buffers controlados e hashing xxHash64).
                </p>
              </div>
            </div>
            <div className="flex-shrink-0 self-start md:self-center px-3.5 py-2 rounded-xl bg-slate-900/90 border border-emerald-500/30 text-xs font-mono text-emerald-300 flex items-center gap-2">
              <span className="inline-block w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
              <span>{currentBw > 0 ? `${currentBw} MB/s` : 'Banda Livre'}</span>
              <span className="text-slate-600">•</span>
              <span>{currentIops > 0 ? `${currentIops} IOPS` : 'IOPS Livre'}</span>
              <span className="text-slate-600">•</span>
              <span>{currentConc} Threads</span>
            </div>
          </div>
        ) : (
          <div className="mb-6 p-4 rounded-xl bg-rose-950/30 border border-rose-500/40 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 shadow-lg">
            <div className="flex items-start gap-3">
              <div className="p-2.5 rounded-xl bg-rose-500/20 text-rose-400 border border-rose-500/30 flex-shrink-0 mt-0.5">
                <ShieldAlert className="w-5 h-5" />
              </div>
              <div>
                <h4 className="text-sm font-bold text-rose-300 uppercase">
                  Atenção: Modo Ilimitado (Sem Throttling de Produção)
                </h4>
                <p className="text-xs text-rose-200/90 mt-0.5">
                  Os limites de vazão e IOPS estão desativados. Isso pode consumir toda a largura de banda de switches e elevar filas de I/O em discos compartilhados. Use apenas em manutenções programadas.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Indicadores dos 3 Pilares de Proteção */}
        <div className="mb-6 grid grid-cols-1 md:grid-cols-3 gap-3">
          <div className="p-3.5 rounded-xl bg-slate-950/60 border border-slate-800 flex items-start gap-3">
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-400 border border-blue-500/20 flex-shrink-0 mt-0.5">
              <Network className="w-4 h-4" />
            </div>
            <div>
              <div className="flex items-center justify-between gap-1">
                <h5 className="text-xs font-bold text-slate-200">Segurança de Rede</h5>
                <span className="text-[10px] text-blue-400 font-mono">Padrão: 50 MB/s</span>
              </div>
              <p className="text-[11px] text-slate-400 mt-1 leading-relaxed">
                Evita saturação de canais SAN/NAS, switches ou links WAN/VPN corporativos.
              </p>
            </div>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-950/60 border border-slate-800 flex items-start gap-3">
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-400 border border-amber-500/20 flex-shrink-0 mt-0.5">
              <HardDrive className="w-4 h-4" />
            </div>
            <div>
              <div className="flex items-center justify-between gap-1">
                <h5 className="text-xs font-bold text-slate-200">Segurança do Host</h5>
                <span className="text-[10px] text-amber-400 font-mono">Padrão: 300 IOPS</span>
              </div>
              <p className="text-[11px] text-slate-400 mt-1 leading-relaxed">
                Evita saturação da controladora e filas de disco para aplicações concorrentes.
              </p>
            </div>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-950/60 border border-slate-800 flex items-start gap-3">
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex-shrink-0 mt-0.5">
              <Layers className="w-4 h-4" />
            </div>
            <div>
              <div className="flex items-center justify-between gap-1">
                <h5 className="text-xs font-bold text-slate-200">Segurança da Cópia</h5>
                <span className="text-[10px] text-emerald-400 font-mono">Padrão: 4 Workers</span>
              </div>
              <p className="text-[11px] text-slate-400 mt-1 leading-relaxed">
                Fluxo cadenciado com buffers gerenciados e integridade xxHash64 garantida.
              </p>
            </div>
          </div>
        </div>

        {/* Seletores Rápidos com Botões: Presets de Produção Segura */}
        <div className="mb-6 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center gap-1.5">
              <Sliders className="w-3.5 h-3.5 text-cyan-400" />
              Presets de Produção Segura
            </span>
            <span className="text-[11px] text-slate-500 font-mono">
              Clique para carregar o perfil
            </span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            {PRODUCTION_PRESETS.map((preset) => {
              const Icon = preset.icon;
              const isSelected = activePreset?.id === preset.id;
              return (
                <button
                  key={preset.id}
                  type="button"
                  disabled={isJobRunning}
                  onClick={() => {
                    setBandwidthLimit(preset.bandwidth);
                    setIopsLimit(preset.iops);
                    setConcurrency(preset.concurrency);
                  }}
                  className={`text-left p-3.5 rounded-xl border transition relative flex flex-col justify-between ${
                    isSelected
                      ? preset.activeStyle
                      : `bg-slate-950/70 border-slate-800 text-slate-400 hover:border-slate-700 hover:text-slate-200 ${preset.cardBorder}`
                  } disabled:opacity-50`}
                >
                  <div>
                    <div className="flex items-center justify-between gap-1 mb-2">
                      <span className="flex items-center gap-1.5 font-bold text-xs text-slate-100">
                        <Icon className="w-4 h-4 flex-shrink-0" />
                        {preset.name}
                      </span>
                      <span className={`text-[10px] font-mono px-1.5 py-0.5 rounded border ${preset.badgeClass}`}>
                        {preset.badge}
                      </span>
                    </div>
                    <div className="text-[11px] font-mono text-cyan-400 font-semibold mb-1">
                      {preset.headline}
                    </div>
                    <p className="text-[11px] text-slate-300 mb-1">
                      {preset.desc}
                    </p>
                    <p className="text-[10px] text-slate-400 line-clamp-2 leading-tight">
                      {preset.detail}
                    </p>
                  </div>

                  <div className="mt-3 pt-2 border-t border-slate-800/80 flex items-center justify-between">
                    {isSelected ? (
                      <span className="flex items-center gap-1 text-[11px] font-bold text-emerald-400">
                        <Check className="w-3.5 h-3.5" /> Ativo
                      </span>
                    ) : (
                      <span className="text-[10px] text-slate-500 hover:text-slate-300">
                        Aplicar Perfil
                      </span>
                    )}
                    {preset.badgeRecommend && (
                      <span className="text-[9px] uppercase tracking-wider font-bold text-emerald-400/90 bg-emerald-500/10 px-1.5 py-0.5 rounded">
                        Padrão
                      </span>
                    )}
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* Ajuste Fino dos Parâmetros */}
        <div className="pt-4 border-t border-slate-800/80">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
              Ajuste Fino dos Parâmetros
            </span>
            <span className="text-[11px] text-slate-500 font-mono">
              Valores manuais contínuos
            </span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            
            {/* Limite de Banda (MB/s) */}
            <div className="space-y-2 p-4 rounded-xl bg-slate-950/70 border border-slate-800">
              <div className="flex justify-between items-center text-xs">
                <div>
                  <span className="font-semibold text-slate-300 block">Limite de Banda</span>
                  <span className="text-[10px] text-slate-500 font-mono">Padrão Seguro: 50 MB/s</span>
                </div>
                <span className="font-mono font-bold text-cyan-400">
                  {parseFloat(bandwidthLimit) === 0 ? 'ILIMITADO' : `${bandwidthLimit} MB/s`}
                </span>
              </div>
              <input
                type="range"
                min="0"
                max="1000"
                step="5"
                value={bandwidthLimit}
                onChange={(e) => setBandwidthLimit(e.target.value)}
                disabled={isJobRunning}
                className="w-full accent-cyan-400 cursor-pointer disabled:opacity-50"
              />
              <div className="flex justify-between text-[10px] text-slate-500 font-mono">
                <span>0 (Max)</span>
                <span className="text-cyan-400 font-bold">50 (Padrão)</span>
                <span>250</span>
                <span>500</span>
                <span>1000</span>
              </div>
              <div className="flex items-center gap-2 mt-2">
                <input
                  type="number"
                  min="0"
                  value={bandwidthLimit}
                  onChange={(e) => setBandwidthLimit(e.target.value)}
                  disabled={isJobRunning}
                  placeholder="0 = Sem limite"
                  className="flex-1 bg-slate-900 border border-slate-800 rounded-lg px-2.5 py-1 text-xs font-mono text-slate-200 disabled:opacity-50"
                />
                <button
                  type="button"
                  onClick={() => setBandwidthLimit(50)}
                  disabled={isJobRunning}
                  className="px-2 py-1 text-[10px] font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 rounded border border-slate-700 disabled:opacity-50"
                >
                  Reset 50
                </button>
              </div>
            </div>

            {/* Limite de IOPS */}
            <div className="space-y-2 p-4 rounded-xl bg-slate-950/70 border border-slate-800">
              <div className="flex justify-between items-center text-xs">
                <div>
                  <span className="font-semibold text-slate-300 block">Limite de IOPS</span>
                  <span className="text-[10px] text-slate-500 font-mono">Padrão Seguro: 300 IOPS</span>
                </div>
                <span className="font-mono font-bold text-amber-400">
                  {parseInt(iopsLimit, 10) === 0 ? 'ILIMITADO' : `${iopsLimit} IOPS`}
                </span>
              </div>
              <input
                type="range"
                min="0"
                max="5000"
                step="25"
                value={iopsLimit}
                onChange={(e) => setIopsLimit(e.target.value)}
                disabled={isJobRunning}
                className="w-full accent-amber-400 cursor-pointer disabled:opacity-50"
              />
              <div className="flex justify-between text-[10px] text-slate-500 font-mono">
                <span>0 (Max)</span>
                <span className="text-amber-400 font-bold">300 (Padrão)</span>
                <span>1.000</span>
                <span>2.500</span>
                <span>5.000</span>
              </div>
              <div className="flex items-center gap-2 mt-2">
                <input
                  type="number"
                  min="0"
                  value={iopsLimit}
                  onChange={(e) => setIopsLimit(e.target.value)}
                  disabled={isJobRunning}
                  placeholder="0 = Sem limite"
                  className="flex-1 bg-slate-900 border border-slate-800 rounded-lg px-2.5 py-1 text-xs font-mono text-slate-200 disabled:opacity-50"
                />
                <button
                  type="button"
                  onClick={() => setIopsLimit(300)}
                  disabled={isJobRunning}
                  className="px-2 py-1 text-[10px] font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 rounded border border-slate-700 disabled:opacity-50"
                >
                  Reset 300
                </button>
              </div>
            </div>

            {/* Concorrência de Trabalhadores */}
            <div className="space-y-2 p-4 rounded-xl bg-slate-950/70 border border-slate-800">
              <div className="flex justify-between items-center text-xs">
                <div>
                  <span className="font-semibold text-slate-300 block">Workers Concorrentes</span>
                  <span className="text-[10px] text-slate-500 font-mono">Padrão Seguro: 4 Threads</span>
                </div>
                <span className="font-mono font-bold text-emerald-400">
                  {concurrency} Threads
                </span>
              </div>
              <input
                type="range"
                min="1"
                max="32"
                value={concurrency}
                onChange={(e) => setConcurrency(e.target.value)}
                disabled={isJobRunning}
                className="w-full accent-emerald-400 cursor-pointer disabled:opacity-50"
              />
              <div className="flex justify-between text-[10px] text-slate-500 font-mono">
                <span>1</span>
                <span className="text-emerald-400 font-bold">4 (Padrão)</span>
                <span>8</span>
                <span>16</span>
                <span>32</span>
              </div>
              <div className="flex items-center gap-2 mt-2">
                <input
                  type="number"
                  min="1"
                  max="64"
                  value={concurrency}
                  onChange={(e) => setConcurrency(e.target.value)}
                  disabled={isJobRunning}
                  className="flex-1 bg-slate-900 border border-slate-800 rounded-lg px-2.5 py-1 text-xs font-mono text-slate-200 disabled:opacity-50"
                />
                <button
                  type="button"
                  onClick={() => setConcurrency(4)}
                  disabled={isJobRunning}
                  className="px-2 py-1 text-[10px] font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 rounded border border-slate-700 disabled:opacity-50"
                >
                  Reset 4
                </button>
              </div>
            </div>

          </div>
        </div>
      </div>

      {/* Bloco 4: Opções de Auditoria */}
      <div className="bg-slate-900/60 rounded-2xl border border-slate-800 p-6 backdrop-blur-sm shadow-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="auditToggle"
            checked={enableAudit}
            onChange={(e) => setEnableAudit(e.target.checked)}
            className="w-5 h-5 rounded border-slate-700 bg-slate-950 text-cyan-500 focus:ring-cyan-500"
          />
          <label htmlFor="auditToggle" className="cursor-pointer">
            <span className="text-sm font-bold text-slate-200 block">
              Habilitar Relatório Detalhado e Sumarizado de Auditoria
            </span>
            <span className="text-xs text-slate-400 block">
              Gera trilha estruturada de conformidade (JSONL e CSV) com hashes xxHash64 de cada arquivo.
            </span>
          </label>
        </div>

        {enableAudit && (
          <div className="flex items-center gap-2 text-xs font-mono">
            <span className="text-slate-400">Pasta de Logs:</span>
            <input
              type="text"
              value={auditDir}
              onChange={(e) => setAuditDir(e.target.value)}
              className="bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1 text-xs font-mono text-slate-300 w-36"
            />
          </div>
        )}
      </div>

      {/* Alerta de Erro de Validação */}
      {estimateError && (
        <div className="p-4 rounded-xl bg-rose-500/10 border border-rose-500/30 flex items-center gap-2 text-rose-300 text-xs">
          <AlertCircle className="w-4 h-4 flex-shrink-0" />
          <span>{estimateError}</span>
        </div>
      )}

      {/* Ações Finais: Estimar Preview e Iniciar Migração */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-4 p-4 rounded-2xl bg-gradient-to-r from-slate-900 via-slate-900 to-slate-950 border border-slate-800">
        <button
          type="button"
          onClick={handlePreviewClick}
          disabled={estimating || isJobRunning}
          className="w-full sm:w-auto px-5 py-3 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-bold flex items-center justify-center gap-2 border border-slate-700 transition disabled:opacity-50"
        >
          <Eye className={`w-4 h-4 ${estimating ? 'animate-pulse text-cyan-400' : ''}`} />
          {estimating ? 'Escaneando & Estimando...' : 'Estimar / Pré-visualizar (Dry Run)'}
        </button>

        <button
          type="button"
          onClick={handleStartClick}
          disabled={isJobRunning}
          className="w-full sm:w-auto px-8 py-3 rounded-xl bg-gradient-to-r from-blue-600 via-indigo-600 to-cyan-600 hover:from-blue-500 hover:via-indigo-500 hover:to-cyan-500 text-white text-xs font-extrabold flex items-center justify-center gap-2.5 shadow-xl shadow-blue-600/30 transition disabled:opacity-50"
        >
          <Play className="w-4 h-4 fill-current" />
          Iniciar Migração no Engine
        </button>
      </div>

    </div>
  );
}
