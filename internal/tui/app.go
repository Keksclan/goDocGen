// Package tui implementiert die interaktive Benutzeroberfläche für godocgen.
package tui

import (
	"godocgen/internal/config"
	"godocgen/internal/engine"
	"godocgen/internal/util"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// state definiert die verschiedenen Ansichten der TUI.
type state int

const (
	stateHelp state = iota
	stateConfig
	stateThemes // Neu: Theme-Katalog
	stateActions
	stateProjects // Neu: Projektauswahl aus Historie
	stateInit
)

// model speichert den Zustand der TUI-Anwendung.
type model struct {
	state           state
	width           int
	height          int
	keys            keyMap
	help            help.Model
	cfg             *config.Config
	history         *config.GlobalConfig
	projectPath     string
	err             error
	selectedAction  int
	selectedProject int
	selectedTheme   int
	customThemes    []string
	statusMsg       string
	initModel       InitModel
	viewport        viewport.Model
	// Formular-Felder für Konfiguration
	inputs             []textinput.Model
	focusedInput       int
	configEditMode     bool
	configSectionsOpen []bool
	focusedSection     int
	logs               []string
	lastGeneratedPath  string
}

const totalInputs = 36

// keyMap definiert die Tastenkombinationen der TUI.
type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Help       key.Binding
	Quit       key.Binding
	Action     key.Binding
	GlobalNext key.Binding
	GlobalPrev key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.GlobalNext, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.GlobalPrev, k.GlobalNext, k.Help, k.Quit, k.Action},
	}
}

// keys enthält die Standard-Tastenbelegung.
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
	GlobalNext: key.NewBinding(
		key.WithKeys("ctrl+right"),
		key.WithHelp("ctrl+→", "nächster Tab (global)"),
	),
	GlobalPrev: key.NewBinding(
		key.WithKeys("ctrl+left"),
		key.WithHelp("ctrl+←", "vorheriger Tab (global)"),
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
		key.WithHelp("enter", "ausführen/wählen"),
	),
}

// InitialModel erstellt das Startmodell für die TUI.
func InitialModel() *model {
	h, _ := config.LoadGlobalConfig()
	cfg, _ := config.LoadConfig("docgen.yml")
	projectPath, _ := filepath.Abs(".")

	initialState := stateHelp
	if cfg == nil {
		if len(h.RecentProjects) > 0 {
			initialState = stateProjects
		} else {
			initialState = stateInit
		}
	}

	m := &model{
		state:              initialState,
		keys:               keys,
		help:               help.New(),
		cfg:                cfg,
		history:            h,
		projectPath:        projectPath,
		initModel:          InitialInitModel("."),
		customThemes:       []string{},
		viewport:           viewport.New(100, 20),
		configSectionsOpen: []bool{true, true, true, true, true, true, true},
		logs:               []string{"Willkommen bei goDocGen!"},
	}

	// Suche nach Custom Themes
	if files, err := os.ReadDir(filepath.Join(projectPath, "themes")); err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
				m.customThemes = append(m.customThemes, f.Name())
			}
		}
	} else {
		// Versuche globales themes verzeichnis
		if configDir, err := os.UserConfigDir(); err == nil {
			globalThemes := filepath.Join(configDir, "godocgen", "themes")
			if files, err := os.ReadDir(globalThemes); err == nil {
				for _, f := range files {
					if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
						m.customThemes = append(m.customThemes, f.Name())
					}
				}
			}
		}
	}

	m.setupInputs()
	return m
}

