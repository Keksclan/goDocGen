package app

import (
	"fmt"
	"time"

	"pdf_generator/internal/converter"
	"pdf_generator/internal/models"
	"pdf_generator/internal/pdf"
	"pdf_generator/internal/scanner"
	"pdf_generator/internal/templates"
)

type App struct {
	Config models.Config
}

func NewApp(cfg models.Config) *App {
	return &App{Config: cfg}
}

func (a *App) Run(inputDir, outputFile string) error {
	fmt.Printf("Scanning directory: %s\n", inputDir)
	chapters, err := scanner.ScanDirectory(inputDir)
	if err != nil {
		return fmt.Errorf("error scanning directory: %w", err)
	}

	if len(chapters) == 0 {
		return fmt.Errorf("no markdown files found in %s", inputDir)
	}

	fmt.Printf("Found %d chapters.\n", len(chapters))

	doc := models.Document{
		Title:    a.Config.HeaderTitle, // Use header title as doc title if not specified
		Author:   "",                   // Could be added to config
		Date:     time.Now().Format("02.01.2006"),
		Chapters: chapters,
		Config:   a.Config,
	}

	if doc.Title == "" {
		doc.Title = "Documentation"
	}

	conv := converter.NewMarkdownConverter()

	var mdContents []string
	for _, ch := range chapters {
		mdContents = append(mdContents, ch.Content)
	}

	fmt.Println("Converting Markdown to HTML...")
	bodyHTML, err := conv.ConvertChapters(mdContents)
	if err != nil {
		return fmt.Errorf("error converting markdown: %w", err)
	}

	fullHTML := templates.GenerateHTML(doc, bodyHTML)
	headerTpl := templates.GetHeaderTemplate(doc)
	footerTpl := templates.GetFooterTemplate(doc)

	gen := pdf.NewGenerator()
	fmt.Printf("Generating PDF: %s\n", outputFile)
	err = gen.Generate(fullHTML, headerTpl, footerTpl, outputFile)
	if err != nil {
		return fmt.Errorf("error generating PDF: %w", err)
	}

	fmt.Println("PDF successfully generated!")
	return nil
}
