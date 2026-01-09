// Package util enthält Hilfsfunktionen für verschiedene Zwecke.
package util

import (
	"bytes"
	"fmt"
	"os/exec"
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