func (m *model) setupInputs() {
	if m.cfg == nil {
		return
	}

	m.inputs = make([]textinput.Model, totalInputs)
	var t textinput.Model

	// 0-2: Basis
	t = textinput.New()
	t.Placeholder = "Titel"
	t.SetValue(m.cfg.Title)
	m.inputs[0] = t

	t = textinput.New()
	t.Placeholder = "Untertitel"
	t.SetValue(m.cfg.Subtitle)
	m.inputs[1] = t

	t = textinput.New()
	t.Placeholder = "Autor"
	t.SetValue(m.cfg.Author)
	m.inputs[2] = t

	// 3-7: Layout & Font
	t = textinput.New()
	t.Placeholder = "Schriftgröße"
	t.SetValue(fmt.Sprintf("%.1f", m.cfg.FontSize))
	m.inputs[3] = t

	t = textinput.New()
	t.Placeholder = "Rand Links"
	t.SetValue(fmt.Sprintf("%.1f", m.cfg.Layout.Margins.Left))
	m.inputs[4] = t

	t = textinput.New()
	t.Placeholder = "Rand Rechts"
	t.SetValue(fmt.Sprintf("%.1f", m.cfg.Layout.Margins.Right))
	m.inputs[5] = t

	t = textinput.New()
	t.Placeholder = "Rand Oben"
	t.SetValue(fmt.Sprintf("%.1f", m.cfg.Layout.Margins.Top))
	m.inputs[6] = t

	t = textinput.New()
	t.Placeholder = "Rand Unten"
	t.SetValue(fmt.Sprintf("%.1f", m.cfg.Layout.Margins.Bottom))
	m.inputs[7] = t

	// 8-9: Themes
	t = textinput.New()
	t.Placeholder = "Theme"
	t.SetValue(m.cfg.Theme)
	m.inputs[8] = t

	t = textinput.New()
	t.Placeholder = "Code-Theme"
	t.SetValue(m.cfg.CodeTheme)
	m.inputs[9] = t

	// 10-11: Header
	t = textinput.New()
	t.Placeholder = "Header Text"
	t.SetValue(m.cfg.Header.Text)
	m.inputs[10] = t

	t = textinput.New()
	t.Placeholder = "Header Bild (Pfad)"
	t.SetValue(m.cfg.Header.Image)
	m.inputs[11] = t

	// 12-13: Footer
	t = textinput.New()
	t.Placeholder = "Footer Text"
	t.SetValue(m.cfg.Footer.Text)
	m.inputs[12] = t

	t = textinput.New()
	t.Placeholder = "Footer Bild (Pfad)"
	t.SetValue(m.cfg.Footer.Image)
	m.inputs[13] = t

	// 14: Seitenzahlen
	t = textinput.New()
	t.Placeholder = "Startseite Seitenzahlen"
	t.SetValue(fmt.Sprintf("%d", m.cfg.PageNumbers.StartPage))
	m.inputs[14] = t

	// 15-16: Ausrichtung
	t = textinput.New()
	t.Placeholder = "Ausrichtung Deckblatt (left/center/right/justify)"
	t.SetValue(m.cfg.Layout.StartPage)
	m.inputs[15] = t

	t = textinput.New()
	t.Placeholder = "Ausrichtung Body (left/center/right/justify)"
	t.SetValue(m.cfg.Layout.Body)
	m.inputs[16] = t

	// 17-21: Gradient
	t = textinput.New()
	t.Placeholder = "Verlauf Aktiviert (true/false)"
	t.SetValue(fmt.Sprintf("%t", m.cfg.Gradient.Enabled))
	m.inputs[17] = t

	t = textinput.New()
	t.Placeholder = "Verlauf Startfarbe"
	t.SetValue(m.cfg.Gradient.Start)
	m.inputs[18] = t

	t = textinput.New()
	t.Placeholder = "Verlauf Endfarbe"
	t.SetValue(m.cfg.Gradient.End)
	m.inputs[19] = t

	t = textinput.New()
	t.Placeholder = "Verlauf Orientierung (vertical/horizontal)"
	t.SetValue(m.cfg.Gradient.Orientation)
	m.inputs[20] = t

	t = textinput.New()
	t.Placeholder = "Verlauf Global (true/false)"
	t.SetValue(fmt.Sprintf("%t", m.cfg.Gradient.Global))
	m.inputs[21] = t

	// 22-27: Fonts
	t = textinput.New()
	t.Placeholder = "Font ZIP Pfad"
	t.SetValue(m.cfg.Fonts.Zip)
	m.inputs[22] = t

	t = textinput.New()
	t.Placeholder = "Font Download URL"
	t.SetValue(m.cfg.Fonts.URL)
	m.inputs[23] = t

	t = textinput.New()
	t.Placeholder = "Font Regular (Dateiname)"
	t.SetValue(m.cfg.Fonts.Regular)
	m.inputs[24] = t

	t = textinput.New()
	t.Placeholder = "Font Bold (Dateiname)"
	t.SetValue(m.cfg.Fonts.Bold)
	m.inputs[25] = t

	t = textinput.New()
	t.Placeholder = "Font Italic (Dateiname)"
	t.SetValue(m.cfg.Fonts.Italic)
	m.inputs[26] = t

	t = textinput.New()
	t.Placeholder = "Font Mono (Dateiname)"
	t.SetValue(m.cfg.Fonts.Mono)
	m.inputs[27] = t

	// 28: Mermaid
	t = textinput.New()
	t.Placeholder = "Mermaid Renderer (mmdc oder leer)"
	t.SetValue(m.cfg.Mermaid.Renderer)
	m.inputs[28] = t

	// 29-31: TOC
	t = textinput.New()
	t.Placeholder = m.T(func(t translation) string { return t.tocEnabled }) + " (true/false)"
	t.SetValue(fmt.Sprintf("%t", m.cfg.TOC.Enabled))
	m.inputs[29] = t

	t = textinput.New()
	t.Placeholder = m.T(func(t translation) string { return t.tocNumbers }) + " (true/false)"
	t.SetValue(fmt.Sprintf("%t", m.cfg.TOC.ShowNumbers))
	m.inputs[30] = t

	t = textinput.New()
	t.Placeholder = m.T(func(t translation) string { return t.tocDots }) + " (true/false)"
	t.SetValue(fmt.Sprintf("%t", m.cfg.TOC.ShowDots))
	m.inputs[31] = t

	// 32: Header Numbering
	t = textinput.New()
	t.Placeholder = m.T(func(t translation) string { return t.headerNumbering }) + " (true/false)"
	t.SetValue(fmt.Sprintf("%t", m.cfg.Layout.HeaderNumbering))
	m.inputs[32] = t

	// 33-35: Footer Designer
	t = textinput.New()
	t.Placeholder = m.T(func(t translation) string { return t.footerLeft })
	t.SetValue(m.cfg.Footer.Left)
	m.inputs[33] = t

	t = textinput.New()
	t.Placeholder = m.T(func(t translation) string { return t.footerCenter })
	t.SetValue(m.cfg.Footer.Center)
	m.inputs[34] = t

	t = textinput.New()
	t.Placeholder = m.T(func(t translation) string { return t.footerRight })
	t.SetValue(m.cfg.Footer.Right)
	m.inputs[35] = t
}

