package cli

import (
	"godocgen/internal/config"
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate project configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath := filepath.Join(projectDir, "docgen.yml")
		_, err := config.LoadConfig(cfgPath)
		if err != nil {
			log.Fatalf("Validation failed: %v", err)
		}
		fmt.Println("Configuration is valid.")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&projectDir, "project", "p", ".", "Project directory")
}
