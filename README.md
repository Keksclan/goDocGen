# goDocGen - Professional PDF DocBuilder

goDocGen ist ein CLI-Tool zur Erzeugung professioneller PDF-Dokumentation aus Markdown-Dateien. Es wurde speziell für die Erstellung von technischen Dokumentationen, API-Referenzen und System-Architekturen entwickelt.

## Features

- 🚀 **Markdown Support**: Volle Unterstützung für CommonMark inklusive GFM-Tabellen.
- 📊 **Erweiterte Mermaid Integration**: Flussdiagramme, Sequenzdiagramme, Klassendiagramme und State-Diagramme.
- 🎨 **Corporate Identity**: Volle Kontrolle über Farben, Schriftarten und Layouts (Default: Catppuccin Theme).
- 💻 **Modernes Code Rendering**: Syntax-Highlighting im IDE-Stil mit abgerundeten Containern und Sprach-Indikatoren.
- 🖱️ **Interaktives TUI**: Starten Sie das Interface mit `./godocgen.exe tui` für Erklärungen und Aktionen.
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
go build -o godocgen.exe ./cmd/docgen
```

## Quick Start

### 1. Projekt initialisieren
Erzeugt eine fertige Struktur mit Beispiel-Content und Konfiguration:
```bash
./godocgen.exe init my_docs
```

### 2. PDF generieren
Baut das Projekt und speichert das Ergebnis (mit automatischer Versionierung) im `dist` Ordner:
```bash
./godocgen.exe build --project ./my_docs --out ./dist
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

## Lizenz
© 2026 goDocGen Team. Die Nutzung ist für private und interne geschäftliche Zwecke gestattet. Der kommerzielle Verkauf der Software ist ausdrücklich untersagt. Weitere Details finden Sie in der [LICENSE](LICENSE) Datei.