func (m *model) saveConfig() {
	if m.cfg == nil {
		return
	}

	m.cfg.Title = m.inputs[0].Value()
	m.cfg.Subtitle = m.inputs[1].Value()
	m.cfg.Author = m.inputs[2].Value()

	if v, err := strconv.ParseFloat(m.inputs[3].Value(), 64); err == nil {
		m.cfg.FontSize = v
	}
	if v, err := strconv.ParseFloat(m.inputs[4].Value(), 64); err == nil {
		m.cfg.Layout.Margins.Left = v
	}
	if v, err := strconv.ParseFloat(m.inputs[5].Value(), 64); err == nil {
		m.cfg.Layout.Margins.Right = v
	}
	if v, err := strconv.ParseFloat(m.inputs[6].Value(), 64); err == nil {
		m.cfg.Layout.Margins.Top = v
	}
	if v, err := strconv.ParseFloat(m.inputs[7].Value(), 64); err == nil {
		m.cfg.Layout.Margins.Bottom = v
	}

	oldTheme := m.cfg.Theme
	m.cfg.Theme = m.inputs[8].Value()
	m.cfg.CodeTheme = m.inputs[9].Value()

	if m.cfg.Theme != oldTheme {
		m.cfg.Colors = config.Colors{}
		config.ApplyTheme(m.cfg)
	}

	// Neue Felder speichern
	m.cfg.Header.Text = m.inputs[10].Value()
	m.cfg.Header.Image = m.inputs[11].Value()
	m.cfg.Footer.Text = m.inputs[12].Value()
	m.cfg.Footer.Image = m.inputs[13].Value()

	if v, err := strconv.Atoi(m.inputs[14].Value()); err == nil {
		m.cfg.PageNumbers.StartPage = v
	}

	m.cfg.Layout.StartPage = m.inputs[15].Value()
	m.cfg.Layout.Body = m.inputs[16].Value()

	if v, err := strconv.ParseBool(m.inputs[17].Value()); err == nil {
		m.cfg.Gradient.Enabled = v
	}
	m.cfg.Gradient.Start = m.inputs[18].Value()
	m.cfg.Gradient.End = m.inputs[19].Value()
	m.cfg.Gradient.Orientation = m.inputs[20].Value()
	if v, err := strconv.ParseBool(m.inputs[21].Value()); err == nil {
		m.cfg.Gradient.Global = v
	}

	// Fonts & Mermaid speichern
	m.cfg.Fonts.Zip = m.inputs[22].Value()
	m.cfg.Fonts.URL = m.inputs[23].Value()
	m.cfg.Fonts.Regular = m.inputs[24].Value()
	m.cfg.Fonts.Bold = m.inputs[25].Value()
	m.cfg.Fonts.Italic = m.inputs[26].Value()
	m.cfg.Fonts.Mono = m.inputs[27].Value()
	m.cfg.Mermaid.Renderer = m.inputs[28].Value()

	// TOC & Header Numbering speichern
	if v, err := strconv.ParseBool(m.inputs[29].Value()); err == nil {
		m.cfg.TOC.Enabled = v
	}
	if v, err := strconv.ParseBool(m.inputs[30].Value()); err == nil {
		m.cfg.TOC.ShowNumbers = v
	}
	if v, err := strconv.ParseBool(m.inputs[31].Value()); err == nil {
		m.cfg.TOC.ShowDots = v
	}
	if v, err := strconv.ParseBool(m.inputs[32].Value()); err == nil {
		m.cfg.Layout.HeaderNumbering = v
	}

	m.cfg.Footer.Left = m.inputs[33].Value()
	m.cfg.Footer.Center = m.inputs[34].Value()
	m.cfg.Footer.Right = m.inputs[35].Value()

	err := m.cfg.Save(filepath.Join(m.projectPath, "docgen.yml"))
	if err != nil {
		m.statusMsg = fmt.Sprintf("%s: %v", m.T(func(t translation) string { return t.statusError }), err)
		m.log("Fehler beim Speichern: " + err.Error())
	} else {
		m.statusMsg = m.T(func(t translation) string { return t.statusSaved })
		m.log("Konfiguration erfolgreich gespeichert.")
	}
}

