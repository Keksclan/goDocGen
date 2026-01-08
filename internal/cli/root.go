package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "godocgen",
	Short: "goDocGen is a professional PDF generator from Markdown",
	Long: `goDocGen ist ein Werkzeug zur Erzeugung professioneller PDF-Dokumentation.
Copyright (c) 2026 goDocGen Team. Alle Rechte vorbehalten.
Der kommerzielle Verkauf dieses Programms ist nicht gestattet.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
