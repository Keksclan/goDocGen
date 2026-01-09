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
	sectionDesign   string
	sectionGradient string
	sectionFonts    string
	sectionTOC      string
	tocEnabled      string
	tocNumbers      string
	tocDots         string
	headerNumbering string
	footerLeft      string
	footerCenter    string
	footerRight     string
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
		sectionLayout:   "📏 Layout & Font",
		sectionHeader:   "🔝 Header & Footer",
		sectionDesign:   "📑 Seitenzahlen & Design",
		sectionGradient: "🌈 Gradient (Farbverlauf)",
		sectionFonts:    "🔡 Fonts & Mermaid",
		sectionTOC:      "📑 Inhaltsverzeichnis (TOC)",
		tocEnabled:      "TOC Aktiviert",
		tocNumbers:      "TOC Nummern",
		tocDots:         "TOC Punkte",
		headerNumbering: "Header Nummerierung",
		footerLeft:      "Footer Links",
		footerCenter:    "Footer Mitte",
		footerRight:     "Footer Rechts",
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
		sectionLayout:   "📏 Layout & Font",
		sectionHeader:   "🔝 Header & Footer",
		sectionDesign:   "📑 Page Numbers & Design",
		sectionGradient: "🌈 Gradient (Background)",
		sectionFonts:    "🔡 Fonts & Mermaid",
		sectionTOC:      "📑 Table of Contents (TOC)",
		tocEnabled:      "TOC Enabled",
		tocNumbers:      "TOC Numbers",
		tocDots:         "TOC Dots",
		headerNumbering: "Header Numbering",
		footerLeft:      "Footer Left",
		footerCenter:    "Footer Center",
		footerRight:     "Footer Right",
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