func (m *model) log(msg string) {
	m.logs = append(m.logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	if len(m.logs) > 30 {
		m.logs = m.logs[1:]
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

// Update verarbeitet Nachrichten und aktualisiert den Zustand des Modells.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.state == stateInit {
		newInitModel, initCmd := m.initModel.Update(msg)
		m.initModel = newInitModel.(InitModel)
		// Wenn Initialisierung abgeschlossen ist, laden wir die Config neu
		if strings.Contains(m.statusMsg, "erfolgreich") || strings.Contains(m.initModel.View(), "fertig") {
			// Kleiner Hack: Wenn initModel fertig ist, versuchen wir docgen.yml zu laden
			if cfg, err := config.LoadConfig(filepath.Join(m.initModel.Path, "docgen.yml")); err == nil {
				m.cfg = cfg
				m.projectPath, _ = filepath.Abs(m.initModel.Path)
				m.setupInputs()
				m.state = stateHelp
				m.history.AddProject(m.projectPath)
				_ = m.history.Save()
			}
		}
		if initCmd != nil {
			return m, initCmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Berechne Viewport-Höhe: Gesamt - Banner(1) - Pfad(2) - Tabs(2) - Status(2) - Footer(2) - Borders(2)
		vHeight := msg.Height - 12
		if vHeight < 5 {
			vHeight = 5
		}

		consoleWidth := 40
		if msg.Width < 100 {
			consoleWidth = 30
		}

		m.viewport.Width = msg.Width - consoleWidth - 12
		if m.viewport.Width > 120 {
			m.viewport.Width = 120
		}
		m.viewport.Height = vHeight
	case tea.KeyMsg:
		// Globale Tab-Navigation (immer verfügbar, auch im Edit-Mode)
		if msg.Type == tea.KeyCtrlRight {
			m.state = (m.state + 1) % 6
			m.statusMsg = ""
			m.configEditMode = false
			return m, nil
		}
		if msg.Type == tea.KeyCtrlLeft {
			m.state = (m.state + 5) % 6
			m.statusMsg = ""
			m.configEditMode = false
			return m, nil
		}

		// Spezialfall für Sprache wechseln (z.B. mit 'L')
		if msg.String() == "L" {
			if m.history.Language == "de" {
				m.history.Language = "en"
			} else {
				m.history.Language = "de"
			}
			_ = m.history.Save()
			m.statusMsg = m.T(func(t translation) string { return t.langSwitch })
			return m, nil
		}

		if m.state == stateConfig && m.cfg != nil {
			if msg.Type == tea.KeyCtrlC {
				return m, tea.Quit
			}

			if !m.configEditMode {
				// Lese-Modus
				if msg.String() == "j" || msg.Type == tea.KeyDown {
					m.focusedSection = (m.focusedSection + 1) % len(m.configSectionsOpen)
					return m, nil
				}
				if msg.String() == "k" || msg.Type == tea.KeyUp {
					m.focusedSection = (m.focusedSection + len(m.configSectionsOpen) - 1) % len(m.configSectionsOpen)
					return m, nil
				}
				if msg.Type == tea.KeyEnter || msg.String() == " " {
					m.configSectionsOpen[m.focusedSection] = !m.configSectionsOpen[m.focusedSection]
					return m, nil
				}
				if msg.String() == "e" || msg.String() == "E" {
					m.configEditMode = true
					m.statusMsg = ""
					// Öffne alle Sektionen beim Bearbeiten für bessere Übersicht
					for i := range m.configSectionsOpen {
						m.configSectionsOpen[i] = true
					}
					for i := range m.inputs {
						if i == m.focusedInput {
							m.inputs[i].Focus()
						} else {
							m.inputs[i].Blur()
						}
					}
					return m, nil
				}
				if msg.Type == tea.KeyEsc {
					m.state = stateHelp
					m.statusMsg = ""
					return m, nil
				}
			} else {
				// Bearbeitungs-Modus
				// Esc führt zurück zum Lese-Modus
				if msg.Type == tea.KeyEsc {
					m.configEditMode = false
					m.statusMsg = ""
					for i := range m.inputs {
						m.inputs[i].Blur()
					}
					return m, nil
				}

				// Fokus-Navigation im Formular
				if msg.Type == tea.KeyTab || msg.Type == tea.KeyDown {
					m.focusedInput = (m.focusedInput + 1) % len(m.inputs)
					for i := range m.inputs {
						if i == m.focusedInput {
							m.inputs[i].Focus()
						} else {
							m.inputs[i].Blur()
						}
					}
					// Auto-Scroll
					m.viewport.SetYOffset(m.focusedInput * 2)
				} else if msg.Type == tea.KeyShiftTab || msg.Type == tea.KeyUp {
					m.focusedInput = (m.focusedInput + len(m.inputs) - 1) % len(m.inputs)
					for i := range m.inputs {
						if i == m.focusedInput {
							m.inputs[i].Focus()
						} else {
							m.inputs[i].Blur()
						}
					}
					m.viewport.SetYOffset(m.focusedInput * 2)
				} else if msg.Type == tea.KeyEnter {
					// Speichern
					m.saveConfig()
				} else if (msg.Type == tea.KeyRight || msg.Type == tea.KeyLeft) && m.focusedInput == 8 {
					// Theme-Switcher im Formular
					themes := config.GetAvailableThemes()
					current := m.inputs[8].Value()
					idx := -1
					for i, t := range themes {
						if t == current {
							idx = i
							break
						}
					}
					if idx == -1 {
						idx = 0
					}
					if msg.Type == tea.KeyRight {
						idx = (idx + 1) % len(themes)
					} else {
						idx = (idx + len(themes) - 1) % len(themes)
					}
					m.inputs[8].SetValue(themes[idx])
					return m, nil
				} else if (msg.Type == tea.KeyRight || msg.Type == tea.KeyLeft) && m.focusedInput == 9 {
					// Code-Theme Switcher
					themes := config.GetAvailableThemes()
					current := m.inputs[9].Value()
					idx := -1
					for i, t := range themes {
						if t == current {
							idx = i
							break
						}
					}
					if idx == -1 {
						idx = 0
					}
					if msg.Type == tea.KeyRight {
						idx = (idx + 1) % len(themes)
					} else {
						idx = (idx + len(themes) - 1) % len(themes)
					}
					m.inputs[9].SetValue(themes[idx])
					return m, nil
				}

				// Aktiven Input aktualisieren
				var cmd tea.Cmd
				m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Right):
			m.state = (m.state + 1) % 6
			m.statusMsg = ""
		case key.Matches(msg, m.keys.Left):
			m.state = (m.state + 5) % 6
			m.statusMsg = ""
		case key.Matches(msg, m.keys.Up):
			if m.state == stateActions {
				m.selectedAction = (m.selectedAction + 3) % 4
			} else if m.state == stateProjects && len(m.history.RecentProjects) > 0 {
				m.selectedProject = (m.selectedProject + len(m.history.RecentProjects) - 1) % len(m.history.RecentProjects)
			} else if m.state == stateThemes {
				themes := config.GetAvailableThemes()
				total := len(themes) + len(m.customThemes)
				if total > 0 {
					m.selectedTheme = (m.selectedTheme + total - 1) % total
				}
			}
		case key.Matches(msg, m.keys.Down):
			if m.state == stateActions {
				m.selectedAction = (m.selectedAction + 1) % 4
			} else if m.state == stateProjects && len(m.history.RecentProjects) > 0 {
				m.selectedProject = (m.selectedProject + 1) % len(m.history.RecentProjects)
			} else if m.state == stateThemes {
				themes := config.GetAvailableThemes()
				total := len(themes) + len(m.customThemes)
				if total > 0 {
					m.selectedTheme = (m.selectedTheme + 1) % total
				}
			}
		case key.Matches(msg, m.keys.Action):
			if m.state == stateActions {
				return m, m.performAction()
			}
			if m.state == stateThemes {
				themes := config.GetAvailableThemes()
				if m.selectedTheme < len(themes) {
					theme := themes[m.selectedTheme]
					m.cfg.Theme = theme
					m.inputs[8].SetValue(theme)
					m.cfg.Colors = config.Colors{}
					config.ApplyTheme(m.cfg)
					_ = m.cfg.Save(filepath.Join(m.projectPath, "docgen.yml"))
					m.statusMsg = "Theme angewendet: " + theme
				} else if m.selectedTheme < len(themes)+len(m.customThemes) {
					theme := m.customThemes[m.selectedTheme-len(themes)]
					m.cfg.Theme = theme
					m.inputs[8].SetValue(theme)
					_ = m.cfg.Save(filepath.Join(m.projectPath, "docgen.yml"))
					m.statusMsg = "Custom Theme angewendet: " + theme
				}
				return m, nil
			}
			if m.state == stateProjects && len(m.history.RecentProjects) > 0 {
				path := m.history.RecentProjects[m.selectedProject]
				cfg, err := config.LoadConfig(filepath.Join(path, "docgen.yml"))
				if err != nil {
					m.statusMsg = fmt.Sprintf("Fehler beim Laden: %v", err)
				} else {
					m.cfg = cfg
					m.projectPath = path
					m.setupInputs()
					m.state = stateHelp
					m.history.AddProject(path)
					_ = m.history.Save()
					m.statusMsg = "Projekt gewechselt zu: " + filepath.Base(path)
				}
				return m, nil
			}
		}
	case actionResultMsg:
		m.statusMsg = string(msg)
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	return m, vpCmd
}

type actionResultMsg string

// performAction führt die gewählte Aktion (Build oder Font-Download) aus.
func (m *model) performAction() tea.Cmd {
	return func() tea.Msg {
		if m.cfg == nil {
			return actionResultMsg("Kein Projekt geladen.")
		}
		switch m.selectedAction {
		case 0: // PDF Generieren
			m.log("Starte PDF-Generierung...")
			builder := engine.NewBuilder(m.projectPath, filepath.Join(m.projectPath, "dist"))
			path, err := builder.Build()
			if err != nil {
				m.log("Build-Fehler: " + err.Error())
				return actionResultMsg(fmt.Sprintf("Fehler: %v", err))
			}
			m.lastGeneratedPath = path
			m.log("PDF erfolgreich generiert: " + filepath.Base(path))
			m.log("Pfad: " + path)
			return actionResultMsg("PDF wurde erfolgreich generiert!")
		case 1: // PDF Öffnen
			if m.lastGeneratedPath == "" {
				return actionResultMsg("Noch kein PDF in dieser Sitzung generiert.")
			}
			m.log("Öffne PDF: " + m.lastGeneratedPath)
			err := util.OpenPath(m.lastGeneratedPath)
			if err != nil {
				return actionResultMsg("Fehler beim Öffnen: " + err.Error())
			}
			return actionResultMsg("PDF geöffnet.")
		case 2: // Fonts verarbeiten
			m.log("Verarbeite Schriften...")
			if m.cfg.Fonts.URL == "" {
				m.log("Abbruch: Keine Font-URL definiert.")
				return actionResultMsg("Keine Font-URL in docgen.yml definiert.")
			}
			builder := engine.NewBuilder(m.projectPath, filepath.Join(m.projectPath, "dist"))
			_, err := builder.Build()
			if err != nil {
				m.log("Verarbeitungs-Fehler: " + err.Error())
				return actionResultMsg(fmt.Sprintf("Verarbeitungs-Fehler: %v", err))
			}
			m.log("Schriften erfolgreich verarbeitet.")
			return actionResultMsg("Fonts wurden erfolgreich verarbeitet!")
		case 3: // Zum PATH hinzufügen
			m.log("Füge Programm zum PATH hinzu...")
			err := util.AddToPath()
			if err != nil {
				m.log("PATH-Fehler: " + err.Error())
				return actionResultMsg("Fehler: " + err.Error())
			}
			m.log("Erfolgreich zum PATH hinzugefügt.")
			return actionResultMsg(m.T(func(t translation) string { return t.statusPathAdded }))
		default:
			return actionResultMsg("Unbekannte Aktion.")
		}
	}
}

// Styling-Definitionen
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

	consoleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cdd6f4")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#45475a")).
		Padding(1, 1)
)

