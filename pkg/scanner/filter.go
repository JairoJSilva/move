package scanner

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EvaluateItem avalia se um item atende às regras configuradas em FilterOptions.
// Retorna (needsCopy, skipReason, error).
func EvaluateItem(sourcePath, relPath string, info os.FileInfo, opts FilterOptions) (bool, string, error) {
	if info.IsDir() {
		// Diretórios são avaliados para criação no destino, mas não contam como arquivo de dados
		return true, "", nil
	}

	size := info.Size()
	modTime := info.ModTime()
	fileName := filepath.Base(relPath)

	// 1. Filtros de Exclusão por Padrão
	for _, pattern := range opts.ExcludePatterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if matchPattern(pattern, fileName, relPath) {
			return false, "excluded_by_pattern: " + pattern, nil
		}
	}

	// 2. Filtros de Inclusão por Padrão (se configurados)
	if len(opts.IncludePatterns) > 0 {
		matched := false
		for _, pattern := range opts.IncludePatterns {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				continue
			}
			if matchPattern(pattern, fileName, relPath) {
				matched = true
				break
			}
		}
		if !matched {
			return false, "not_in_include_patterns", nil
		}
	}

	// 3. Filtros de Tamanho em Bytes
	if opts.MinSizeBytes > 0 && size < opts.MinSizeBytes {
		return false, "smaller_than_min_size", nil
	}
	if opts.MaxSizeBytes > 0 && size > opts.MaxSizeBytes {
		return false, "greater_than_max_size", nil
	}

	// 4. Filtros Temporais por Modo
	switch opts.Mode {
	case ModeFull:
		// Copia tudo que passar pelos padrões e tamanhos

	case ModeDateRange:
		if opts.StartDate != nil && modTime.Before(*opts.StartDate) {
			return false, "mod_time_before_start_date", nil
		}
		if opts.EndDate != nil && modTime.After(*opts.EndDate) {
			return false, "mod_time_after_end_date", nil
		}

	case ModeYear:
		if opts.Year > 0 && modTime.Year() != opts.Year {
			return false, "year_mismatch", nil
		}

	case ModeMonth:
		if opts.Year > 0 && modTime.Year() != opts.Year {
			return false, "year_mismatch", nil
		}
		if opts.Month > 0 && int(modTime.Month()) != opts.Month {
			return false, "month_mismatch", nil
		}

	case ModeDay:
		if opts.TargetDay != nil {
			target := opts.TargetDay.Truncate(24 * time.Hour)
			current := modTime.Truncate(24 * time.Hour)
			if !target.Equal(current) {
				return false, "day_mismatch", nil
			}
		}

	case ModeRetention:
		if opts.OlderThanDays > 0 {
			cutoff := time.Now().AddDate(0, 0, -opts.OlderThanDays)
			if !modTime.Before(cutoff) {
				return false, "newer_than_retention_cutoff", nil
			}
		}

	case ModeDelta:
		// No modo delta puro, a decisão depende da comparação com o destino logo abaixo
	}

	// 5. Verificação Delta contra o Arquivo no Destino (se aplicável)
	// Se for ModeDelta OU se o modo exigir comparação de integridade com destino
	if opts.DestinationDir != "" {
		destPath := filepath.Join(opts.DestinationDir, relPath)
		destInfo, err := os.Stat(destPath)
		if err == nil && !destInfo.IsDir() {
			// Arquivo já existe no destino. Compara tamanho e timestamp
			destSize := destInfo.Size()
			destModTime := destInfo.ModTime()

			// Tolerância de 1.5 segundos para compensar limitações de resolução de FAT/NTFS/ext4
			diffSecs := math.Abs(modTime.Sub(destModTime).Seconds())

			if size == destSize && diffSecs <= 2.0 {
				if opts.Mode == ModeDelta {
					return false, "delta_identical_size_and_mtime", nil
				}
			}
		}
	}

	return true, "", nil
}

// matchPattern avalia se o arquivo corresponde a um padrão glob simples ou caminho
func matchPattern(pattern, fileName, relPath string) bool {
	// Checagem de glob direto no nome do arquivo (ex: *.tmp, doc_*.pdf)
	if matched, _ := filepath.Match(pattern, fileName); matched {
		return true
	}

	// Checagem no caminho relativo normalizado (ex: node_modules/*, .git/*)
	normRel := filepath.ToSlash(relPath)
	normPattern := filepath.ToSlash(pattern)

	if matched, _ := filepath.Match(normPattern, normRel); matched {
		return true
	}

	// Suporte a pastas ignoradas (ex: node_modules ou .git)
	patternClean := strings.Trim(normPattern, "/*")
	if strings.Contains("/"+normRel+"/", "/"+patternClean+"/") {
		return true
	}

	return false
}
