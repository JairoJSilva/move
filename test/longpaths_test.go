package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"migrations-engine/pkg/audit"
	"migrations-engine/pkg/engine"
	"migrations-engine/pkg/scanner"
)

// Constantes de caminhos longos (>260 e >350 caracteres) contendo UTF-8 e acentuação
var (
	// Exemplo exato fornecido no requisito da missão
	promptExactExamplePath = filepath.Join(
		"Área_Transferência_Médica",
		"Laudos_2026",
		"Exames_Tomografia_Computadorizada_Com_Contraste_Protocolo_Expandido",
		"paciente_12345.dat",
	)

	// Exemplo do prompt aninhado em estrutura profunda > 260 caracteres (270 caracteres / runes)
	promptDeepPath270Runes = filepath.Join(
		"Secretaria_De_Estado_Da_Saúde_Governo_Do_Estado_De_São_Paulo_Brasil",
		"Hospital_Geral_De_Alta_Complexidade_Professor_Doutor_Euryclides_Jesus_Zerbini",
		"Área_Transferência_Médica",
		"Laudos_2026",
		"Exames_Tomografia_Computadorizada_Com_Contraste_Protocolo_Expandido",
		"paciente_12345.dat",
	)

	// Exemplo do prompt aninhado em estrutura profunda > 350 caracteres (359 caracteres / runes)
	promptDeepPath359Runes = filepath.Join(
		"Secretaria_De_Estado_Da_Saúde_Governo_Do_Estado_De_São_Paulo_Brasil",
		"Hospital_Geral_De_Alta_Complexidade_Professor_Doutor_Euryclides_Jesus_Zerbini",
		"Departamento_Especializado_De_Neurologia_E_Neurocirurgia_De_Alta_Precisão_E_Tecnologia_SP",
		"Área_Transferência_Médica",
		"Laudos_2026",
		"Exames_Tomografia_Computadorizada_Com_Contraste_Protocolo_Expandido",
		"paciente_12345.dat",
	)

	// Path 1: > 260 caracteres (277 caracteres / runes)
	longPath277Runes = filepath.Join(
		"Área_Transferência_Médica_Hospitalar_Nível_Central_2026",
		"Laudos_Periciais_Especializados_Com_Assinatura_ICP_Brasil_Ano_2026",
		"Exames_Tomografia_Computadorizada_Com_Contraste_Protocolo_Expandido_Neurologia",
		"paciente_protocolo_12345_laudo_completo_assinado_digitalmente.dat",
	)

	// Path 2: > 260 caracteres (336 caracteres / runes)
	longPath336Runes = filepath.Join(
		"Área_Transferência_Médica_Hospitalar_Nível_Central_2026",
		"Laudos_Periciais_Especializados_Com_Assinatura_ICP_Brasil_Ano_2026",
		"Exames_Tomografia_Computadorizada_Com_Contraste_Protocolo_Expandido_Neurologia",
		"Subdepartamento_Diagnóstico_Por_Imagem_Ressonância_Magnética_3T",
		"exame_ressonancia_cranio_contraste_seq_flair_paciente_54321.dcm",
	)

	// Path 3: > 350 caracteres (420 caracteres / runes)
	longPath420Runes = filepath.Join(
		"Área_Transferência_Médica_Hospitalar_Nível_Central_2026",
		"Laudos_Periciais_Especializados_Com_Assinatura_ICP_Brasil_Ano_2026",
		"Exames_Tomografia_Computadorizada_Com_Contraste_Protocolo_Expandido_Neurologia",
		"Subdepartamento_Diagnóstico_Por_Imagem_Ressonância_Magnética_3T",
		"Relatórios_Auditoria_Qualidade_Acreditação_ONA_Hospital_Geral_Estadual_São_Paulo",
		"paciente_protocolo_99887_laudo_expandido_alta_complexidade_versão_final_aprovada.dat",
	)

	// Path 4: > 350 caracteres (362 caracteres / runes com espaços e acentos variados)
	longPath362Runes = filepath.Join(
		"Prontuários_Eletrônicos_Clínica_Médica_São_Sebastião_2026",
		"Histórico_Ambulatorial_Cardiologia_Hemodinâmica_Eletrofisiologia",
		"Avaliações_Pré_Operatórias_Cirurgia_Cardiovascular_Alta_Complexidade",
		"Resultados_Cateterismo_Angioplastia_Coronária_Stent_Farmacológico_Bifurcação",
		"registro_médico_paciente_id_88776655_versão_assinada_dr_joão_silva_crm_123456_sp.pdf",
	)
)