func (m *model) renderConsole() string {
	var s strings.Builder
	title := lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true).Render("💻 Console")
	s.WriteString(title + "\n\n")

	// Nur die letzten N Zeilen anzeigen, die in den Viewport passen
	maxLogs := m.viewport.Height
	start := len(m.logs) - maxLogs
	if start < 0 {
		start = 0
	}

	for i := start; i < len(m.logs); i++ {
		log := m.logs[i]
		if len(log) > 35 {
			log = log[:32] + "..."
		}
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#bac2de")).Render(log) + "\n")
	}
	return s.String()
}

// View rendert die Benutzeroberfläche als String.
func (m *model) View() string {
	var s strings.Builder

	// Banner
	banner := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cba6f7")).
		Bold(true).
		Render("🚀 goDocGen Professional PDF Builder")
	s.WriteString(banner + "\n")

	if m.cfg != nil {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Render(m.T(func(t translation) string { return t.project })+": "+m.projectPath) + "\n\n")
	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Render(m.T(func(t translation) string { return t.noProject })) + "\n\n")
	}

	// Header / Tabs
	tabs := []string{
		m.T(func(t translation) string { return t.helpTab }),
		m.T(func(t translation) string { return t.configTab }),
		m.T(func(t translation) string { return t.themeTab }),
		m.T(func(t translation) string { return t.actionsTab }),
		m.T(func(t translation) string { return t.projectsTab }),
		m.T(func(t translation) string { return t.initTab }),
	}
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
	case stateThemes:
		content = m.themesView()
	case stateActions:
		content = m.actionsView()
	case stateProjects:
		content = m.projectsView()
	case stateInit:
		content = m.initModel.View()
	}

	m.viewport.SetContent(content)

	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#45475a")).
		Padding(1, 2).
		Width(m.viewport.Width + 4).
		Height(m.viewport.Height + 2)

	mainView := contentStyle.Render(m.viewport.View())

	// Console View
	consoleWidth := 40
	if m.width < 100 {
		consoleWidth = 30
	}
	cStyle := consoleStyle.Copy().
		Width(consoleWidth).
		Height(m.viewport.Height + 2)

	consoleView := cStyle.Render(m.renderConsole())

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, mainView, consoleView))

	if m.statusMsg != "" {
		s.WriteString("\n\n" + infoStyle.Render("ℹ️ "+m.statusMsg))
	}

	// Footer
	s.WriteString("\n\n" + m.help.View(m.keys))

	return docStyle.Render(s.String())
}

