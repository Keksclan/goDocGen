// Package fonts bietet Funktionen zum Herunterladen und Extrahieren von Schriftarten.
package fonts

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"godocgen/internal/util"
)

// DownloadFonts lädt eine ZIP-Datei mit Schriftarten von der angegebenen URL herunter
// und speichert sie im Cache-Verzeichnis.
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
		return "", fmt.Errorf("Download-Verzeichnis konnte nicht erstellt werden: %w", err)
	}

	// Download ausführen
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Fehler beim Herunterladen der Schriften: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Fehler beim Herunterladen der Schriften: Statuscode %d", resp.StatusCode)
	}

	// In Datei speichern
	out, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("ZIP-Datei konnte nicht erstellt werden: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("ZIP-Datei konnte nicht gespeichert werden: %w", err)
	}

	return zipPath, nil
}