// ---------------------------------------------------------------------------
// 1. Teste de Descoberta do Scanner com Caminhos Longos (>260 e >350 runes UTF-8)
// ---------------------------------------------------------------------------
func TestLongPaths_ScannerDiscovery(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "scanner_src")
	destDir := filepath.Join(tempDir, "scanner_dst")

	// Validação preliminar do comprimento dos caminhos de teste
	runes277 := utf8.RuneCountInString(longPath277Runes)
	runes336 := utf8.RuneCountInString(longPath336Runes)
	runes420 := utf8.RuneCountInString(longPath420Runes)
	runes362 := utf8.RuneCountInString(longPath362Runes)

	if runes277 <= 260 {
		t.Fatalf("longPath277Runes deve ter >260 runes, obteve %d", runes277)
	}
	if runes336 <= 260 {
		t.Fatalf("longPath336Runes deve ter >260 runes, obteve %d", runes336)
	}
	if runes420 <= 350 {
		t.Fatalf("longPath420Runes deve ter >350 runes, obteve %d", runes420)
	}
	if runes362 <= 350 {
		t.Fatalf("longPath362Runes deve ter >350 runes, obteve %d", runes362)
	}
	runesP270 := utf8.RuneCountInString(promptDeepPath270Runes)
	runesP359 := utf8.RuneCountInString(promptDeepPath359Runes)

	if runesP270 <= 260 {
		t.Fatalf("promptDeepPath270Runes deve ter >260 runes, obteve %d", runesP270)
	}
	if runesP359 <= 350 {
		t.Fatalf("promptDeepPath359Runes deve ter >350 runes, obteve %d", runesP359)
	}

	testFiles := []struct {
		relPath string
		size    int64
	}{
		{promptExactExamplePath, 512},
		{promptDeepPath270Runes, 1024},
		{promptDeepPath359Runes, 2048},
		{longPath277Runes, 1024},
		{longPath336Runes, 2048},
		{longPath420Runes, 4096},
		{longPath362Runes, 8192},
	}

	var expectedTotalBytes int64
	for _, tf := range testFiles {
		createTestFile(t, filepath.Join(srcDir, tf.relPath), tf.size, time.Now())
		expectedTotalBytes += tf.size
	}

	// 1.1 Executa scanner.Preview
	opts := scanner.FilterOptions{
		Mode: scanner.ModeFull,
	}
	summary, err := scanner.Preview(t.Context(), srcDir, destDir, opts)
	if err != nil {
		t.Fatalf("scanner.Preview falhou: %v", err)
	}

	if summary.TotalFilesDiscovered != int64(len(testFiles)) {
		t.Errorf("scanner.Preview: esperado %d arquivos descobertos, obteve %d", len(testFiles), summary.TotalFilesDiscovered)
	}
	if summary.EligibleFilesCount != int64(len(testFiles)) {
		t.Errorf("scanner.Preview: esperado %d arquivos elegíveis, obteve %d", len(testFiles), summary.EligibleFilesCount)
	}
	if summary.EligibleBytesTotal != expectedTotalBytes {
		t.Errorf("scanner.Preview: esperado %d bytes, obteve %d", expectedTotalBytes, summary.EligibleBytesTotal)
	}

	// 1.2 Executa scanner.Scan em streaming
	itemsChan := make(chan scanner.ScannedItem, 50)
	errChan := make(chan error, 1)

	go func() {
		errChan <- scanner.Scan(t.Context(), srcDir, destDir, opts, itemsChan)
	}()

	var scannedFiles []scanner.ScannedItem
	var scannedDirs []scanner.ScannedItem

	for item := range itemsChan {
		if item.IsDir {
			scannedDirs = append(scannedDirs, item)
		} else {
			scannedFiles = append(scannedFiles, item)
		}
	}

	if err := <-errChan; err != nil {
		t.Fatalf("scanner.Scan falhou: %v", err)
	}

	if len(scannedFiles) != len(testFiles) {
		t.Fatalf("scanner.Scan: esperado %d arquivos, obteve %d", len(testFiles), len(scannedFiles))
	}

	// Valida que cada caminho relativo recuperado preserva os caracteres UTF-8 e o comprimento
	for _, sf := range scannedFiles {
		runes := utf8.RuneCountInString(sf.RelPath)
		if sf.RelPath != promptExactExamplePath && runes <= 260 {
			t.Errorf("caminho escaneado tem apenas %d runes: %s", runes, sf.RelPath)
		}
		if !strings.Contains(sf.RelPath, "Área") && !strings.Contains(sf.RelPath, "Prontuários") && !strings.Contains(sf.RelPath, "Secretaria") {
			t.Errorf("caracteres acentuados corrompidos no caminho: %s", sf.RelPath)
		}
		if sf.Size <= 0 {
			t.Errorf("tamanho inválido para %s: %d", sf.RelPath, sf.Size)
		}
		if !sf.NeedsCopy {
			t.Errorf("NeedsCopy deveria ser true no modo Full para %s", sf.RelPath)
		}
	}
}