func (m *model) helpView() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render(m.T(func(t translation) string { return t.helpHeader })))
	s.WriteString("\n")
	if m.history.Language == "de" {
		s.WriteString("Dieses Tool generiert professionelle PDFs aus Markdown-Dateien.\n\n")
	} else {
		s.WriteString("This tool generates professional PDFs from Markdown files.\n\n")
	}

	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Render

	if m.history.Language == "de" {
		s.WriteString(accent("• ") + "Markdown zu PDF: Konvertiert komplexe MD-Strukturen inkl. Tabellen & Listen.\n")
		s.WriteString(accent("• ") + "Mermaid Diagramme: Automatische Einbindung von Diagrammen (Flowcharts, etc.).\n")
		s.WriteString(accent("• ") + "Syntax Highlighting: Schöner Code dank Chroma (Catppuccin Support).\n")
		s.WriteString(accent("• ") + "Custom Fonts: Unterstützung für eigene Schriftarten (lokal .zip oder URL).\n")
		s.WriteString(accent("• ") + "Themes: Vordefinierte Themes (Catppuccin Latte/Mocha) oder eigene Farben.\n")
		s.WriteString(accent("• ") + "TOC: Automatisches Inhaltsverzeichnis basierend auf Überschriften.\n")
		s.WriteString("\n")
		s.WriteString(accent("Nutzung:") + "\n")
		s.WriteString("1. Konfiguration in 'docgen.yml' anpassen (oder via TUI).\n")
		s.WriteString("2. Markdown-Inhalte in 'content/' ablegen.\n")
		s.WriteString("3. Assets (Bilder) in 'assets/' ablegen.\n")
		s.WriteString("4. Generierung via 'Aktionen' Tab starten.\n")
	} else {
		s.WriteString(accent("• ") + "Markdown to PDF: Convert complex MD structures including tables & lists.\n")
		s.WriteString(accent("• ") + "Mermaid Diagrams: Automatic integration of diagrams (flowcharts, etc.).\n")
		s.WriteString(accent("• ") + "Syntax Highlighting: Beautiful code via Chroma (Catppuccin support).\n")
		s.WriteString(accent("• ") + "Custom Fonts: Support for custom fonts (local .zip or URL).\n")
		s.WriteString(accent("• ") + "Themes: Predefined themes (Catppuccin Latte/Mocha) or custom colors.\n")
		s.WriteString(accent("• ") + "TOC: Automatic table of contents based on headings.\n")
		s.WriteString("\n")
		s.WriteString(accent("Usage:") + "\n")
		s.WriteString("1. Adjust configuration in 'docgen.yml' (or via TUI).\n")
		s.WriteString("2. Place Markdown content in 'content/'.\n")
		s.WriteString("3. Place assets (images) in 'assets/'.\n")
		s.WriteString("4. Start generation via 'Actions' tab.\n")
	}
	return s.String()
}

