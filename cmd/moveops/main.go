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
	"migrations-engine/pkg/tray"
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
	trayMode := flag.Bool("tray", false, "Executar no modo Ícone na Bandeja do Relógio (System Tray)")
	openBrowser := flag.Bool("open", false, "Abrir automaticamente o navegador padrão ao iniciar o MoveOps")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		os.Exit(0)
	}

	log.Printf("=====================================================")
	log.Printf("  %s v%s - High-Performance Migration Engine", AppName, AppVersion)
	log.Printf("=====================================================")

	// Configura prioridade de I/O em background no início do processo
	if err := platform.SetBackgroundPriority(); err != nil {
		log.Printf("[MoveOps] Aviso: prioridade de I/O em background não pôde ser aplicada: %v", err)
	} else {
		log.Printf("[MoveOps] Prioridade de I/O em background ativada (IDLE / Background Mode)")
	}

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

	appURL := fmt.Sprintf("http://localhost:%d", *port)

	// Abertura automática de navegador quando solicitado ou quando no modo bandeja
	if *openBrowser || *trayMode {
		go func() {
			time.Sleep(300 * time.Millisecond)
			log.Printf("[MoveOps] Abrindo navegador padrão em: %s", appURL)
			if err := tray.OpenBrowser(appURL); err != nil {
				log.Printf("[MoveOps] Aviso ao abrir navegador: %v", err)
			}
		}()
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[MoveOps] Erro fatal no servidor HTTP: %v", err)
		}
	}()

	if *trayMode {
		trayCtrl := tray.NewTray(tray.Config{
			Title: AppName,
			Port:  *port,
			URL:   appURL,
			OnOpen: func() {
				_ = tray.OpenBrowser(appURL)
			},
			OnStatus: func() string {
				m := migrationEngine.GetMetrics()
				if m.Status == engine.StateRunning {
					return fmt.Sprintf("Migração em andamento: %.1f%% (Vazão: %.1f MB/s)", m.ProgressPercent, m.CurrentThroughputMBs)
				}
				if m.Status == engine.StatePaused {
					return fmt.Sprintf("Migração pausada (Job: %s)", m.JobID)
				}
				return fmt.Sprintf("Status: %s (Porta %d)", m.Status, *port)
			},
			OnToggle: func() (bool, string) {
				m := migrationEngine.GetMetrics()
				if m.Status == engine.StateRunning {
					if err := migrationEngine.PauseJob(); err == nil {
						return true, "Migração pausada com sucesso"
					}
					return false, "Falha ao pausar migração"
				} else if m.Status == engine.StatePaused {
					if err := migrationEngine.ResumeJob(); err == nil {
						return false, "Migração retomada com sucesso"
					}
					return true, "Falha ao retomar migração"
				}
				return false, "Nenhum trabalho em andamento para pausar/retomar"
			},
			OnQuit: func() {
				select {
				case shutdownChan <- syscall.SIGTERM:
				default:
				}
			},
		})

		go func() {
			sig := <-shutdownChan
			log.Printf("\n[MoveOps] Sinal recebido: %v. Encerrando bandeja...", sig)
			trayCtrl.Stop()
		}()

		if err := trayCtrl.Run(); err != nil {
			log.Printf("[MoveOps] Aviso no loop da bandeja: %v", err)
		}
	} else {
		sig := <-shutdownChan
		log.Printf("\n[MoveOps] Sinal recebido: %v. Encerrando graciosamente...", sig)
	}

	_ = migrationEngine.StopJob()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[MoveOps] Erro ao desligar servidor: %v", err)
	}

	log.Printf("[MoveOps] Encerrado com sucesso.")
}