// ---------------------------------------------------------------------------
// 2. Teste de Migração Completa com Caminhos Longos (>260 e >350 caracteres)
// ---------------------------------------------------------------------------
func TestLongPaths_Engine_FullSyncTransfer(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "engine_long_src")
	destDir := filepath.Join(tempDir, "engine_long_dst")
	auditDir := filepath.Join(tempDir, "engine_long_audit")

	t0 := time.Now().Add(-5 * time.Hour).Truncate(time.Second)

	filesSpec := []struct {
		relPath string
		size    int64
		runes   int
	}{
		{promptExactExamplePath, 64 * 1024, utf8.RuneCountInString(promptExactExamplePath)},
		{promptDeepPath270Runes, 128 * 1024, utf8.RuneCountInString(promptDeepPath270Runes)},
		{promptDeepPath359Runes, 256 * 1024, utf8.RuneCountInString(promptDeepPath359Runes)},
		{longPath277Runes, 128 * 1024, utf8.RuneCountInString(longPath277Runes)},
		{longPath336Runes, 256 * 1024, utf8.RuneCountInString(longPath336Runes)},
		{longPath420Runes, 512 * 1024, utf8.RuneCountInString(longPath420Runes)},
		{longPath362Runes, 1024 * 1024, utf8.RuneCountInString(longPath362Runes)},
	}

	var expectedTotalBytes int64
	for _, spec := range filesSpec {
		createTestFile(t, filepath.Join(srcDir, spec.relPath), spec.size, t0)
		expectedTotalBytes += spec.size
	}

	eng := engine.NewEngine()

	var mu sync.Mutex
	var auditedEvents []audit.AuditEvent
	eng.SetEventCallback(func(evt audit.AuditEvent) {
		mu.Lock()
		auditedEvents = append(auditedEvents, evt)
		mu.Unlock()
	})

	cfg := engine.JobConfig{
		JobID:          "longpath-fullsync-01",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull,
		},
		Concurrency: 4,
		AuditDir:    auditDir,
	}

	if err := eng.StartJob(cfg); err != nil {
		t.Fatalf("falha ao iniciar job de caminhos longos: %v", err)
	}

	metrics := waitForJobCompletion(t, eng, 10*time.Second)

	if metrics.Status != engine.StateCompleted {
		t.Fatalf("job com caminhos longos falhou: status=%s", metrics.Status)
	}
	if metrics.FilesCopied != int64(len(filesSpec)) {
		t.Fatalf("esperado %d arquivos copiados, obteve %d", len(filesSpec), metrics.FilesCopied)
	}
	if metrics.FilesFailed != 0 {
		t.Fatalf("esperado 0 falhas, obteve %d", metrics.FilesFailed)
	}
	if metrics.BytesTransferred != expectedTotalBytes {
		t.Fatalf("esperado %d bytes transferidos, obteve %d", expectedTotalBytes, metrics.BytesTransferred)
	}

	// Validação minuciosa de integridade, metadados e xxHash64 no destino
	for _, spec := range filesSpec {
		srcPath := filepath.Join(srcDir, spec.relPath)
		dstPath := filepath.Join(destDir, spec.relPath)

		// 1. Existência
		fiDst, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("arquivo de destino não encontrado para caminho longo: %s (erro: %v)", spec.relPath, err)
		}

		// 2. Tamanho
		if fiDst.Size() != spec.size {
			t.Errorf("tamanho divergente em %s: esperado %d, obteve %d", spec.relPath, spec.size, fiDst.Size())
		}

		// 3. mtime preservado
		if fiDst.ModTime().Unix() != t0.Unix() {
			t.Errorf("mtime não preservado para %s: esperado %v, obteve %v", spec.relPath, t0, fiDst.ModTime())
		}

		// 4. Integridade de Hash xxHash64
		srcHash, err1 := engine.ComputeFileHashXX64(srcPath)
		dstHash, err2 := engine.ComputeFileHashXX64(dstPath)
		if err1 != nil || err2 != nil {
			t.Fatalf("erro ao calcular hashes para %s: errSrc=%v errDst=%v", spec.relPath, err1, err2)
		}
		if srcHash != dstHash {
			t.Errorf("corrupção de hash em caminho longo %s: src=%s dst=%s", spec.relPath, srcHash, dstHash)
		}
	}

	// 5. Validação dos eventos de auditoria registrados
	mu.Lock()
	defer mu.Unlock()
	if len(auditedEvents) != len(filesSpec) {
		t.Errorf("esperado %d eventos de auditoria, obteve %d", len(filesSpec), len(auditedEvents))
	}
	for _, evt := range auditedEvents {
		if evt.Status != audit.StatusSuccess {
			t.Errorf("evento de auditoria com falha: %+v", evt)
		}
		if evt.SourceHashXX64 == "" || evt.DestHashXX64 == "" {
			t.Errorf("evento sem hash: %+v", evt)
		}
		if evt.SourceHashXX64 != evt.DestHashXX64 {
			t.Errorf("divergência de hash no evento: src=%s dst=%s", evt.SourceHashXX64, evt.DestHashXX64)
		}
	}
}

