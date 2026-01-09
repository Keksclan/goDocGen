// Package util enthält Hilfsfunktionen für verschiedene Zwecke.
package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// RunCommand führt einen externen Systembefehl aus und gibt die Standardausgabe zurück.
// Im Fehlerfall wird auch der Inhalt von Stderr zurückgegeben.
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Befehl fehlgeschlagen: %s %v\nFehler: %w\nStderr: %s", name, args, err, stderr.String())
	}

	return stdout.String(), nil
}

// OpenPath öffnet einen Pfad mit dem Standardprogramm des Betriebssystems.
func OpenPath(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // Linux und andere
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}

// AddToPath fügt das Verzeichnis der aktuellen Ausführbaren Datei zum System-PATH hinzu.
func AddToPath() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(exePath)

	switch runtime.GOOS {
	case "windows":
		return addToPathWindows(dir)
	case "darwin":
		return addToPathMac(dir)
	default:
		return fmt.Errorf("Betriebssystem %s wird für 'AddToPath' nicht unterstützt", runtime.GOOS)
	}
}

func addToPathWindows(dir string) error {
	// Nutze PowerShell um den PATH permanent für den User zu setzen
	script := fmt.Sprintf(`
$oldPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($oldPath -split ";" -notcontains "%s") {
    $newPath = "$oldPath;%s".Trim(';')
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    return "OK"
}
return "ALREADY"
`, dir, dir)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PowerShell Fehler: %w, Output: %s", err, string(output))
	}

	return nil
}

func addToPathMac(dir string) error {
	// Auf Mac fügen wir es zur .zshrc hinzu (Standard-Shell)
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	zshrc := filepath.Join(home, ".zshrc")
	line := fmt.Sprintf("\nexport PATH=\"$PATH:%s\"\n", dir)

	// Prüfen ob es schon drin ist
	content, _ := os.ReadFile(zshrc)
	if strings.Contains(string(content), dir) {
		return nil // Schon vorhanden
	}

	f, err := os.OpenFile(zshrc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(line)
	return err
}
