package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"migrations-engine/pkg/engine"
)

func TestAPIRoutes(t *testing.T) {
	eng := engine.NewEngine()
	srv := NewServer(ConfigServer{
		Port:   8080,
		Engine: eng,
	})

	// 1. Testa GET /api/v1/disks
	req := httptest.NewRequest(http.MethodGet, "/api/v1/disks", nil)
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/disks falhou com código %d: %s", w.Code, w.Body.String())
	}

	var disksResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &disksResp); err != nil {
		t.Fatalf("falha ao decodificar resposta JSON de discos: %v", err)
	}
	if _, ok := disksResp["disks"]; !ok {
		t.Fatalf("campo 'disks' ausente na resposta")
	}

	// 2. Testa GET /api/v1/jobs/status
	reqStatus := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/status", nil)
	wStatus := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wStatus, reqStatus)

	if wStatus.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/jobs/status falhou com código %d", wStatus.Code)
	}

	var statusResp engine.TelemetryMetrics
	if err := json.Unmarshal(wStatus.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("falha ao decodificar status: %v", err)
	}
	if statusResp.Status != engine.StateIdle {
		t.Fatalf("esperado status IDLE, obteve %s", statusResp.Status)
	}

	// 3. Testa POST /api/v1/limits
	limitsPayload := []byte(`{"max_bandwidth_mb": 120.5, "max_iops": 400}`)
	reqLimits := httptest.NewRequest(http.MethodPost, "/api/v1/limits", bytes.NewBuffer(limitsPayload))
	reqLimits.Header.Set("Content-Type", "application/json")
	wLimits := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wLimits, reqLimits)

	if wLimits.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/limits falhou com código %d: %s", wLimits.Code, wLimits.Body.String())
	}

	// 4. Testa POST /api/v1/browse
	browsePayload := []byte(`{"path": ".", "operation": "read"}`)
	reqBrowse := httptest.NewRequest(http.MethodPost, "/api/v1/browse", bytes.NewBuffer(browsePayload))
	reqBrowse.Header.Set("Content-Type", "application/json")
	wBrowse := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wBrowse, reqBrowse)

	if wBrowse.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/browse falhou com código %d", wBrowse.Code)
	}
}

func TestAPISecurityValidations(t *testing.T) {
	eng := engine.NewEngine()
	srv := NewServer(ConfigServer{
		Port:   8080,
		Engine: eng,
	})

	// 1. Testa rejeição de Path Traversal no StartMigration (job_id malicioso)
	badJobPayload := []byte(`{"job_id": "../../etc/passwd", "source_dir": "/tmp", "destination_dir": "/tmp"}`)
	reqBadJob := httptest.NewRequest(http.MethodPost, "/api/v1/migration/start", bytes.NewBuffer(badJobPayload))
	reqBadJob.Header.Set("Content-Type", "application/json")
	wBadJob := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wBadJob, reqBadJob)

	if wBadJob.Code != http.StatusConflict {
		t.Fatalf("esperado código 409 Conflict para job_id com path traversal, obteve %d", wBadJob.Code)
	}

	// 2. Testa rejeição de caracteres nulos no Browse
	nullBrowsePayload := []byte("{\"path\": \"/var/\x00malicious\"}")
	reqNullBrowse := httptest.NewRequest(http.MethodPost, "/api/v1/browse", bytes.NewBuffer(nullBrowsePayload))
	reqNullBrowse.Header.Set("Content-Type", "application/json")
	wNullBrowse := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wNullBrowse, reqNullBrowse)

	if wNullBrowse.Code != http.StatusBadRequest {
		t.Fatalf("esperado código 400 Bad Request para path com caractere nulo, obteve %d", wNullBrowse.Code)
	}

	// 3. Testa rejeição de limites negativos no RateLimit
	negLimitsPayload := []byte(`{"max_bandwidth_mb": -10, "max_iops": -5}`)
	reqNegLimits := httptest.NewRequest(http.MethodPost, "/api/v1/limits", bytes.NewBuffer(negLimitsPayload))
	reqNegLimits.Header.Set("Content-Type", "application/json")
	wNegLimits := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wNegLimits, reqNegLimits)

	if wNegLimits.Code != http.StatusBadRequest {
		t.Fatalf("esperado código 400 Bad Request para limites negativos, obteve %d", wNegLimits.Code)
	}

	// 4. Testa rejeição de formato desconhecido no ExportReport
	reqBadFormat := httptest.NewRequest(http.MethodGet, "/api/v1/reports/export?format=exe", nil)
	wBadFormat := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wBadFormat, reqBadFormat)

	if wBadFormat.Code != http.StatusBadRequest {
		t.Fatalf("esperado código 400 Bad Request para formato inválido, obteve %d", wBadFormat.Code)
	}
}

