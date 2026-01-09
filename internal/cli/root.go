// Package cli implementiert die Kommandozeilen-Schnittstelle der Anwendung.
package cli

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd repräsentiert das Basis-Kommando ohne Argumente.
var rootCmd = &cobra.Command{
	Use:   "godocgen",
	Short: "goDocGen ist ein professioneller PDF-Generator aus Markdown",
	Long: `goDocGen ist ein Werkzeug zur Erzeugung professioneller PDF-Dokumentation.
Copyright (c) 2026 goDocGen Team. Alle Rechte vorbehalten.
Der kommerzielle Verkauf dieses Programms ist nicht gestattet.`,
}

// Execute fügt alle Kind-Kommandos zum Root-Kommando hinzu und setzt die Flags entsprechend.
// Dies wird von main.main() aufgerufen.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Initialisierungen für das Root-Kommando können hier vorgenommen werden.
}
