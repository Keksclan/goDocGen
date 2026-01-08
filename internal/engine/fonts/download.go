package fonts

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"godocgen/internal/util"
)

func DownloadFonts(url, cacheDir string) (string, error) {
	// Erstelle einen Hash der URL, um einen eindeutigen Dateinamen zu haben
	urlHash := util.HashString(url)
	zipPath := filepath.Join(cacheDir, "downloads", urlHash+".zip")

	// Prüfe, ob die Datei bereits existiert
	if _, err := os.Stat(zipPath); err == nil {
		return zipPath, nil
	}

	// Verzeichnis erstellen
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		return "", fmt.Errorf("could not create download dir: %w", err)
	}

	// Download
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download fonts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download fonts: status code %d", resp.StatusCode)
	}

	// In Datei speichern
	out, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("could not create zip file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save zip file: %w", err)
	}

	return zipPath, nil
}

