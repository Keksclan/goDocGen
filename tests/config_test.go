package tests

import (
	"godocgen/internal/config"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `
title: "Test Doc"
font_size: 12
fonts:
  zip: "fonts.zip"
  regular: "font.ttf"
`
	err := os.WriteFile("test_docgen.yml", []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_docgen.yml")

	cfg, err := config.LoadConfig("test_docgen.yml")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cfg.Title != "Test Doc" {
		t.Errorf("Expected title 'Test Doc', got '%s'", cfg.Title)
	}

	if cfg.Colors.Title != "#1e66f5" {
		t.Errorf("Expected default color #1e66f5, got %s", cfg.Colors.Title)
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	content := `
title: ""
`
	os.WriteFile("invalid_docgen.yml", []byte(content), 0644)
	defer os.Remove("invalid_docgen.yml")

	_, err := config.LoadConfig("invalid_docgen.yml")
	if err == nil {
		t.Error("Expected error for missing required fields, got nil")
	}
}