func (m *model) configView() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("⚙️ " + m.T(func(t translation) string { return t.configTab }) + " (docgen.yml)"))
	s.WriteString("\n")

	if m.cfg == nil {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Render("❌ " + m.T(func(t translation) string { return t.noProject })))
		return s.String()
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94e2d5")).Width(20).Render

	renderInput := func(idx int) string {
		if idx >= len(m.inputs) {
			return ""
		}
		label := m.inputs[idx].Placeholder
		if !m.configEditMode {
			val := m.inputs[idx].Value()
			if val == "" {
				val = "-"
			}
			return keyStyle("  "+label+":") + lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")).Render(val)
		}

		if m.focusedInput == idx {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true).Render("➜ "+label+": ") + m.inputs[idx].View()
		}
		return keyStyle("  "+label+":") + m.inputs[idx].View()
	}

	sections := []struct {
		title   string
		indices []int
		color   string
	}{
		{m.T(func(t translation) string { return t.sectionBasis }), []int{0, 1, 2}, "#cba6f7"},
		{m.T(func(t translation) string { return t.sectionLayout }), []int{3, 15, 16, 4, 5, 6, 7, 32}, "#89b4fa"},
		{m.T(func(t translation) string { return t.sectionHeader }), []int{10, 11, 12, 13, 33, 34, 35}, "#f9e2af"},
		{m.T(func(t translation) string { return t.sectionDesign }), []int{14, 8, 9}, "#a6e3a1"},
		{m.T(func(t translation) string { return t.sectionTOC }), []int{29, 30, 31}, "#fab387"},
		{m.T(func(t translation) string { return t.sectionGradient }), []int{17, 18, 19, 20, 21}, "#eba0ac"},
		{m.T(func(t translation) string { return t.sectionFonts }), []int{22, 23, 24, 25, 26, 27, 28}, "#f5c2e7"},
	}

	for i, sec := range sections {
		isOpen := m.configSectionsOpen[i]
		prefix := "▶ "
		if isOpen {
			prefix = "▼ "
		}

		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(sec.color)).Bold(true)
		if !m.configEditMode && m.focusedSection == i {
			headerStyle = headerStyle.Background(lipgloss.Color("#313244"))
			prefix = "➜ "
		}

		s.WriteString("\n" + headerStyle.Render(prefix+sec.title) + "\n")

		if isOpen {
			var rows []string
			for j := 0; j < len(sec.indices); j += 2 {
				leftIdx := sec.indices[j]
				left := renderInput(leftIdx)

				if j+1 < len(sec.indices) {
					rightIdx := sec.indices[j+1]
					right := renderInput(rightIdx)
					rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
						lipgloss.NewStyle().Width(m.viewport.Width/2).Render(left),
						lipgloss.NewStyle().Width(m.viewport.Width/2).Render(right),
					)+"\n")
				} else {
					rows = append(rows, left+"\n")
				}
			}
			s.WriteString(strings.Join(rows, ""))
		}
	}

	if !m.configEditMode {
		s.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Bold(true).Render("▶ "+m.T(func(t translation) string { return t.pressEToEdit })))
		s.WriteString("\n" + infoStyle.Render("💡 ↑/↓: Sektion wählen | ENTER: Auf/Zu | E: Bearbeiten | ESC: Hilfe"))
	} else {
		s.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true).Render("📝 "+m.T(func(t translation) string { return t.editMode })))
		s.WriteString("\n" + infoStyle.Render("💡 ENTER: Speichern | TAB/↑↓: Navigieren | ESC: Zurück | CTRL+←/→: Tab"))
	}
	return s.String()
}

