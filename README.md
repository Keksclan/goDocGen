# goDocGen - Professional PDF DocBuilder

goDocGen ist ein CLI-Tool zur Erzeugung professioneller PDF-Dokumentation aus Markdown-Dateien. Es wurde speziell für die Erstellung von technischen Dokumentationen, API-Referenzen und System-Architekturen entwickelt.

## Features

- 🚀 **Markdown Support**: Volle Unterstützung für CommonMark inklusive GFM-Tabellen.
- 📊 **Erweiterte Mermaid Integration**: Flussdiagramme, Sequenzdiagramme, Klassendiagramme und State-Diagramme.
- 🎨 **Corporate Identity**: Volle Kontrolle über Farben, Schriftarten und Layouts (Default: Catppuccin Theme).
- 💻 **Modernes Code Rendering**: Syntax-Highlighting im IDE-Stil mit abgerundeten Containern und Sprach-Indikatoren.
- 🖱️ **Interaktives TUI**: Starten Sie das Interface mit `godocgen tui`. Es merkt sich zuletzt geöffnete Projekte für schnellen Zugriff.
- ⏬ **Font Downloader**: Laden Sie Schriftarten direkt via URL in der Konfiguration.
- 📑 **Interaktive Navigation**: Automatische Inhaltsverzeichnisse mit klickbaren Links zu den Kapiteln.
- 📁 **Hierarchische Struktur**: Ordnerbasierte Organisation des Contents wird automatisch in die Dokumentenstruktur übernommen.
- 📦 **Publishing Ready**: Automatisierte Versionierung der PDFs im `dist` Ordner.

## Installation

### Voraussetzungen
- **Go 1.24+**
- **mermaid-cli (optional)**: `npm install -g @mermaid-js/mermaid-cli` (für schnellere Diagramm-Generierung). Falls nicht vorhanden, nutzt goDocGen automatisch Chrome/Chromium.

### Build
```bash
go build -o godocgen.exe ./cmd/godocgen
```

### Automatische Installation (empfohlen)
Führen Sie das mitgelieferte PowerShell-Skript aus, um `godocgen` automatisch zu Ihrem PATH hinzuzufügen:

```powershell
.\install.ps1
```
*Hinweis: Dies fügt das aktuelle Verzeichnis zu Ihrem Benutzer-PATH hinzu.*

### Manuelle Installation
Damit Sie `godocgen` von überall aus aufrufen können, fügen Sie das Verzeichnis mit der `godocgen.exe` zu Ihrer PATH-Umgebungsvariable hinzu.

#### Windows
1. Drücken Sie `Win + R`, geben Sie `sysdm.cpl` ein und drücken Sie Enter.
2. Gehen Sie auf den Reiter **Erweitert** und klicken Sie auf **Umgebungsvariablen**.
3. Wählen Sie unter "Benutzervariablen" den Eintrag **Path** aus und klicken Sie auf **Bearbeiten**.
4. Klicken Sie auf **Neu** und geben Sie den Pfad zum Ordner an, in dem die `godocgen.exe` gespeichert ist.
5. Bestätigen Sie alles mit OK und starten Sie Ihr Terminal neu.

#### Linux / macOS
Fügen Sie folgende Zeile zu Ihrer `.bashrc` oder `.zshrc` hinzu:
```bash
export PATH=$PATH:/pfad/zu/deinem/ordner
```

## Quick Start

### 1. Projekt initialisieren
Erzeugt eine fertige Struktur mit Beispiel-Content und Konfiguration:
```bash
godocgen init my_docs
```

### 2. PDF generieren
Baut das Projekt und speichert das Ergebnis (mit automatischer Versionierung) im `dist` Ordner:
```bash
godocgen build --project ./my_docs --out ./dist
```

## Publishing & Deployment

goDocGen eignet sich hervorragend für CI/CD Pipelines:
1. **GitHub Actions**: Nutzen Sie ein Go-Environment, um bei jedem Push auf `main` eine neue Dokumenten-Version zu generieren.
2. **Artifact Storage**: Die generierten PDFs in `dist/` können als Build-Artefakte gespeichert oder direkt auf Servern veröffentlicht werden.
3. **Versionierung**: Durch das automatische Anhängen von `_v1`, `_v2` usw. bleiben alte Stände erhalten.

