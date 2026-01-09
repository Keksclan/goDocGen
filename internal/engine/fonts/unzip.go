package fonts

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"godocgen/internal/util"
)

// ExtractFonts entpackt ein ZIP-Archiv mit Schriftarten in ein Cache-Verzeichnis.
// Es verwendet einen Hash des ZIP-Inhalts, um unnötiges erneutes Entpacken zu vermeiden.
func ExtractFonts(zipPath, cacheDir string) (string, error) {
	data, err := os.ReadFile(zipPath)
	if err != nil {
		return "", fmt.Errorf("Font-ZIP konnte nicht gelesen werden: %w", err)
	}

	hash := util.HashBytes(data)
	targetDir := filepath.Join(cacheDir, "fonts", hash)

	// Wenn das Verzeichnis bereits existiert, wurde es bereits entpackt
	if _, err := os.Stat(targetDir); err == nil {
		return targetDir, nil
	}

	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return "", fmt.Errorf("Font-Cache-Verzeichnis konnte nicht erstellt werden: %w", err)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("ZIP konnte nicht geöffnet werden: %w", err)
	}
	defer reader.Close()

	for _, f := range reader.File {
		path := filepath.Join(targetDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return "", err
		}

		file, err := f.Open()
		if err != nil {
			return "", err
		}

		dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			file.Close()
			return "", err
		}

		if _, err := io.Copy(dst, file); err != nil {
			file.Close()
			dst.Close()
			return "", err
		}
		file.Close()
		dst.Close()
	}

	return targetDir, nil
}
