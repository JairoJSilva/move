package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"migrations-engine/pkg/api"
	"migrations-engine/pkg/engine"
	"migrations-engine/pkg/platform"
	"migrations-engine/pkg/ui"
)

const (
	AppVersion = "1.0.0"
	AppName    = "MoveOps"
)

func main() {
	port := flag.Int("port", 8080, "Porta do servidor HTTP e WebSocket")
	webDir := flag.String("dir", "", "Diretório de arquivos estáticos da UI (ex: ./web/dist). Se não informado, utiliza a UI embarcada (embed.FS)")
	auditDir := flag.String("audit-dir", "./audit_logs", "Diretório para armazenamento de logs de auditoria e relatórios")
	showVersion := flag.Bool("version", false, "Exibir versão do MoveOps")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		os.Exit(0)
	}

	log.Printf("=====================================================")
	log.Printf("  %s v%s - Engine Inicializado", AppName, AppVersion)
	log.Printf("=====================================================")

	// Configura prioridade de I/O em background
	if err := platform.SetBackgroundPriority(); err != nil {
		log.Printf("[MoveOps] Aviso: prioridade de I/O em background não pôde ser aplicada: %v", err)
	} else {
		log.Printf("[MoveOps] Prioridade de I/O em background ativada (IDLE / Background Mode)")
	}

	// Se não informado dir estático, verifica fallback para UI embarcada ou pasta física
	staticPath := *webDir
	if staticPath == "" {
		if ui.HasEmbeddedUI() {
			log.Printf("[MoveOps] Frontend SPA embarcado ativado (pkg/ui/embed)")
		} else if fi, err := os.Stat("./web/dist"); err == nil && fi.IsDir() {
			staticPath = "./web/dist"
			log.Printf("[MoveOps] Frontend estático localizado em: %s", staticPath)
		}
	} else {
		log.Printf("[MoveOps] Servindo frontend a partir do diretório: %s", staticPath)
	}

	_ = os.MkdirAll(*auditDir, 0755)

	migrationEngine := engine.NewEngine()

	srv := api.NewServer(api.ConfigServer{
		Port:      *port,
		StaticDir: staticPath,
		Engine:    migrationEngine,
	})

	// Captura sinais de interrupção para encerramento gracioso
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[MoveOps] Erro fatal no servidor HTTP: %v", err)
		}
	}()

	// Aguarda sinal de encerramento
	sig := <-shutdownChan
	log.Printf("\n[MoveOps] Sinal recebido: %v. Iniciando encerramento gracioso...", sig)

	// Interrompe qualquer job em execução de forma segura
	_ = migrationEngine.StopJob()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[MoveOps] Erro durante o desligamento do servidor: %v", err)
	}

	log.Printf("[MoveOps] Sistema finalizado com segurança.")
}
