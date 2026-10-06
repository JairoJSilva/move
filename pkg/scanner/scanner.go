package scanner

import (
	"context"
	"io/fs"
	"path/filepath"
)

// Scan inicia a varredura contínua e emite os itens elegíveis no itemChan.
// itemChan deve ser lido pelo consumidor com backpressure para manter uso de memória O(1).
func Scan(ctx context.Context, srcDir, destDir string, opts FilterOptions, itemChan chan<- ScannedItem) error {
	defer close(itemChan)

	opts.DestinationDir = destDir

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			// Não aborta a varredura completa se um arquivo inacessível for encontrado
			return nil
		}

		if path == srcDir {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		needsCopy, skipReason, _ := EvaluateItem(path, relPath, info, opts)

		item := ScannedItem{
			RelPath:    relPath,
			SourcePath: path,
			DestPath:   filepath.Join(destDir, relPath),
			Size:       info.Size(),
			ModTime:    info.ModTime(),
			Mode:       info.Mode(),
			IsDir:      d.IsDir(),
			NeedsCopy:  needsCopy,
			SkipReason: skipReason,
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case itemChan <- item:
		}

		return nil
	})

	return err
}

// Preview realiza uma varredura preliminar para estimar volumetria e quantidade de arquivos sem transferir dados.
func Preview(ctx context.Context, srcDir, destDir string, opts FilterOptions) (ScanSummary, error) {
	opts.DestinationDir = destDir
	var summary ScanSummary

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil
		}

		if path == srcDir {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if d.IsDir() {
			summary.TotalDirsDiscovered++
			return nil
		}

		summary.TotalFilesDiscovered++
		summary.TotalBytesDiscovered += info.Size()

		relPath, _ := filepath.Rel(srcDir, path)
		needsCopy, _, _ := EvaluateItem(path, relPath, info, opts)

		if needsCopy {
			summary.EligibleFilesCount++
			summary.EligibleBytesTotal += info.Size()
		} else {
			summary.SkippedFilesCount++
			summary.SkippedBytesTotal += info.Size()
		}

		return nil
	})

	return summary, err
}
