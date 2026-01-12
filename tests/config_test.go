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

	if cfg.Colors.Title != "#cba6f7" {
		t.Errorf("Expected default color #cba6f7, got %s", cfg.Colors.Title)
	}
}

func TestStartPageValidation(t *testing.T) {
	content := `
title: "Test Doc"
font_size: 12
fonts:
  zip: "fonts.zip"
  regular: "font.ttf"
layout:
  startpage: "center"
`
	err := os.WriteFile("test_startpage_valid.yml", []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_startpage_valid.yml")

	_, err = config.LoadConfig("test_startpage_valid.yml")
	if err != nil {
		t.Errorf("Expected no error for valid start_page, got %v", err)
	}
}

func TestStartPageInvalidValidation(t *testing.T) {
	content := `
title: "Test Doc"
font_size: 12
fonts:
  zip: "fonts.zip"
  regular: "font.ttf"
layout:
  startpage: "invalid"
`
	err := os.WriteFile("test_startpage_invalid.yml", []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_startpage_invalid.yml")

	_, err = config.LoadConfig("test_startpage_invalid.yml")
	if err == nil {
		t.Error("Expected error for invalid start_page, got nil")
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
