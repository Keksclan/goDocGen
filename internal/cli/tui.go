package cli

import (
	"fmt"
	"godocgen/internal/config"
	"godocgen/internal/tui"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// tuiCmd repräsentiert den Befehl zum Starten der interaktiven TUI.
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Startet das interaktive TUI",
	Run: func(cmd *cobra.Command, args []string) {
		// Prüfe ob im aktuellen Verzeichnis ein Projekt ist
		if _, err := os.Stat("docgen.yml"); err == nil {
			gCfg, _ := config.LoadGlobalConfig()
			absPath, _ := filepath.Abs(".")
			gCfg.AddProject(absPath)
			_ = gCfg.Save()
		}

		p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Ein Fehler ist aufgetreten: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