// ---------------------------------------------------------------------------
// 3. Teste de Delta Sync com Estruturas Profundas (>350 caracteres)
// ---------------------------------------------------------------------------
func TestLongPaths_Engine_DeltaSync(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "delta_long_src")
	destDir := filepath.Join(tempDir, "delta_long_dst")
	auditDir := filepath.Join(tempDir, "delta_long_audit")

	t0 := time.Now().Add(-10 * time.Hour).Truncate(time.Second)

	// Arquivos iniciais com caminhos longos
	createTestFile(t, filepath.Join(srcDir, longPath277Runes), 1024, t0)
	createTestFile(t, filepath.Join(srcDir, longPath420Runes), 2048, t0)

	// Passo 1: Executa Baseline Full Sync
	engFull := engine.NewEngine()
	cfgFull := engine.JobConfig{
		JobID:          "longpath-delta-base",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeFull,
		},
		Concurrency: 2,
		AuditDir:    auditDir,
	}
	if err := engFull.StartJob(cfgFull); err != nil {
		t.Fatalf("falha ao iniciar baseline: %v", err)
	}
	mFull := waitForJobCompletion(t, engFull, 5*time.Second)
	if mFull.FilesCopied != 2 {
		t.Fatalf("esperado 2 arquivos copiados na baseline, obteve %d", mFull.FilesCopied)
	}

	// Passo 2: Altera um arquivo >350 runes e adiciona outro >350 runes
	tModified := time.Now().Truncate(time.Second)

	// Modifica o arquivo de 420 runes (tamanho e conteúdo novos)
	createTestFile(t, filepath.Join(srcDir, longPath420Runes), 8192, tModified)

	// Adiciona novo arquivo de 362 runes
	createTestFile(t, filepath.Join(srcDir, longPath362Runes), 4096, tModified)

	// Passo 3: Executa Delta Sync
	engDelta := engine.NewEngine()
	cfgDelta := engine.JobConfig{
		JobID:          "longpath-delta-job",
		SourceDir:      srcDir,
		DestinationDir: destDir,
		Filters: scanner.FilterOptions{
			Mode: scanner.ModeDelta,
		},
		Concurrency: 2,
		AuditDir:    auditDir,
	}

	var skippedCount, copiedCount int
	var mu sync.Mutex
	engDelta.SetEventCallback(func(evt audit.AuditEvent) {
		mu.Lock()
		defer mu.Unlock()
		if evt.Action == audit.ActionSkipped {
			skippedCount++
		} else if evt.Action == audit.ActionCopied {
			copiedCount++
		}
	})

	if err := engDelta.StartJob(cfgDelta); err != nil {
		t.Fatalf("falha ao iniciar delta sync: %v", err)
	}
	mDelta := waitForJobCompletion(t, engDelta, 5*time.Second)

	if mDelta.Status != engine.StateCompleted {
		t.Fatalf("delta sync falhou com status %s", mDelta.Status)
	}

	// Esperado:
	// 1 arquivo pulado (longPath277Runes - inalterado)
	// 2 arquivos copiados (longPath420Runes - alterado, longPath362Runes - novo)
	if mDelta.FilesSkipped != 1 {
		t.Errorf("esperado 1 arquivo pulado pelo delta sync, obteve %d", mDelta.FilesSkipped)
	}
	if mDelta.FilesCopied != 2 {
		t.Errorf("esperado 2 arquivos copiados pelo delta sync, obteve %d", mDelta.FilesCopied)
	}

	// Validação final de integridade de hash dos 3 arquivos no destino
	allFiles := []string{longPath277Runes, longPath420Runes, longPath362Runes}
	for _, relPath := range allFiles {
		srcP := filepath.Join(srcDir, relPath)
		dstP := filepath.Join(destDir, relPath)

		sHash, err1 := engine.ComputeFileHashXX64(srcP)
		dHash, err2 := engine.ComputeFileHashXX64(dstP)
		if err1 != nil || err2 != nil {
			t.Fatalf("falha ao computar hash pós-delta (%s): %v / %v", relPath, err1, err2)
		}
		if sHash != dHash {
			t.Errorf("hash não confere pós-delta para %s: src=%s dst=%s", relPath, sHash, dHash)
		}
	}
}

