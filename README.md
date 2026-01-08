# goDocGen - Professional PDF DocBuilder

goDocGen ist ein CLI-Tool zur Erzeugung professioneller PDF-Dokumentation aus Markdown-Dateien. Es wurde speziell für die Erstellung von technischen Dokumentationen, API-Referenzen und System-Architekturen entwickelt.

## Features

- 🚀 **Markdown Support**: Volle Unterstützung für CommonMark inklusive GFM-Tabellen.
- 📊 **Erweiterte Mermaid Integration**: Flussdiagramme, Sequenzdiagramme, Klassendiagramme und State-Diagramme.
- 🎨 **Corporate Identity**: Volle Kontrolle über Farben, Schriftarten und Layouts (Default: Catppuccin Theme).
- 💻 **Modernes Code Rendering**: Syntax-Highlighting im IDE-Stil mit abgerundeten Containern und Sprach-Indikatoren.
- 🖱️ **Interaktives TUI**: Starten Sie das Interface mit `godocgen tui` für Erklärungen und Aktionen.
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

## Lizenz
© 2026 goDocGen Team. Die Nutzung ist für private und interne geschäftliche Zwecke gestattet. Der kommerzielle Verkauf der Software ist ausdrücklich untersagt. Weitere Details finden Sie in der [LICENSE](LICENSE) Datei.
