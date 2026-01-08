package cli

import (
	"godocgen/internal/engine"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var projectDir string
var outDir string
var configName string

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build PDF from project",
	Run: func(cmd *cobra.Command, args []string) {
		builder := engine.NewBuilder(projectDir, outDir)
		builder.ConfigName = configName
		err := builder.Build()
		if err != nil {
			log.Fatalf("Build failed: %v", err)
		}
		fmt.Printf("Successfully generated PDF in %s\n", outDir)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&projectDir, "project", "p", ".", "Project directory")
	buildCmd.Flags().StringVarP(&outDir, "out", "o", "./dist", "Output directory")
	buildCmd.Flags().StringVarP(&configName, "config", "c", "docgen.yml", "Config file name")
}