func TestEmbeddedSPAServing(t *testing.T) {
	eng := engine.NewEngine()
	// NewServer sem StaticDir nem StaticFS deve acionar automaticamente a UI embarcada
	srv := NewServer(ConfigServer{
		Port:   8080,
		Engine: eng,
	})

	// 1. Testa rota raiz "/" entregando a SPA
	reqRoot := httptest.NewRequest(http.MethodGet, "/", nil)
	wRoot := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wRoot, reqRoot)

	if wRoot.Code != http.StatusOK {
		t.Fatalf("esperado status 200 na raiz '/', obteve %d", wRoot.Code)
	}
	bodyRoot := wRoot.Body.String()
	if !strings.Contains(bodyRoot, "MoveOps") && !strings.Contains(bodyRoot, "<div id=\"root\">") && !strings.Contains(bodyRoot, "OmniData HyperSync") {
		t.Fatalf("conteúdo inesperado na raiz: %s", bodyRoot)
	}

	// 2. Testa rota desconhecida da SPA com fallback para index.html (ex: "/dashboard")
	reqSpa := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	wSpa := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wSpa, reqSpa)

	if wSpa.Code != http.StatusOK {
		t.Fatalf("esperado status 200 no fallback da SPA, obteve %d", wSpa.Code)
	}
	if !strings.Contains(wSpa.Body.String(), "<div id=\"root\">") && !strings.Contains(wSpa.Body.String(), "MoveOps") && !strings.Contains(wSpa.Body.String(), "OmniData HyperSync") {
		t.Fatalf("fallback SPA não entregou o index.html da UI")
	}

	// 3. Testa rota de API inexistente retornando 404 e NÃO o index.html
	reqBadAPI := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	wBadAPI := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wBadAPI, reqBadAPI)

	if wBadAPI.Code != http.StatusNotFound {
		t.Fatalf("esperado status 404 para rota /api inexistente, obteve %d", wBadAPI.Code)
	}
}

func TestSPAPathTraversalProtection(t *testing.T) {
	eng := engine.NewEngine()

	// 1. Testa segurança de Path Traversal no modo Embedded
	srvEmbedded := NewServer(ConfigServer{
		Port:   8080,
		Engine: eng,
	})

	traversalVectors := []string{
		"/../../../../etc/passwd",
		"/..%2f..%2f..%2fetc%2fshadow",
		"/..\\..\\windows\\system32\\cmd.exe",
		"/nonexistent/../../../etc/hosts",
	}

	for _, vector := range traversalVectors {
		req := httptest.NewRequest(http.MethodGet, vector, nil)
		w := httptest.NewRecorder()
		srvEmbedded.httpServer.Handler.ServeHTTP(w, req)

		// No embedded SPA, o arquivo nunca deve vazar dados do host (deve servir index.html SPA ou 404)
		body := w.Body.String()
		if strings.Contains(body, "root:x:") || strings.Contains(body, "[boot loader]") {
			t.Fatalf("vazamento de arquivo de sistema host detectado para vetor %s", vector)
		}
	}

	// 2. Testa segurança de Path Traversal no modo Disco com verificação estrita de confinamento (filepath.Rel)
	tempStatic := t.TempDir()
	srvDisk := NewServer(ConfigServer{
		Port:      8081,
		StaticDir: tempStatic,
		Engine:    eng,
	})

	for _, vector := range traversalVectors {
		req := httptest.NewRequest(http.MethodGet, vector, nil)
		w := httptest.NewRecorder()
		srvDisk.httpServer.Handler.ServeHTTP(w, req)

		// Se o ServeMux emitir redirecionamento canônico (301/307) para limpar caminhos '..', segue o redirecionamento
		if w.Code == http.StatusTemporaryRedirect || w.Code == http.StatusMovedPermanently {
			loc := w.Header().Get("Location")
			reqRedirected := httptest.NewRequest(http.MethodGet, loc, nil)
			w = httptest.NewRecorder()
			srvDisk.httpServer.Handler.ServeHTTP(w, reqRedirected)
		}

		// Em modo disco sem index.html criado, tentativas de escape devem resultar em 404
		if w.Code != http.StatusNotFound {
			t.Fatalf("tentativa de escape em modo disco %s não retornou 404 (status: %d)", vector, w.Code)
		}

		body := w.Body.String()
		if strings.Contains(body, "root:x:") || strings.Contains(body, "[boot loader]") {
			t.Fatalf("vazamento de arquivo de sistema host detectado em modo disco para vetor %s", vector)
		}
	}
}
