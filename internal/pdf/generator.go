package pdf

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Generator struct {
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(htmlContent string, headerTemplate, footerTemplate string, outputPath string) error {
	// Create temporary HTML file
	tmpFile, err := ioutil.TempFile("", "doc-*.html")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(htmlContent)); err != nil {
		return err
	}
	tmpFile.Close()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate("file:///"+tmpFile.Name()),
		// Wait for mermaid to render
		chromedp.Sleep(2*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, _, err = page.PrintToPDF().
				WithDisplayHeaderFooter(true).
				WithHeaderTemplate(headerTemplate).
				WithFooterTemplate(footerTemplate).
				WithPrintBackground(true).
				WithMarginTop(1).
				WithMarginBottom(1).
				WithMarginLeft(0.5).
				WithMarginRight(0.5).
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(outputPath, buf, 0644)
}
