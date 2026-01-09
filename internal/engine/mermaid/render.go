// Package mermaid bietet Funktionen zum Rendern von Mermaid-Diagrammen in SVG und PNG.
package mermaid

import (
	"context"
	"godocgen/internal/util"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

// Render versucht ein Mermaid-Diagramm zu rendern.
// Es nutzt mmdc (Mermaid CLI), falls installiert, ansonsten erfolgt ein Fallback auf ChromeDP.
func Render(content string, cacheDir string) (string, string, error) {
	hash := util.HashString(content)
	svgPath := filepath.Join(cacheDir, "mermaid", hash+".svg")
	pngPath := filepath.Join(cacheDir, "mermaid", hash+".png")

	// Cache-Prüfung
	if _, err := os.Stat(svgPath); err == nil {
		if _, err := os.Stat(pngPath); err == nil {
			return svgPath, pngPath, nil
		}
	}

	os.MkdirAll(filepath.Dir(svgPath), 0755)

	// Versuche mmdc (schneller und bessere Qualität)
	err := renderWithMmdc(content, svgPath, pngPath)
	if err == nil {
		return svgPath, pngPath, nil
	}

	// Fallback auf ChromeDP (benötigt installierten Chrome/Chromium)
	fmt.Printf("Warnung: mmdc fehlgeschlagen oder nicht installiert, nutze ChromeDP für Mermaid: %v\n", err)
	err = renderWithChrome(content, pngPath)
	if err != nil {
		return "", "", fmt.Errorf("Mermaid-Rendering fehlgeschlagen (mmdc und chromedp): %w", err)
	}

	// ChromeDP liefert nur PNG, wir nutzen es für beide Pfade als Fallback
	return pngPath, pngPath, nil
}

// renderWithMmdc nutzt die Mermaid CLI (mmdc) zum Rendern.
func renderWithMmdc(content string, svgPath, pngPath string) error {
	hash := util.HashString(content)
	tmpFile := filepath.Join(os.TempDir(), hash+".mmd")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	_, err = util.RunCommand("mmdc", "-i", tmpFile, "-o", svgPath, "-b", "transparent")
	if err != nil {
		return err
	}

	_, err = util.RunCommand("mmdc", "-i", tmpFile, "-o", pngPath, "-b", "transparent", "--scale", "3")
	return err
}

// renderWithChrome nutzt einen headless Browser (via ChromeDP) zum Rendern.
func renderWithChrome(content string, outputPath string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <script>
        mermaid.initialize({ startOnLoad: true, theme: 'default' });
    </script>
    <style>
        body { margin: 0; background: transparent; }
        #container { display: inline-block; padding: 10px; background: white; }
    </style>
</head>
<body>
    <div id="container">
        <div class="mermaid">
%s
        </div>
    </div>
</body>
</html>
`, content)

	tmpFile := filepath.Join(os.TempDir(), util.HashString(content)+".html")
	if err := os.WriteFile(tmpFile, []byte(html), 0644); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("file://"+tmpFile),
		chromedp.WaitVisible(".mermaid svg", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Screenshot("#container", &buf, chromedp.ByID),
	)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, buf, 0644)
}
