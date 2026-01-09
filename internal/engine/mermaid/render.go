// Package mermaid bietet Funktionen zum Rendern von Mermaid-Diagrammen in SVG und PNG.
package mermaid

import (
	"context"
	"encoding/json"
	"godocgen/internal/util"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	mmdcPath, err := EnsureMmdc(cacheDir)
	if err == nil {
		err = renderWithMmdc(mmdcPath, content, svgPath, pngPath)
		if err == nil {
			return svgPath, pngPath, nil
		}
	}

	// Fallback auf ChromeDP (benötigt installierten Chrome/Chromium)
	fmt.Printf("Warnung: mmdc fehlgeschlagen oder nicht installiert, nutze ChromeDP für Mermaid: %v\n", err)
	err = renderWithChrome(content, pngPath, cacheDir)
	if err != nil {
		return "", "", fmt.Errorf("Mermaid-Rendering fehlgeschlagen (mmdc und chromedp): %w", err)
	}

	// ChromeDP liefert nur PNG, wir nutzen es für beide Pfade als Fallback
	return pngPath, pngPath, nil
}

// renderWithMmdc nutzt die Mermaid CLI (mmdc) zum Rendern.
func renderWithMmdc(mmdcPath, content string, svgPath, pngPath string) error {
	hash := util.HashString(content)
	tmpFile := filepath.Join(os.TempDir(), hash+".mmd")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	_, err = util.RunCommand(mmdcPath, "-i", tmpFile, "-o", svgPath, "-b", "transparent")
	if err != nil {
		return err
	}

	_, err = util.RunCommand(mmdcPath, "-i", tmpFile, "-o", pngPath, "-b", "transparent", "--scale", "8")
	return err
}

// renderWithChrome nutzt einen headless Browser (via ChromeDP) zum Rendern.
func renderWithChrome(content string, outputPath string, cacheDir string) error {
	// Lokales Mermaid JS sicherstellen
	jsPath, err := EnsureMermaidJS(cacheDir)
	var scriptContent []byte
	if err == nil {
		scriptContent, _ = os.ReadFile(jsPath)
	}

	jsScript := ""
	if len(scriptContent) > 0 {
		jsScript = "<script>" + string(scriptContent) + "</script>"
	} else {
		jsScript = `<script src="https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"></script>`
	}

	encodedContent, _ := json.Marshal(content)
	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    {{MERMAID_JS}}
    <style>
        body { margin: 0; background: transparent; overflow: hidden; }
        #container { display: inline-block; padding: 20px; background: white; }
        svg { display: block; }
    </style>
</head>
<body>
    <div id="container"></div>
    <script>
        async function render() {
            try {
                mermaid.initialize({ startOnLoad: false, theme: 'default' });
                const { svg } = await mermaid.render('mermaid-svg', {{CONTENT}});
                document.getElementById('container').innerHTML = svg;
                window.mermaidReady = true;
            } catch (e) {
                document.getElementById('container').innerHTML = 'Error: ' + e.message;
                window.mermaidError = e.message;
            }
        }
        render();
    </script>
</body>
</html>
`
	html = strings.Replace(html, "{{MERMAID_JS}}", jsScript, 1)
	html = strings.Replace(html, "{{CONTENT}}", string(encodedContent), 1)

	tmpFile := filepath.Join(os.TempDir(), util.HashString(content)+".html")
	if err := os.WriteFile(tmpFile, []byte(html), 0644); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	// ChromeDP Optionen
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("force-device-scale-factor", "4"), // Erhöhe Pixeldichte für schärfere Screenshots
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate("file://"+tmpFile),
		chromedp.WaitVisible("#container svg", chromedp.ByQuery),
		// Zusätzliche Zeit für finales Rendering (manchmal braucht Mermaid nach dem Einfügen des SVGs noch kurz)
		chromedp.Sleep(1000*time.Millisecond),
		chromedp.Screenshot("#container", &buf, chromedp.ByID),
	); err != nil {
		return fmt.Errorf("ChromeDP Fehler: %w (Prüfen Sie Ihre Internetverbindung)", err)
	}

	if len(buf) == 0 {
		return fmt.Errorf("Diagramm-Rendering ergab ein leeres Bild")
	}

	return os.WriteFile(outputPath, buf, 0644)
}