## Projektstruktur

```
my-docs/
├── docgen.yml      # Zentrale Konfiguration (Farben, Fonts, Margins)
├── content/        # Markdown Dateien (verschachtelte Ordner möglich)
├── assets/         # Bilder & Grafiken
└── fonts/          # ZIP mit TTF-Dateien (Arial, Courier, etc.)
```

## Konfiguration (docgen.yml)

Die `docgen.yml` steuert das gesamte Erscheinungsbild Ihres Dokuments. Hier ist eine Übersicht aller verfügbaren Optionen:

### Basis-Informationen
- `title`: Der Haupttitel des Dokuments (erscheint auf dem Deckblatt).
- `subtitle`: Ein Untertitel für das Deckblatt.
- `author`: Name des Autors.

### Layout & Abstände
- `font_size`: Standard-Schriftgröße für den Fließtext (z.B. `12`).
- `layout`:
  - `startpage`: Ausrichtung des Titels (`left`, `center`, `right`, `justify`).
  - `body`: Standard-Textausrichtung (`left`, `center`, `right`, `justify`).
  - `margins`: Seitenränder in mm (`left`, `right`, `top`, `bottom`).
- `page_numbers`:
  - `start_page`: Die physische Seite, ab der die Seitennummerierung im Footer beginnt (z.B. `3`).

### Farben & Design
- `colors`:
  - `title`: Hex-Farbe für Überschriften (z.B. `#1e66f5`).
  - `header`: Hex-Farbe für den Header-Text.
  - `text`: Standard-Textfarbe.
  - `background`: Hintergrundfarbe der Seiten.
  - `accent`: Farbe für Akzent-Elemente (z.B. die Linie vor H1).
- `gradient`:
  - `enabled`: `true` um einen Farbverlauf zu aktivieren.
  - `start` / `end`: Start- und Endfarbe des Verlaufs (Hex).
  - `orientation`: `vertical` oder `horizontal`.
  - `global`: `true` um den Verlauf auf allen Seiten anzuzeigen, `false` nur für das Deckblatt.

### Header & Footer
- `header` / `footer`:
  - `text`: Der anzuzeigende Text (für Footer veraltet, nutze `left`/`center`/`right`).
  - `image`: Pfad zu einem Bild (Logo), das im Header/Footer angezeigt werden soll.
  - `left` / `center` / `right`: Definieren Sie den Inhalt für die drei Zonen im Footer. Unterstützt Platzhalter: `{page}`, `{total}`, `{title}`, `{author}`, `{date}`.

### Schriften (Fonts)
- `fonts`:
  - `zip`: Pfad zu einem ZIP-Archiv, das die `.ttf` Dateien enthält.
  - `url`: Alternativ eine URL zum Download eines Font-Zips.
  - `regular`: Dateiname der normalen Schriftart (muss im ZIP sein).
  - `bold`: Dateiname der fettgedruckten Variante.
  - `italic`: Dateiname der kursiven Variante.
  - `mono`: Dateiname der Monospace-Schriftart für Code.

### Diagramme & Code
- `code_theme`: Name des Chroma-Highlights Themes (z.B. `github`, `monokai`, `catppuccin-mocha`).
- `mermaid`:
  - `renderer`: `mmdc` (nutzt mermaid-cli) oder leer lassen für den automatischen Chrome-Fallback.

### Beispiel Konfiguration

```yaml
title: "System-Architektur 2026"
subtitle: "Interne Dokumentation v2"
author: "Max Mustermann"

font_size: 11

colors:
  title: "#1e66f5"
  accent: "#f38ba8"
  text: "#4c4f69"

fonts:
  zip: "fonts/fonts.zip"
  regular: "Inter-Regular.ttf"
  bold: "Inter-Bold.ttf"
  mono: "JetBrainsMono-Regular.ttf"

page_numbers:
  start_page: 2

layout:
  body: "justify"
  margins:
    left: 20
    right: 20
    top: 20
    bottom: 20

code_theme: "catppuccin-latte"
```

## Lizenz
© 2026 goDocGen Team. Die Nutzung ist für private und interne geschäftliche Zwecke gestattet. Der kommerzielle Verkauf der Software ist ausdrücklich untersagt. Weitere Details finden Sie in der [LICENSE](LICENSE) Datei.
