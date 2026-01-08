package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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

		err := initializeProject(target)
		if err != nil {
			log.Fatalf("Initialization failed: %v", err)
		}
		fmt.Printf("Successfully initialized DocGen project in %s\n", target)
	},
}

func initializeProject(target string) error {
	// Create directories
	dirs := []string{
		"content",
		"assets",
		"fonts",
	}

	for _, d := range dirs {
		err := os.MkdirAll(filepath.Join(target, d), 0755)
		if err != nil {
			return err
		}
	}

	// Create docgen.yml
	configContent := `title: "My Documentation"
subtitle: "Generated with DocGen"
header:
  text: "Organization Name"
footer:
  text: "© 2026 Organization Name"
fonts:
  zip: "fonts/fonts.zip"
  regular: "Arial.ttf"
  bold: "Arial-Bold.ttf"
  italic: "Arial-Italic.ttf"
  mono: "Courier.ttf"
font_size: 12
layout:
  startpage: "center"
  body: "justify"
  margins:
    left: 10
    right: 10
    top: 10
    bottom: 10
code_theme: "catppuccin-latte"
`
	err := os.WriteFile(filepath.Join(target, "docgen.yml"), []byte(configContent), 0644)
	if err != nil {
		return err
	}

	// Create sample content
	sampleContent := `# Welcome to DocGen

This is your new documentation project.

## Features
- **Markdown** support
- **Mermaid** diagrams
- **Code** syntax highlighting
- **Customizable** layouts
`
	err = os.WriteFile(filepath.Join(target, "content", "01_intro.md"), []byte(sampleContent), 0644)
	return err
}

func init() {
	// init is added to root in root.go
}
