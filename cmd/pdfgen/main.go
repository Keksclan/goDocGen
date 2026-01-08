package main

import (
	"fmt"
	"log"
	"os"

	"pdf_generator/internal/app"
	"pdf_generator/internal/models"

	"github.com/spf13/cobra"
)

var (
	inputDir    string
	outputFile  string
	title       string
	author      string
	headerTitle string
	footerText  string
	titleImage  string
	headerLogo  string
	footerLogo  string
	showTOC     bool
)

var rootCmd = &cobra.Command{
	Use:   "pdfgen",
	Short: "A markdown to PDF generator with style",
	Run: func(cmd *cobra.Command, args []string) {
		generatePDF()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&inputDir, "input", "i", ".", "Input directory containing markdown files")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "output.pdf", "Output PDF file path")
	rootCmd.Flags().StringVarP(&title, "title", "t", "Documentation", "Document title")
	rootCmd.Flags().StringVarP(&author, "author", "a", "", "Author name")
	rootCmd.Flags().StringVar(&headerTitle, "header-title", "", "Header title")
	rootCmd.Flags().StringVar(&footerText, "footer-text", "", "Footer text")
	rootCmd.Flags().StringVar(&titleImage, "title-image", "", "Path to title image")
	rootCmd.Flags().StringVar(&headerLogo, "header-logo", "", "Path to header logo image")
	rootCmd.Flags().StringVar(&footerLogo, "footer-logo", "", "Path to footer logo image")
	rootCmd.Flags().BoolVar(&showTOC, "toc", true, "Show Table of Contents")
}

func generatePDF() {
	cfg := models.Config{
		HeaderTitle: headerTitle,
		FooterText:  footerText,
		TitleImage:  titleImage,
		HeaderLogo:  headerLogo,
		FooterLogo:  footerLogo,
		ShowTOC:     showTOC,
	}

	application := app.NewApp(cfg)
	err := application.Run(inputDir, outputFile)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
