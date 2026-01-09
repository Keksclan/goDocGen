package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Gibt die aktuelle Version von goDocGen aus",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("goDocGen version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
