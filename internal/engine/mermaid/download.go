package mermaid

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const MermaidJSURL = "https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"

// EnsureMermaidJS stellt sicher, dass die mermaid.min.js lokal vorhanden ist.
func EnsureMermaidJS(cacheDir string) (string, error) {
	jsPath := filepath.Join(cacheDir, "mermaid", "mermaid.min.js")

	// Prüfe, ob die Datei bereits existiert
	if _, err := os.Stat(jsPath); err == nil {
		return jsPath, nil
	}

	// Verzeichnis erstellen
	if err := os.MkdirAll(filepath.Dir(jsPath), 0755); err != nil {
		return "", fmt.Errorf("Mermaid-Cache-Verzeichnis konnte nicht erstellt werden: %w", err)
	}

	fmt.Printf("Lade Mermaid JS herunter (%s)...\n", MermaidJSURL)

	// Download ausführen
	resp, err := http.Get(MermaidJSURL)
	if err != nil {
		return "", fmt.Errorf("Fehler beim Herunterladen von Mermaid JS: %w (Internetverbindung erforderlich)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Fehler beim Herunterladen von Mermaid JS: Statuscode %d", resp.StatusCode)
	}

	// In Datei speichern
	out, err := os.Create(jsPath)
	if err != nil {
		return "", fmt.Errorf("Mermaid JS-Datei konnte nicht erstellt werden: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("Mermaid JS-Datei konnte nicht gespeichert werden: %w", err)
	}

	return jsPath, nil
}

// EnsureMmdc stellt sicher, dass die Mermaid CLI (mmdc) vorhanden ist und funktioniert.
func EnsureMmdc(cacheDir string) (string, error) {
	// 1. Prüfen, ob mmdc bereits im PATH ist und funktioniert
	if path, err := exec.LookPath("mmdc"); err == nil {
		if checkMmdc(path) {
			return path, nil
		}
	}

	// 2. Prüfen, ob mmdc im lokalen Cache/bin Verzeichnis ist und funktioniert
	localMmdc := filepath.Join(cacheDir, "mermaid", "node_modules", ".bin", "mmdc")
	if runtime.GOOS == "windows" {
		localMmdc += ".cmd"
	}

	if _, err := os.Stat(localMmdc); err == nil {
		if checkMmdc(localMmdc) {
			return localMmdc, nil
		}
		// Falls es da ist, aber nicht funktioniert, löschen wir das node_modules Verzeichnis
		// um eine saubere Neuinstallation zu ermöglichen
		fmt.Println("Lokale mmdc-Installation scheint defekt zu sein. Versuche Neuinstallation...")
		os.RemoveAll(filepath.Join(cacheDir, "mermaid", "node_modules"))
	}

	// 3. Falls nicht, versuchen via npm zu installieren
	fmt.Println("mmdc (Mermaid CLI) wird nicht gefunden oder ist defekt. Versuche Installation via npm...")

	_, err := exec.LookPath("npm")
	if err != nil {
		return "", fmt.Errorf("npm ist nicht installiert. Bitte installieren Sie Node.js oder mmdc manuell (npm install -g @mermaid-js/mermaid-cli)")
	}

	// In den Cache-Ordner installieren
	mermaidDir := filepath.Join(cacheDir, "mermaid")
	if err := os.MkdirAll(mermaidDir, 0755); err != nil {
		return "", err
	}

	// package.json erstellen, falls nicht vorhanden
	pkgJson := filepath.Join(mermaidDir, "package.json")
	if _, err := os.Stat(pkgJson); os.IsNotExist(err) {
		content := `{"name": "mermaid-renderer", "version": "1.0.0"}`
		_ = os.WriteFile(pkgJson, []byte(content), 0644)
	}

	// npm install ausführen
	// Wir setzen PUPPETEER_SKIP_DOWNLOAD, um SSL-Fehler beim Browser-Download zu vermeiden.
	// mmdc kann oft auch den installierten Chrome nutzen oder wir fallen auf ChromeDP zurück.
	cmd := exec.Command("npm", "install", "@mermaid-js/mermaid-cli")
	cmd.Dir = mermaidDir
	cmd.Env = append(os.Environ(),
		"PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true",
		"PUPPETEER_SKIP_DOWNLOAD=true",
	)
	cmd.Stdout = io.Discard // Weniger Rauschen, wir loggen oben schon
	cmd.Stderr = io.Discard

	fmt.Println("Führe 'npm install @mermaid-js/mermaid-cli' aus... Dies kann einen Moment dauern.")
	err = cmd.Run()
	if err != nil {
		// Bei Fehler aufräumen
		os.RemoveAll(filepath.Join(mermaidDir, "node_modules"))
		return "", fmt.Errorf("npm install fehlgeschlagen: %w", err)
	}

	if _, err := os.Stat(localMmdc); err == nil {
		if checkMmdc(localMmdc) {
			return localMmdc, nil
		}
		return "", fmt.Errorf("mmdc wurde installiert, ist aber nicht funktionsfähig (Check fehlgeschlagen)")
	}

	return "", fmt.Errorf("mmdc wurde installiert, aber die ausführbare Datei wurde nicht unter %s gefunden", localMmdc)
}

// checkMmdc führt einen einfachen Funktionstest für mmdc aus.
func checkMmdc(path string) bool {
	cmd := exec.Command(path, "--version")
	err := cmd.Run()
	return err == nil
}
