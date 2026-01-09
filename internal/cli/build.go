package cli

import (
	"godocgen/internal/config"
	"godocgen/internal/engine"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var projectDir string
var outDir string
var configName string

// buildCmd repräsentiert den Befehl zum Generieren eines PDFs aus einem Projekt.
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Erzeugt ein PDF aus dem Projekt",
	Run: func(cmd *cobra.Command, args []string) {
		builder := engine.NewBuilder(projectDir, outDir)
		builder.ConfigName = configName
		path, err := builder.Build()
		if err != nil {
			log.Fatalf("Build fehlgeschlagen: %v", err)
		}

		// Speichere das Projekt in der Historie
		gCfg, _ := config.LoadGlobalConfig()
		gCfg.AddProject(projectDir)
		_ = gCfg.Save()

		fmt.Printf("PDF erfolgreich generiert: %s\n", path)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&projectDir, "project", "p", ".", "Project directory")
	buildCmd.Flags().StringVarP(&outDir, "out", "o", "./dist", "Output directory")
	buildCmd.Flags().StringVarP(&configName, "config", "c", "docgen.yml", "Config file name")
}