func (m *model) themesView() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("🎨 " + m.T(func(t translation) string { return t.themeTab })))
	s.WriteString("\n\n")

	themes := config.GetAvailableThemes()

	s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true).Render("✨ App Themes") + "\n")
	for i, t := range themes {
		prefix := "  "
		style := lipgloss.NewStyle().PaddingLeft(2)
		if m.selectedTheme == i {
			prefix = "➜ "
			style = selectedStyle
		}

		// Farbvorschau
		themeColors := config.GetThemeColors(t)
		dot := "●"
		preview := " " + lipgloss.NewStyle().Foreground(lipgloss.Color(themeColors.Title)).Render(dot) +
			lipgloss.NewStyle().Foreground(lipgloss.Color(themeColors.Header)).Render(dot) +
			lipgloss.NewStyle().Foreground(lipgloss.Color(themeColors.Accent)).Render(dot)

		s.WriteString(style.Render(prefix+t) + preview + "\n")
	}

	if len(m.customThemes) > 0 {
		s.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true).Render("🛠️ Custom Themes (.json)") + "\n")
		for i, t := range m.customThemes {
			idx := i + len(themes)
			prefix := "  "
			style := lipgloss.NewStyle().PaddingLeft(2)
			if m.selectedTheme == idx {
				prefix = "➜ "
				style = selectedStyle
			}
			s.WriteString(style.Render(prefix+t) + "\n")
		}
	}

	s.WriteString("\n" + infoStyle.Render("⌨️ ENTER: Theme anwenden | ↑↓: Navigieren"))
	return s.String()
}

func (m *model) actionsView() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("⚡ " + m.T(func(t translation) string { return t.actionsTab })))
	s.WriteString("\n")

	actions := []string{
		m.T(func(t translation) string { return t.buildPdf }),
		m.T(func(t translation) string { return t.openPdf }),
		m.T(func(t translation) string { return t.downloadFont }),
		m.T(func(t translation) string { return t.addToPath }),
	}
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

func (m *model) projectsView() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("📂 " + m.T(func(t translation) string { return t.projectsTab })))
	s.WriteString("\n")

	if len(m.history.RecentProjects) == 0 {
		if m.history.Language == "de" {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Render("Keine Projekte in der Historie."))
		} else {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Render("No projects in history."))
		}
		return s.String()
	}

	for i, p := range m.history.RecentProjects {
		prefix := "  "
		style := lipgloss.NewStyle().PaddingLeft(2)
		if m.selectedProject == i {
			prefix = "➜ "
			style = selectedStyle
		}

		displayPath := p
		if len(p) > 60 {
			displayPath = "..." + p[len(p)-57:]
		}

		s.WriteString(style.Render(prefix+displayPath) + "\n")
	}

	if m.history.Language == "de" {
		s.WriteString("\n" + infoStyle.Render("⌨️ ENTER zum Öffnen des gewählten Projekts"))
	} else {
		s.WriteString("\n" + infoStyle.Render("⌨️ ENTER to open the selected project"))
	}
	return s.String()
}
