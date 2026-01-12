package tui

type translation struct {
	helpTab         string
	configTab       string
	themeTab        string
	actionsTab      string
	projectsTab     string
	initTab         string
	title           string
	subtitle        string
	author          string
	theme           string
	codeTheme       string
	fontSize        string
	margins         string
	save            string
	buildPdf        string
	openPdf         string
	downloadFont    string
	addToPath       string
	project         string
	noProject       string
	statusSaved     string
	statusError     string
	statusPathAdded string
	helpHeader      string
	langSwitch      string
	marginLabel     string
	pressEToEdit    string
	editMode        string
	sectionBasis    string
	sectionLayout   string
	sectionHeader   string
	sectionFooter   string
	sectionDesign   string
	sectionGradient string
	sectionFonts    string
	sectionTOC      string
	sectionCode     string
	sectionColors   string
	sectionMermaid  string
	tocEnabled      string
	tocNumbers      string
	tocDots         string
	tocLineSpacing  string
	tocBoldHeadings string
	tocFontSize     string
	tocIndent       string
	headerNumbering string
	footerLeft      string
	footerCenter    string
	footerRight     string
	footerStyle     string
	codeFontSize    string
	codeMinFontSize string
	codeAutoScale   string
	codeMaxLines    string
	codeMaxLineLen  string
	colorTitle      string
	colorHeader     string
	colorBackground string
	colorText       string
	colorAccent     string
	mermaidRenderer string
	mermaidWidth    string
	mermaidScale    string
}

