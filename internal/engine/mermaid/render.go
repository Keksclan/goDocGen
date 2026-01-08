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

func Render(content string, cacheDir string) (string, string, error) {
	hash := util.HashString(content)
	svgPath := filepath.Join(cacheDir, "mermaid", hash+".svg")
	pngPath := filepath.Join(cacheDir, "mermaid", hash+".png")

	if _, err := os.Stat(svgPath); err == nil {
		if _, err := os.Stat(pngPath); err == nil {
			return svgPath, pngPath, nil
		}
	}

	os.MkdirAll(filepath.Dir(svgPath), 0755)

	// Try mmdc first
	err := renderWithMmdc(content, svgPath, pngPath)
	if err == nil {
		return svgPath, pngPath, nil
	}

	// Fallback to ChromeDP
	fmt.Printf("Warning: mmdc failed or not found, falling back to ChromeDP for Mermaid: %v\n", err)
	err = renderWithChrome(content, pngPath)
	if err != nil {
		return "", "", fmt.Errorf("mermaid rendering failed (mmdc and chromedp): %w", err)
	}

	// We don't have SVG from ChromeDP easily, so we just return the same path for both or empty SVG
	return pngPath, pngPath, nil
}

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

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("file://"+tmpFile),
		chromedp.WaitVisible(".mermaid svg", chromedp.ByQuery),
		// Give it a moment to stabilize
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Screenshot("#container", &buf, chromedp.ByID),
	)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, buf, 0644)
}

