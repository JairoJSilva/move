import React, { useState, useEffect, useRef } from 'react';
import { Navbar } from './components/Navbar';
import { OverviewTab } from './components/OverviewTab';
import { StorageTab } from './components/StorageTab';
import { MigrationConfig } from './components/MigrationConfig';
import { LiveDashboard } from './components/LiveDashboard';
import { LiveFeedConsole } from './components/LiveFeedConsole';
import { AuditReports } from './components/AuditReports';
import { DocsTab } from './components/DocsTab';
import { FileBrowserModal } from './components/FileBrowserModal';
import { PreviewModal } from './components/PreviewModal';
import { api } from './services/api';

export default function App() {
  // Estado de Navegação e Tema (6 Abas Segmentadas)
  const [activeTab, setActiveTab] = useState('overview'); // 'overview', 'storage', 'config', 'telemetry', 'audit', 'docs'
  const [darkMode, setDarkMode] = useState(true);

  // Estado de Comunicação e Telemetria
  const [wsStatus, setWsStatus] = useState('DISCONNECTED'); // CONNECTED, CONNECTING, DISCONNECTED
  const [metrics, setMetrics] = useState({
    job_id: '',
    status: 'IDLE',
    current_phase: 'PHASE_1_BASELINE',
    current_throughput_mbs: 0,
    current_iops: 0,
    limit_bandwidth_mb: 0,
    limit_iops: 0,
    total_files_discovered: 0,
    files_copied: 0,
    files_skipped: 0,
    files_failed: 0,
    total_bytes_discovered: 0,
    bytes_transferred: 0,
    progress_percent: 0,
    active_workers: 0,
    current_file: '',
    elapsed_seconds: 0,
    eta_seconds: 0,
  });
  const [throughputHistory, setThroughputHistory] = useState([]);
  const [eventsFeed, setEventsFeed] = useState([]);

  // Toast / Notificações Rápidas
  const [toast, setToast] = useState(null);
  const showToast = (message, type = 'info') => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 4000);
  };

  // Parâmetros de Configuração da Migração
  const [sourceDir, setSourceDir] = useState('');
  const [destDir, setDestDir] = useState('');
  const [bandwidthLimit, setBandwidthLimit] = useState(50);
  const [iopsLimit, setIopsLimit] = useState(300);
  const [concurrency, setConcurrency] = useState(4);
  const [enableAudit, setEnableAudit] = useState(true);
  const [auditDir, setAuditDir] = useState('./audit_logs');

  // Filtros Temporais e Padrões
  const [filterMode, setFilterMode] = useState('FULL'); // FULL, DELTA_ONLY, DATE_RANGE, BY_YEAR, BY_MONTH, BY_DAY, RETENTION
  const [yearFilter, setYearFilter] = useState(new Date().getFullYear());
  const [monthFilter, setMonthFilter] = useState(new Date().getMonth() + 1);
  const [dayFilter, setDayFilter] = useState(new Date().toISOString().split('T')[0]);
  const [startDateFilter, setStartDateFilter] = useState('');
  const [endDateFilter, setEndDateFilter] = useState('');
  const [olderThanDays, setOlderThanDays] = useState(30);
  const [includePatterns, setIncludePatterns] = useState('');
  const [excludePatterns, setExcludePatterns] = useState('');

  // Modais
  const [browserModalOpen, setBrowserModalOpen] = useState(false);
  const [browserTarget, setBrowserTarget] = useState('source');
  const [previewModalOpen, setPreviewModalOpen] = useState(false);
  const [previewData, setPreviewData] = useState(null);

  // Inicialização e Conexão WebSocket
  useEffect(() => {
    const wsClient = api.createWebSocketClient({
      onMetrics: (newMetrics) => {
        setMetrics(newMetrics);
        setThroughputHistory((prev) => {
          const next = [...prev, newMetrics.current_throughput_mbs || 0];
          return next.slice(-30);
        });
      },
      onFileEvent: (evt) => {
        setEventsFeed((prev) => {
          const next = [...prev, evt];
          return next.slice(-250); // Mantém últimos 250 eventos para renderização veloz
        });
      },
      onStatusChange: (status) => {
        setWsStatus(status);
      },
    });

    // Fallback de polling caso WebSocket demore a conectar ou reconectar
    const pollInterval = setInterval(async () => {
      try {
        const currentStatus = await api.getStatus();
        if (currentStatus) {
          setMetrics((prev) => {
            // Só atualiza se o job_id ou status diferirem para evitar rerenders excessivos
            if (prev.status !== currentStatus.status || prev.bytes_transferred !== currentStatus.bytes_transferred) {
              return currentStatus;
            }
            return prev;
          });
        }
      } catch (e) {
        // Ignora silenciosamente erros intermitentes de polling
      }
    }, 1500);

    return () => {
      wsClient.disconnect();
      clearInterval(pollInterval);
    };
  }, []);

  // Handlers do Ciclo de Vida
  const handleStartMigration = async (customConfig) => {
    try {
      const config = customConfig || {
        source_dir: sourceDir,
        destination_dir: destDir,
        filters: { mode: filterMode },
        max_bandwidth_mb: parseFloat(bandwidthLimit) || 0,
        max_iops: parseInt(iopsLimit, 10) || 0,
        concurrency: parseInt(concurrency, 10) || 4,
        audit_dir: enableAudit ? auditDir : '',
      };

      const res = await api.startMigration(config);
      showToast(`Migração iniciada com sucesso! Job ID: ${res.job_id}`, 'success');
      setActiveTab('telemetry'); // Leva o usuário direto para a telemetria
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  const handlePause = async () => {
    try {
      await api.pauseMigration();
      showToast('Migração pausada.', 'warning');
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  const handleResume = async () => {
    try {
      await api.resumeMigration();
      showToast('Migração retomada com sucesso!', 'success');
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  const handleStop = async () => {
    try {
      await api.stopMigration();
      showToast('Migração abortada com segurança.', 'error');
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  const openBrowserModal = (target, customPath) => {
    setBrowserTarget(target);
    if (customPath) {
      if (target === 'source') setSourceDir(customPath);
      else setDestDir(customPath);
    }
    setBrowserModalOpen(true);
  };

  const handleSelectPath = (path) => {
    if (browserTarget === 'source') {
      setSourceDir(path);
    } else {
      setDestDir(path);
    }
    showToast(`Pasta de ${browserTarget === 'source' ? 'Origem' : 'Destino'} atualizada!`, 'success');
  };

  const handleOpenPreview = (summary) => {
    setPreviewData(summary);
    setPreviewModalOpen(true);
  };

  return (
    <div className={`min-h-screen ${darkMode ? 'dark bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'}`}>
      
      {/* Toast Notification */}
      {toast && (
        <div className={`fixed bottom-5 right-5 z-50 px-4 py-3 rounded-xl shadow-2xl text-xs font-semibold flex items-center gap-2 border animate-in slide-in-from-bottom duration-200 ${
          toast.type === 'success'
            ? 'bg-emerald-950/90 text-emerald-300 border-emerald-500/50'
            : toast.type === 'error'
            ? 'bg-rose-950/90 text-rose-300 border-rose-500/50'
            : toast.type === 'warning'
            ? 'bg-amber-950/90 text-amber-300 border-amber-500/50'
            : 'bg-slate-900 text-slate-200 border-slate-700'
        }`}>
          <span>{toast.message}</span>
        </div>
      )}

      {/* Top Navigation */}
      <Navbar
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        wsStatus={wsStatus}
        metrics={metrics}
        darkMode={darkMode}
        setDarkMode={setDarkMode}
      />

      {/* Main Content Area */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        
        {/* ABA 1: VISÃO GERAL (OVERVIEW) */}
        {activeTab === 'overview' && (
          <OverviewTab
            metrics={metrics}
            setActiveTab={setActiveTab}
            sourceDir={sourceDir}
            destDir={destDir}
          />
        )}

        {/* ABA 2: ARMAZENAMENTO & VOLUMES */}
        {activeTab === 'storage' && (
          <StorageTab
            sourceDir={sourceDir}
            setSourceDir={setSourceDir}
            destDir={destDir}
            setDestDir={setDestDir}
            onOpenBrowser={openBrowserModal}
            setActiveTab={setActiveTab}
          />
        )}

        {/* ABA 3: NOVA MIGRAÇÃO (CONFIGURAÇÃO) */}
        {activeTab === 'config' && (
          <MigrationConfig
            sourceDir={sourceDir}
            setSourceDir={setSourceDir}
            destDir={destDir}
            setDestDir={setDestDir}
            onOpenBrowser={openBrowserModal}
            onOpenPreview={handleOpenPreview}
            onStartMigration={handleStartMigration}
            isJobRunning={metrics.status === 'RUNNING' || metrics.status === 'PAUSED'}
            bandwidthLimit={bandwidthLimit}
            setBandwidthLimit={setBandwidthLimit}
            iopsLimit={iopsLimit}
            setIopsLimit={setIopsLimit}
            concurrency={concurrency}
            setConcurrency={setConcurrency}
            enableAudit={enableAudit}
            setEnableAudit={setEnableAudit}
            auditDir={auditDir}
            setAuditDir={setAuditDir}
            filterMode={filterMode}
            setFilterMode={setFilterMode}
            yearFilter={yearFilter}
            setYearFilter={setYearFilter}
            monthFilter={monthFilter}
            setMonthFilter={setMonthFilter}
            dayFilter={dayFilter}
            setDayFilter={setDayFilter}
            startDateFilter={startDateFilter}
            setStartDateFilter={setStartDateFilter}
            endDateFilter={endDateFilter}
            setEndDateFilter={setEndDateFilter}
            olderThanDays={olderThanDays}
            setOlderThanDays={setOlderThanDays}
            includePatterns={includePatterns}
            setIncludePatterns={setIncludePatterns}
            excludePatterns={excludePatterns}
            setExcludePatterns={setExcludePatterns}
            onSwitchToStorage={() => setActiveTab('storage')}
          />
        )}

        {/* ABA 4: TELEMETRIA EM TEMPO REAL */}
        {activeTab === 'telemetry' && (
          <div className="space-y-8 animate-in fade-in duration-200">
            {/* Dashboard Master de Métricas e Controles */}
            <LiveDashboard
              metrics={metrics}
              throughputHistory={throughputHistory}
              onStart={() => handleStartMigration()}
              onPause={handlePause}
              onResume={handleResume}
              onStop={handleStop}
              onSwitchToConfig={() => setActiveTab('config')}
            />

            {/* Live Feed Terminal com Verificação de Integridade xxHash64 */}
            <LiveFeedConsole
              events={eventsFeed}
              onClearEvents={() => setEventsFeed([])}
            />
          </div>
        )}

        {/* ABA 5: AUDITORIA & RELATÓRIOS */}
        {activeTab === 'audit' && (
          <div className="animate-in fade-in duration-200">
            <AuditReports currentJobId={metrics?.job_id} />
          </div>
        )}

        {/* ABA 6: DOCUMENTAÇÃO TÉCNICA OFICIAL */}
        {activeTab === 'docs' && (
          <DocsTab />
        )}

      </main>

      {/* Modais Globais */}
      <FileBrowserModal
        isOpen={browserModalOpen}
        onClose={() => setBrowserModalOpen(false)}
        initialPath={browserTarget === 'source' ? sourceDir : destDir}
        targetType={browserTarget}
        onSelectPath={handleSelectPath}
      />

      <PreviewModal
        isOpen={previewModalOpen}
        onClose={() => setPreviewModalOpen(false)}
        previewData={previewData}
        maxBandwidthMB={parseFloat(bandwidthLimit) || 0}
        onStartMigration={() => handleStartMigration()}
      />

    </div>
  );
}