var translations = map[string]translation{
	"de": {
		helpTab:         "📖 Hilfe",
		configTab:       "⚙️ Konfig",
		themeTab:        "🎨 Themes",
		actionsTab:      "⚡ Aktionen",
		projectsTab:     "📂 Projekte",
		initTab:         "🏗️ Init",
		title:           "Titel",
		subtitle:        "Untertitel",
		author:          "Autor",
		theme:           "Theme",
		codeTheme:       "Code-Theme",
		fontSize:        "Schriftgröße",
		margins:         "Seitenränder (mm)",
		save:            "Speichern",
		buildPdf:        "📄 PDF generieren",
		openPdf:         "📖 Letztes PDF öffnen",
		downloadFont:    "📥 Fonts herunterladen",
		addToPath:       "🚀 Zum PATH hinzufügen",
		project:         "Projekt",
		noProject:       "Kein Projekt geladen",
		statusSaved:     "Konfiguration gespeichert!",
		statusError:     "Fehler",
		statusPathAdded: "Erfolgreich zum PATH hinzugefügt!",
		helpHeader:      "📘 goDocGen - Hilfe & Funktionen",
		langSwitch:      "Sprache wechseln (DE/EN)",
		marginLabel:     "L:%v R:%v T:%v B:%v",
		pressEToEdit:    "Drücke 'E' zum Bearbeiten",
		editMode:        "BEARBEITUNGS-MODUS",
		sectionBasis:    "📁 Basis Information",
		sectionLayout:   "📏 Layout & Abstände",
		sectionHeader:   "🔝 Header",
		sectionFooter:   "🔻 Footer",
		sectionDesign:   "🎨 Design & Theme",
		sectionGradient: "🌈 Gradient (Farbverlauf)",
		sectionFonts:    "🔡 Schriftarten",
		sectionTOC:      "📑 Inhaltsverzeichnis",
		sectionCode:     "💻 Code-Blöcke",
		sectionColors:   "🎨 Farben",
		sectionMermaid:  "📊 Mermaid-Diagramme",
		tocEnabled:      "TOC Aktiviert",
		tocNumbers:      "Nummern anzeigen",
		tocDots:         "Punkte anzeigen",
		tocLineSpacing:  "Zeilenabstand",
		tocBoldHeadings: "Fett darstellen",
		tocFontSize:     "Schriftgröße",
		tocIndent:       "Einrückung (mm)",
		headerNumbering: "Header Nummerierung",
		footerLeft:      "Links",
		footerCenter:    "Mitte",
		footerRight:     "Rechts",
		footerStyle:     "Style (fixed/inline)",
		codeFontSize:    "Schriftgröße",
		codeMinFontSize: "Min. Schriftgröße",
		codeAutoScale:   "Auto-Skalierung",
		codeMaxLines:    "Max. Zeilen",
		codeMaxLineLen:  "Max. Zeilenlänge",
		colorTitle:      "Überschriften",
		colorHeader:     "Header-Text",
		colorBackground: "Hintergrund",
		colorText:       "Text",
		colorAccent:     "Akzent",
		mermaidRenderer: "Renderer",
		mermaidWidth:    "Breite (mm)",
		mermaidScale:    "Skalierung",
	},
	"en": {
		helpTab:         "📖 Help",
		configTab:       "⚙️ Config",
		themeTab:        "🎨 Themes",
		actionsTab:      "⚡ Actions",
		projectsTab:     "📂 Projects",
		initTab:         "🏗️ Init",
		title:           "Title",
		subtitle:        "Subtitle",
		author:          "Author",
		theme:           "Theme",
		codeTheme:       "Code Theme",
		fontSize:        "Font Size",
		margins:         "Margins (mm)",
		save:            "Save",
		buildPdf:        "📄 Generate PDF",
		openPdf:         "📖 Open latest PDF",
		downloadFont:    "📥 Download Fonts",
		addToPath:       "🚀 Add to PATH",
		project:         "Project",
		noProject:       "No project loaded",
		statusSaved:     "Configuration saved!",
		statusError:     "Error",
		statusPathAdded: "Successfully added to PATH!",
		helpHeader:      "📘 goDocGen - Help & Features",
		langSwitch:      "Switch Language (DE/EN)",
		marginLabel:     "L:%v R:%v T:%v B:%v",
		pressEToEdit:    "Press 'E' to edit",
		editMode:        "EDIT MODE",
		sectionBasis:    "📁 Basic Information",
		sectionLayout:   "📏 Layout & Spacing",
		sectionHeader:   "🔝 Header",
		sectionFooter:   "🔻 Footer",
		sectionDesign:   "🎨 Design & Theme",
		sectionGradient: "🌈 Gradient (Background)",
		sectionFonts:    "🔡 Fonts",
		sectionTOC:      "📑 Table of Contents",
		sectionCode:     "💻 Code Blocks",
		sectionColors:   "🎨 Colors",
		sectionMermaid:  "📊 Mermaid Diagrams",
		tocEnabled:      "TOC Enabled",
		tocNumbers:      "Show Numbers",
		tocDots:         "Show Dots",
		tocLineSpacing:  "Line Spacing",
		tocBoldHeadings: "Bold Headings",
		tocFontSize:     "Font Size",
		tocIndent:       "Indent (mm)",
		headerNumbering: "Header Numbering",
		footerLeft:      "Left",
		footerCenter:    "Center",
		footerRight:     "Right",
		footerStyle:     "Style (fixed/inline)",
		codeFontSize:    "Font Size",
		codeMinFontSize: "Min. Font Size",
		codeAutoScale:   "Auto Scale",
		codeMaxLines:    "Max. Lines",
		codeMaxLineLen:  "Max. Line Length",
		colorTitle:      "Headings",
		colorHeader:     "Header Text",
		colorBackground: "Background",
		colorText:       "Text",
		colorAccent:     "Accent",
		mermaidRenderer: "Renderer",
		mermaidWidth:    "Width (mm)",
		mermaidScale:    "Scale",
	},
}

func (m *model) T(key func(translation) string) string {
	lang := "de"
	if m.history != nil && m.history.Language != "" {
		lang = m.history.Language
	}
	t, ok := translations[lang]
	if !ok {
		t = translations["de"]
	}
	return key(t)
}
