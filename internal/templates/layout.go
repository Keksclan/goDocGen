package templates

import (
	"fmt"
	"pdf_generator/internal/models"
)

func GenerateHTML(doc models.Document, bodyHTML string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <script>mermaid.initialize({startOnLoad:true});</script>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 0;
        }
        .title-page {
            height: 100vh;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            text-align: center;
            page-break-after: always;
        }
        .title-image {
            max-width: 80%%;
            max-height: 400px;
            margin-bottom: 50px;
        }
        .toc {
            page-break-after: always;
        }
        .chapter {
            padding: 20px;
        }
        .page-break {
            page-break-after: always;
        }
        pre {
            background-color: #272822;
            color: #f8f8f2;
            padding: 10px;
            border-radius: 5px;
            overflow-x: auto;
        }
        code {
            font-family: "Courier New", Courier, monospace;
        }
        
        /* Header and Footer styles for the PDF generator */
        #header-template, #footer-template {
            font-size: 10px;
            color: #555;
            width: 100%%;
            padding: 0 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .logo {
            height: 20px;
        }
    </style>
</head>
<body>
    <div class="title-page">
        %s
        <h1>%s</h1>
        <h3>%s</h3>
        <p>%s</p>
    </div>

    %s

    <div class="content">
        %s
    </div>
</body>
</html>
`, doc.Title, renderTitleImage(doc.Config.TitleImage), doc.Title, doc.Author, doc.Date, renderTOC(doc), bodyHTML)
}

func renderTitleImage(img string) string {
	if img == "" {
		return ""
	}
	return fmt.Sprintf(`<img src="%s" class="title-image">`, img)
}

func renderTOC(doc models.Document) string {
	if !doc.Config.ShowTOC {
		return ""
	}
	html := `<div class="toc"><h1>Inhaltsverzeichnis</h1><ul>`
	for _, ch := range doc.Chapters {
		html += fmt.Sprintf(`<li>%s</li>`, ch.Title)
	}
	html += `</ul></div>`
	return html
}

func GetHeaderTemplate(doc models.Document) string {
	return fmt.Sprintf(`
<div id="header-template">
    <div>%s</div>
    <div>%s</div>
    %s
</div>`, doc.Config.HeaderTitle, doc.Title, renderLogo(doc.Config.HeaderLogo))
}

func GetFooterTemplate(doc models.Document) string {
	return fmt.Sprintf(`
<div id="footer-template">
    <div>%s</div>
    <div>Seite <span class="pageNumber"></span> von <span class="totalPages"></span></div>
    %s
</div>`, doc.Config.FooterText, renderLogo(doc.Config.FooterLogo))
}

func renderLogo(logo string) string {
	if logo == "" {
		return ""
	}
	return fmt.Sprintf(`<img src="%s" class="logo">`, logo)
}
