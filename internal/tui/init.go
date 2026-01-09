package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InitModel steuert den Prozess der Projekt-Initialisierung in der TUI.
type InitModel struct {
	pathInput     textinput.Model
	themeChoices  []string
	cursor        int
	step          int // 0: Pfadeingabe, 1: Theme-Auswahl
	done          bool
	err           error
	selectedTheme string
	Path          string // Öffentlich für Zugriff aus app.go
}

// InitialInitModel erstellt ein neues Modell für die Initialisierung.
func InitialInitModel(defaultPath string) InitModel {
	ti := textinput.New()
	ti.Placeholder = "Projektpfad (z.B. ./mein-projekt)"
	ti.SetValue(defaultPath)
	ti.Focus()

	return InitModel{
		pathInput:    ti,
		themeChoices: []string{"Catppuccin Mocha (Dunkel)", "Catppuccin Latte (Hell)", "Catppuccin Frappe", "Catppuccin Macchiato", "Red White (Modern)", "IHK-Standard (Arial, 11pt, 1.5-zeilig)"},
		step:         0,
		Path:         defaultPath,
	}
}

func (m InitModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update verarbeitet Eingaben während der Initialisierung.
func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			// Esc macht hier nichts oder bricht ab, ohne das Programm zu beenden
			return m, nil

		case tea.KeyEnter:
			if m.step == 0 {
				if m.pathInput.Value() == "" {
					m.pathInput.SetValue(".")
				}
				m.Path = m.pathInput.Value()
				m.step = 1
				return m, nil
			} else {
				m.selectedTheme = m.themeChoices[m.cursor]
				m.done = true
				return m, m.finishInit()
			}

		case tea.KeyUp, tea.KeyLeft:
			if m.step == 1 && m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown, tea.KeyRight:
			if m.step == 1 && m.cursor < len(m.themeChoices)-1 {
				m.cursor++
			}
		}
	}

	if m.step == 0 {
		m.pathInput, cmd = m.pathInput.Update(msg)
		m.Path = m.pathInput.Value()
	}

	return m, cmd
}

// finishInit führt die eigentliche Dateierstellung durch.
func (m InitModel) finishInit() tea.Cmd {
	return func() tea.Msg {
		target := m.Path
		theme := "catppuccin-mocha"
		accent := "#89b4fa"
		titleColor := "#89b4fa"
		fontSize := 12.0
		lineSpacing := 1.0
		marginLeft, marginRight, marginTop, marginBottom := 10.0, 10.0, 10.0, 10.0

		switch m.cursor {
		case 1:
			theme = "catppuccin-latte"
			accent = "#1e66f5"
			titleColor = "#1e66f5"
		case 2:
			theme = "catppuccin-frappe"
			accent = "#8caaee"
			titleColor = "#8caaee"
		case 3:
			theme = "catppuccin-macchiato"
			accent = "#8ad1fa"
			titleColor = "#8aadf4"
		case 4:
			theme = "red-white"
			accent = "#e30613"
			titleColor = "#e30613"
		case 5:
			theme = "github"
			accent = "#24292e"
			titleColor = "#24292e"
			fontSize = 11.0
			lineSpacing = 1.5
			marginLeft = 25.0
			marginRight = 25.0
			marginTop = 20.0
			marginBottom = 20.0
		}

		dirs := []string{"content", "assets", "fonts"}
		for _, d := range dirs {
			if err := os.MkdirAll(filepath.Join(target, d), 0755); err != nil {
				return err
			}
		}

		configContent := fmt.Sprintf(`title: "Projektdokumentation"
subtitle: "Betriebliche Projektarbeit"
author: "Dein Name"
header:
  text: "Abschlussprüfung Sommer 2026"
footer:
  left: "{author}"
  center: "{title}"
  right: "Seite {page} von {total}"
colors:
  title: "%s"
  header: "%s"
  accent: "%s"
fonts:
  zip: "fonts/fonts.zip"
  regular: "Arial.ttf"
  bold: "Arial-Bold.ttf"
  italic: "Arial-Italic.ttf"
  mono: "Courier.ttf"
font_size: %.1f
layout:
  startpage: "center"
  body: "justify"
  line_spacing: %.1f
  margins:
    left: %.1f
    right: %.1f
    top: %.1f
    bottom: %.1f
gradient:
  enabled: false
  start: "#1e1e2e"
  end: "#89b4fa"
  orientation: "vertical"
  global: false
code_theme: "%s"
`, titleColor, titleColor, accent, fontSize, lineSpacing, marginLeft, marginRight, marginTop, marginBottom, theme)

		if err := os.WriteFile(filepath.Join(target, "docgen.yml"), []byte(configContent), 0644); err != nil {
			return err
		}

		sample := "# Willkommen\n\nDies ist dein neues Projekt."
		os.WriteFile(filepath.Join(target, "content", "01_intro.md"), []byte(sample), 0644)

		gitInit := exec.Command("git", "init")
		gitInit.Dir = target
		gitInit.Run()

		return nil
	}
}

// View rendert die Initialisierungs-Ansicht.
func (m InitModel) View() string {
	if m.done {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6e3a1")).
			Background(lipgloss.Color("#313244")).
			Padding(1, 2).
			Bold(true).
			Render(fmt.Sprintf("✓ Projekt in %s erfolgreich initialisiert!\nTheme: %s\nGit wurde ebenfalls initialisiert.", m.Path, m.selectedTheme))
	}

	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true).Underline(true).Render("🚀 goDocGen Projekt Initialisierung"))
	s.WriteString("\n\n")

	if m.step == 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Render("📂 Wo soll das Projekt erstellt werden?\n"))
		s.WriteString(m.pathInput.View())
		s.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Italic(true).Render("(Enter zum Bestätigen)"))
	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Render("🎨 Wähle ein Catppuccin Theme:\n\n"))
		for i, choice := range m.themeChoices {
			cursor := "  "
			style := lipgloss.NewStyle()
			if m.cursor == i {
				cursor = "➜ "
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true).Background(lipgloss.Color("#313244")).PaddingRight(1)
				s.WriteString(style.Render(cursor+choice) + "\n")
			} else {
				s.WriteString(fmt.Sprintf("%s%s\n", cursor, choice))
			}
		}
		s.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Italic(true).Render("(Enter zum Erstellen)"))
	}

	return s.String()
}
