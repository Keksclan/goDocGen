// Package main ist der Einstiegspunkt für die godocgen CLI-Anwendung.
package main

import "godocgen/internal/cli"

// main initialisiert und führt die Root-Kommandozeilen-Schnittstelle aus.
func main() {
	cli.Execute()
}
