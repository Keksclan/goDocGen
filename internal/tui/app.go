package tui

import (
	"godocgen/internal/config"
	"godocgen/internal/engine"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateHelp state = iota
	stateConfig
	stateActions
	stateInit
)

type model struct {
	state          state
	width          int
	height         int
	keys           keyMap
	help           help.Model
	cfg            *config.Config
	err            error
	selectedAction int
	statusMsg      string
	initModel      InitModel
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Help   key.Binding
	Quit   key.Binding
	Action key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Help, k.Quit, k.Action},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "hoch"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "runter"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "vorheriger Tab"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "nächster Tab"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "hilfe"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "beenden"),
	),
	Action: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "ausführen"),
	),
}

func InitialModel() model {
	cfg, _ := config.LoadConfig("docgen.yml")
	return model{
		state:     stateHelp,
		keys:      keys,
		help:      help.New(),
		cfg:       cfg,
		initModel: InitialInitModel("."),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.state == stateInit {
		newInitModel, initCmd := m.initModel.Update(msg)
		m.initModel = newInitModel.(InitModel)
		if initCmd != nil {
			return m, initCmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Right):
			m.state = (m.state + 1) % 4
			m.statusMsg = ""
		case key.Matches(msg, m.keys.Left):
			m.state = (m.state + 3) % 4
			m.statusMsg = ""
		case key.Matches(msg, m.keys.Up):
			if m.state == stateActions {
				m.selectedAction = (m.selectedAction + 1) % 2
			}
		case key.Matches(msg, m.keys.Down):
			if m.state == stateActions {
				m.selectedAction = (m.selectedAction + 1) % 2
			}
		case key.Matches(msg, m.keys.Action):
			if m.state == stateActions {
				return m, m.performAction()
			}
			if m.state == stateConfig && m.cfg != nil {
				// Theme toggle als Beispiel für "settings treffen"
				if m.cfg.Theme == "catppuccin-latte" {
					m.cfg.Theme = "catppuccin-mocha"
				} else {
					m.cfg.Theme = "catppuccin-latte"
				}
				m.statusMsg = fmt.Sprintf("Theme auf %s geändert (nur für diese Sitzung)", m.cfg.Theme)
			}
		}
	case actionResultMsg:
		m.statusMsg = string(msg)
	}
	return m, cmd
}

type actionResultMsg string

func (m model) performAction() tea.Cmd {
	return func() tea.Msg {
		if m.selectedAction == 0 {
			// PDF Generieren
			builder := engine.NewBuilder(".", "dist")
			err := builder.Build()
			if err != nil {
				return actionResultMsg(fmt.Sprintf("Fehler: %v", err))
			}
			return actionResultMsg("PDF wurde erfolgreich generiert! (siehe dist/ Ordner)")
		} else {
			// Fonts herunterladen
			if m.cfg == nil || m.cfg.Fonts.URL == "" {
				return actionResultMsg("Keine Font-URL in docgen.yml definiert.")
			}
			// Wir rufen den Builder Build() auf, da dieser den Download triggert
			// Oder wir rufen direkt die Download-Funktion auf, wenn wir nur downloaden wollen.
			// Da Build() alles macht, ist das der sicherste Weg.
			builder := engine.NewBuilder(".", "dist")
			err := builder.Build() // Triggert Download falls URL vorhanden
			if err != nil {
				return actionResultMsg(fmt.Sprintf("Download-Fehler: %v", err))
			}
			return actionResultMsg("Fonts wurden erfolgreich verarbeitet!")
		}
	}
}

var (
	activeTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#11111b")).
		Background(lipgloss.Color("#cba6f7")).
		Padding(0, 2).
		Bold(true).
		MarginRight(1)

	inactiveTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cdd6f4")).
		Background(lipgloss.Color("#313244")).
		Padding(0, 2).
		MarginRight(1)

	docStyle = lipgloss.NewStyle().Padding(1, 4, 1, 4)

	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cba6f7")).
		Bold(true).
		MarginBottom(1)

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89b4fa")).
		Italic(true)

	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#11111b")).
		Background(lipgloss.Color("#cba6f7")).
		Padding(0, 1).
		Bold(true)
)