// ---------------------------------------------------------------------------
// 4. Teste de Normalização de Caminhos Estendidos (NormalizePath)
// ---------------------------------------------------------------------------
func TestLongPaths_PathNormalization(t *testing.T) {
	// Caminho vazio
	if got := engine.NormalizePath(""); got != "" {
		t.Errorf("NormalizePath(\"\") esperado \"\", obteve %q", got)
	}

	// Caminho normal Unix/Linux
	unixPath := "/var/data/medical/" + longPath420Runes
	gotUnix := engine.NormalizePath(unixPath)
	if !strings.Contains(gotUnix, "Área_Transferência_Médica") {
		t.Errorf("NormalizePath corrompeu caminho Unix: %s", gotUnix)
	}

	// Caminho com pontos relativos que devem ser limpos
	dirtyPath := "/var/data/medical/./subfolder/../" + longPath277Runes
	cleaned := engine.NormalizePath(dirtyPath)
	if strings.Contains(cleaned, "..") || strings.Contains(cleaned, "./") {
		t.Errorf("NormalizePath falhou ao limpar pontos relativos: %s", cleaned)
	}
}

// ---------------------------------------------------------------------------
// 5. Teste de Auditoria JSONL e CSV com Caminhos UTF-8 >350 Caracteres
// ---------------------------------------------------------------------------
func TestLongPaths_AuditTrail_UTF8Preservation(t *testing.T) {
	tempDir := t.TempDir()
	jobID := "audit-longpath-01"

	logger, err := audit.NewLogger(jobID, tempDir)
	if err != nil {
		t.Fatalf("falha ao criar logger: %v", err)
	}

	evt := audit.AuditEvent{
		EventID:        "evt-long-001",
		JobID:          jobID,
		Timestamp:      time.Now().Truncate(time.Millisecond),
		Action:         audit.ActionCopied,
		SourcePath:     "/mnt/source/" + longPath420Runes,
		DestPath:       "/mnt/dest/" + longPath420Runes,
		SizeBytes:      512000,
		DurationMs:     120,
		ThroughputMBs:  4.27,
		SourceHashXX64: "a1b2c3d4e5f60718",
		DestHashXX64:   "a1b2c3d4e5f60718",
		SourceMTime:    time.Now().Truncate(time.Millisecond),
		DestMTime:      time.Now().Truncate(time.Millisecond),
		Status:         audit.StatusSuccess,
	}

	if err := logger.LogEvent(evt); err != nil {
		t.Fatalf("falha ao logar evento com caminho longo: %v", err)
	}

	if err := logger.Close(); err != nil {
		t.Fatalf("falha ao fechar logger: %v", err)
	}

	// 5.1 Verifica se o JSONL foi gravado preservando UTF-8
	jsonlPath := filepath.Join(tempDir, jobID+"_events.jsonl")
	data, err := os.ReadFile(jsonlPath)
	if err != nil {
		t.Fatalf("falha ao ler JSONL: %v", err)
	}

	var parsed audit.AuditEvent
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("falha ao desserializar JSONL: %v", err)
	}

	if parsed.SourcePath != evt.SourcePath {
		t.Errorf("SourcePath divergente no JSONL:\nesperado: %s\nobteve:    %s", evt.SourcePath, parsed.SourcePath)
	}
	if !strings.Contains(parsed.SourcePath, "Área_Transferência_Médica") {
		t.Errorf("caracteres UTF-8 foram alterados no JSONL: %s", parsed.SourcePath)
	}
	if utf8.RuneCountInString(parsed.SourcePath) <= 350 {
		t.Errorf("esperado SourcePath com >350 runes no JSONL, obteve %d", utf8.RuneCountInString(parsed.SourcePath))
	}

	// 5.2 Verifica se o CSV preserva o caminho completo
	csvPath := filepath.Join(tempDir, jobID+"_report.csv")
	csvData, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("falha ao ler CSV: %v", err)
	}
	csvContent := string(csvData)
	if !strings.Contains(csvContent, "paciente_protocolo_99887_laudo_expandido") {
		t.Errorf("caminho longo ausente no CSV gravado")
	}
	if !strings.Contains(csvContent, "Área_Transferência_Médica") {
		t.Errorf("caracteres acentuados corrompidos no CSV")
	}
}
