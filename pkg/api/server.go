package api

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"migrations-engine/pkg/audit"
	"migrations-engine/pkg/engine"
	"migrations-engine/pkg/ui"
)

// Server encapsula o servidor HTTP e WebSockets do MoveOps.
type Server struct {
	httpServer *http.Server
	engine     *engine.Engine
	hub        *WSHub
	handler    *Handler
	staticDir  string
	staticFS   fs.FS
	port       int
}

// ConfigServer define parâmetros de inicialização do servidor.
type ConfigServer struct {
	Port      int
	StaticDir string
	StaticFS  fs.FS
	Engine    *engine.Engine
}

// NewServer inicializa o servidor de API e WebSockets.
func NewServer(cfg ConfigServer) *Server {
	if cfg.Port <= 0 {
		cfg.Port = 8080
	}
	if cfg.Engine == nil {
		cfg.Engine = engine.NewEngine()
	}

	hub := NewWSHub()
	h := NewHandler(cfg.Engine, hub)

	// Registra callback de progresso do motor para broadcast instantâneo no WebSocket
	cfg.Engine.SetProgressCallback(func(m engine.TelemetryMetrics) {
		hub.Broadcast(WSMessage{
			Type:    "METRICS_UPDATE",
			Payload: m,
		})
	})

	// Registra callback de eventos de arquivos para broadcast instantâneo no WebSocket
	cfg.Engine.SetEventCallback(func(evt audit.AuditEvent) {
		hub.Broadcast(WSMessage{
			Type:    "FILE_EVENT",
			Payload: evt,
		})
	})

	var staticFS fs.FS = cfg.StaticFS
	if cfg.StaticDir == "" && staticFS == nil {
		if ui.HasEmbeddedUI() {
			if sub, err := ui.GetFS(); err == nil {
				staticFS = sub
				log.Printf("[MoveOps] Web UI embarcada ativada prioritariamente (embed.FS)")
			}
		}
	}

	srv := &Server{
		engine:    cfg.Engine,
		hub:       hub,
		handler:   h,
		staticDir: cfg.StaticDir,
		staticFS:  staticFS,
		port:      cfg.Port,
	}

	mux := http.NewServeMux()
	srv.setupRoutes(mux)

	corsHandler := srv.corsMiddleware(mux)

	srv.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      corsHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return srv
}

func (s *Server) setupRoutes(mux *http.ServeMux) {
	// 1. Endpoints de Descoberta e Navegação
	mux.HandleFunc("/api/v1/disks", s.handler.HandleGetDisks)
	mux.HandleFunc("/api/v1/browse", s.handler.HandleBrowse)

	// 2. Estimativa e Preview
	mux.HandleFunc("/api/v1/migration/preview", s.handler.HandlePreview)
	mux.HandleFunc("/api/v1/jobs/preview", s.handler.HandlePreview)

	// 3. Controle do Ciclo de Vida da Migração
	mux.HandleFunc("/api/v1/migration/start", s.handler.HandleStartMigration)
	mux.HandleFunc("/api/v1/jobs/start", s.handler.HandleStartMigration)

	mux.HandleFunc("/api/v1/migration/pause", s.handler.HandlePauseMigration)
	mux.HandleFunc("/api/v1/jobs/pause", s.handler.HandlePauseMigration)

	mux.HandleFunc("/api/v1/migration/resume", s.handler.HandleResumeMigration)
	mux.HandleFunc("/api/v1/jobs/resume", s.handler.HandleResumeMigration)

	mux.HandleFunc("/api/v1/migration/stop", s.handler.HandleStopMigration)
	mux.HandleFunc("/api/v1/jobs/stop", s.handler.HandleStopMigration)

	mux.HandleFunc("/api/v1/migration/status", s.handler.HandleGetStatus)
	mux.HandleFunc("/api/v1/jobs/status", s.handler.HandleGetStatus)

	// 4. Rate Limiting Dinâmico a Quente
	mux.HandleFunc("/api/v1/migration/rate-limit", s.handler.HandleRateLimit)
	mux.HandleFunc("/api/v1/limits", s.handler.HandleRateLimit)

	// 5. Relatórios e Auditoria
	mux.HandleFunc("/api/v1/reports/summary", s.handler.HandleSummaryReport)
	mux.HandleFunc("/api/v1/reports/export", s.handler.HandleExportReport)

	// 6. Canal de WebSocket para Streaming de Telemetria
	mux.HandleFunc("/api/v1/ws", func(w http.ResponseWriter, r *http.Request) {
		s.hub.HandleWS(w, r, func() *WSMessage {
			return &WSMessage{
				Type:    "METRICS_UPDATE",
				Payload: s.engine.GetMetrics(),
			}
		})
	})

	// 7. Servir Frontend SPA (Web UI) com fallback para index.html
	if s.staticDir != "" {
		mux.HandleFunc("/", s.serveDiskSPA)
	} else if s.staticFS != nil {
		mux.HandleFunc("/", s.serveEmbeddedSPA)
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>MoveOps API</title></head>
<body style="font-family: sans-serif; padding: 2rem; background: #0f172a; color: #f8fafc;">
  <h1>MoveOps Engine</h1>
  <p>Backend API está ativo e respondendo na porta ` + fmt.Sprintf("%d", s.port) + `.</p>
  <p>Acesse os endpoints em <code>/api/v1/disks</code> ou conecte-se ao WebSocket em <code>/api/v1/ws</code>.</p>
</body>
</html>`))
		})
	}
}

func (s *Server) serveDiskSPA(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	cleanPath := filepath.Clean(r.URL.Path)
	targetPath := filepath.Join(s.staticDir, cleanPath)

	// Garante que o arquivo solicitado reside estritamente dentro do diretório estático
	rel, err := filepath.Rel(s.staticDir, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		http.NotFound(w, r)
		return
	}

	fi, err := os.Stat(targetPath)
	if err == nil && !fi.IsDir() {
		http.ServeFile(w, r, targetPath)
		return
	}

	// Fallback para index.html (SPA Router)
	indexPath := filepath.Join(s.staticDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		http.ServeFile(w, r, indexPath)
		return
	}

	http.NotFound(w, r)
}

func (s *Server) serveEmbeddedSPA(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	reqPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if reqPath == "" || reqPath == "." {
		reqPath = "index.html"
	}

	// 1. Tenta abrir o arquivo solicitado no FS embarcado
	f, err := s.staticFS.Open(reqPath)
	if err == nil {
		defer f.Close()
		stat, err := f.Stat()
		if err == nil && !stat.IsDir() {
			http.FileServer(http.FS(s.staticFS)).ServeHTTP(w, r)
			return
		}
	}

	// 2. Fallback para index.html (SPA Router React)
	indexData, err := fs.ReadFile(s.staticFS, "index.html")
	if err == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(indexData)
		return
	}

	http.NotFound(w, r)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Start inicia o servidor em segundo plano.
func (s *Server) Start() error {
	log.Printf("[MoveOps] Servidor HTTP iniciado em http://localhost:%d", s.port)
	return s.httpServer.ListenAndServe()
}

// Shutdown finaliza o servidor graciosamente.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Printf("[MoveOps] Encerrando servidor HTTP...")
	return s.httpServer.Shutdown(ctx)
}