func (m model) View() string {
	var s strings.Builder

	// Banner
	banner := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cba6f7")).
		Bold(true).
		Render("🚀 goDocGen Professional PDF Builder")
	s.WriteString(banner + "\n\n")

	// Header / Tabs
	tabs := []string{"📖 Hilfe", "⚙️ Konfiguration", "⚡ Aktionen", "🏗️ Init"}
	var renderedTabs []string
	for i, t := range tabs {
		if int(m.state) == i {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, inactiveTabStyle.Render(t))
		}
	}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...) + "\n\n")

	// Content Box
	var content string
	switch m.state {
	case stateHelp:
		content = m.helpView()
	case stateConfig:
		content = m.configView()
	case stateActions:
		content = m.actionsView()
	case stateInit:
		content = m.initModel.View()
	}

	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#45475a")).
		Padding(1, 2).
		Width(80)

	s.WriteString(contentStyle.Render(content))

	if m.statusMsg != "" {
		s.WriteString("\n\n" + infoStyle.Render("ℹ️ "+m.statusMsg))
	}

	// Footer
	s.WriteString("\n\n" + m.help.View(m.keys))

	return docStyle.Render(s.String())
}

func (m model) helpView() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("📘 goDocGen - Hilfe & Funktionen"))
	s.WriteString("\n")
	s.WriteString("Dieses Tool generiert professionelle PDFs aus Markdown-Dateien.\n\n")

	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Render

	s.WriteString(accent("• ") + "Markdown zu PDF: Konvertiert komplexe MD-Strukturen inkl. Tabellen & Listen.\n")
	s.WriteString(accent("• ") + "Mermaid Diagramme: Automatische Einbindung von Diagrammen (Flowcharts, etc.).\n")
	s.WriteString(accent("• ") + "Syntax Highlighting: Schöner Code dank Chroma (Catppuccin Support).\n")
	s.WriteString(accent("• ") + "Custom Fonts: Unterstützung für eigene Schriftarten (lokal .zip oder URL).\n")
	s.WriteString(accent("• ") + "Themes: Vordefinierte Themes (Catppuccin Latte/Mocha) oder eigene Farben.\n")
	s.WriteString(accent("• ") + "TOC: Automatisches Inhaltsverzeichnis basierend auf Überschriften.\n")
	s.WriteString("\n")
	s.WriteString(accent("Nutzung:") + "\n")
	s.WriteString("1. Konfiguration in 'docgen.yml' anpassen.\n")
	s.WriteString("2. Markdown-Inhalte in 'content/' ablegen.\n")
	s.WriteString("3. Assets (Bilder) in 'assets/' ablegen.\n")
	s.WriteString("4. Generierung via 'Aktionen' Tab starten.\n")
	return s.String()
}

func (m model) configView() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("⚙️ Aktuelle Konfiguration (docgen.yml)"))
	s.WriteString("\n")
	if m.cfg == nil {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Render("❌ Keine Konfigurationsdatei gefunden oder Fehler beim Laden."))
	} else {
		keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94e2d5")).Width(18).Render
		valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")).Render

		s.WriteString(keyStyle("📝 Titel:") + valStyle(m.cfg.Title) + "\n")
		s.WriteString(keyStyle("🎨 Theme:") + valStyle(m.cfg.Theme) + "\n")
		s.WriteString(keyStyle("💻 Code-Theme:") + valStyle(m.cfg.CodeTheme) + "\n")
		s.WriteString(keyStyle("🔤 Font (Reg):") + valStyle(m.cfg.Fonts.Regular) + "\n")
		if m.cfg.Fonts.URL != "" {
			s.WriteString(keyStyle("🌐 Font URL:") + valStyle(m.cfg.Fonts.URL) + "\n")
		} else {
			s.WriteString(keyStyle("📦 Font Zip:") + valStyle(m.cfg.Fonts.Zip) + "\n")
		}
		s.WriteString(keyStyle("📏 Schriftgröße:") + valStyle(fmt.Sprintf("%.1f", m.cfg.FontSize)) + "\n")
	}
	s.WriteString("\n" + infoStyle.Render("💡 Hinweis: Bearbeiten Sie die docgen.yml direkt für dauerhafte Änderungen."))
	s.WriteString("\n" + infoStyle.Render("⌨️ Drücken Sie ENTER, um zwischen Latte/Mocha zu wechseln."))
	return s.String()
}

func (m model) actionsView() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("⚡ Verfügbare Aktionen"))
	s.WriteString("\n")

	actions := []string{"📄 PDF generieren", "📥 Fonts herunterladen/aktualisieren"}
	for i, a := range actions {
		prefix := "  "
		style := lipgloss.NewStyle().PaddingLeft(2)
		if m.selectedAction == i {
			prefix = "➜ "
			style = selectedStyle
		}
		s.WriteString(style.Render(prefix+a) + "\n")
	}

	return s.String()
}
