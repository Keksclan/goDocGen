package cli

import (
	"log"

	"godocgen/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new documentation project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}

		p := tea.NewProgram(tui.InitialInitModel(target))
		if _, err := p.Run(); err != nil {
			log.Fatalf("TUI error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

