package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("error parsing yaml: %w", err)
	}

	if cfg.Theme == "" {
		cfg.Theme = "catppuccin-mocha"
	}
	ApplyTheme(cfg)
	setDefaults(cfg)

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return cfg, nil
}

func ApplyTheme(cfg *Config) {
	if cfg.Theme == "" {
		return
	}

	switch cfg.Theme {
	case "catppuccin-latte":
		if cfg.Colors.Title == "" {
			cfg.Colors.Title = "#8839ef" // Latte Mauve
		}
		if cfg.Colors.Header == "" {
			cfg.Colors.Header = "#1e66f5" // Latte Blue
		}
		if cfg.Colors.Background == "" {
			cfg.Colors.Background = "#eff1f5"
		}
		if cfg.Colors.Text == "" {
			cfg.Colors.Text = "#4c4f69"
		}
		if cfg.Colors.Accent == "" {
			cfg.Colors.Accent = "#ea76cb"
		}
		if cfg.CodeTheme == "" {
			cfg.CodeTheme = "catppuccin-latte"
		}
	case "catppuccin-frappe":
		if cfg.Colors.Title == "" {
			cfg.Colors.Title = "#ca9ee6" // Frappe Mauve
		}
		if cfg.Colors.Header == "" {
			cfg.Colors.Header = "#8caaee" // Frappe Blue
		}
		if cfg.Colors.Background == "" {
			cfg.Colors.Background = "#303446"
		}
		if cfg.Colors.Text == "" {
			cfg.Colors.Text = "#c6d0f5"
		}
		if cfg.Colors.Accent == "" {
			cfg.Colors.Accent = "#f4b8e4"
		}
		if cfg.CodeTheme == "" {
			cfg.CodeTheme = "catppuccin-frappe"
		}
	case "catppuccin-macchiato":
		if cfg.Colors.Title == "" {
			cfg.Colors.Title = "#c6a0f6" // Macchiato Mauve
		}
		if cfg.Colors.Header == "" {
			cfg.Colors.Header = "#8aadf4" // Macchiato Blue
		}
		if cfg.Colors.Background == "" {
			cfg.Colors.Background = "#24273a"
		}
		if cfg.Colors.Text == "" {
			cfg.Colors.Text = "#cad3f5"
		}
		if cfg.Colors.Accent == "" {
			cfg.Colors.Accent = "#f5bde6"
		}
		if cfg.CodeTheme == "" {
			cfg.CodeTheme = "catppuccin-macchiato"
		}
	case "catppuccin-mocha":
		if cfg.Colors.Title == "" {
			cfg.Colors.Title = "#cba6f7" // Mocha Mauve
		}
		if cfg.Colors.Header == "" {
			cfg.Colors.Header = "#89b4fa" // Mocha Blue
		}
		if cfg.Colors.Background == "" {
			cfg.Colors.Background = "#1e1e2e"
		}
		if cfg.Colors.Text == "" {
			cfg.Colors.Text = "#cdd6f4"
		}
		if cfg.Colors.Accent == "" {
			cfg.Colors.Accent = "#f5c2e7"
		}
		if cfg.CodeTheme == "" {
			cfg.CodeTheme = "catppuccin-mocha"
		}
	}
}

func setDefaults(cfg *Config) {
	if cfg.Colors.Title == "" {
		cfg.Colors.Title = "#1e66f5" // Default Blue (statt E.ON Red)
	}
	if cfg.Colors.Header == "" {
		cfg.Colors.Header = "#1e66f5" // Default Blue
	}
	if cfg.Mermaid.Renderer == "" {
		cfg.Mermaid.Renderer = "mmdc"
	}
	if cfg.Layout.StartPage == "" {
		cfg.Layout.StartPage = "center"
	}
	if cfg.Layout.Body == "" {
		cfg.Layout.Body = "justify"
	}
	if cfg.Layout.Margins.Left == 0 {
		cfg.Layout.Margins.Left = 10
	}
	if cfg.Layout.Margins.Right == 0 {
		cfg.Layout.Margins.Right = 10
	}
	if cfg.Layout.Margins.Top == 0 {
		cfg.Layout.Margins.Top = 10
	}
	if cfg.Layout.Margins.Bottom == 0 {
		cfg.Layout.Margins.Bottom = 10
	}
	if cfg.Gradient.Orientation == "" {
		cfg.Gradient.Orientation = "vertical"
	}
	if cfg.FontSize == 0 {
		cfg.FontSize = 12.0
	}
	if cfg.CodeTheme == "" {
		cfg.CodeTheme = "catppuccin-latte"
	}
}
