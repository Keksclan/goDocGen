package cli

import (
	"godocgen/internal/config"
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
)

// validateCmd repräsentiert den Befehl zum Validieren der Projektkonfiguration.
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validiert die Projektkonfiguration",
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath := filepath.Join(projectDir, "docgen.yml")
		_, err := config.LoadConfig(cfgPath)
		if err != nil {
			log.Fatalf("Validierung fehlgeschlagen: %v", err)
		}
		fmt.Println("Konfiguration ist gültig.")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&projectDir, "project", "p", ".", "Project directory")
}
