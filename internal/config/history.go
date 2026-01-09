package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// GlobalConfig speichert tool-weite Einstellungen wie die Historie der Projekte.
type GlobalConfig struct {
	RecentProjects []string `json:"recent_projects"`
	Language       string   `json:"language"` // "de" oder "en"
}

// GetGlobalConfigPath gibt den Pfad zur globalen Konfigurationsdatei zurück.
func GetGlobalConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "godocgen")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(dir, "history.json"), nil
}

// LoadGlobalConfig lädt die globale Konfiguration.
func LoadGlobalConfig() (*GlobalConfig, error) {
	path, err := GetGlobalConfigPath()
	if err != nil {
		return &GlobalConfig{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return &GlobalConfig{RecentProjects: []string{}}, nil
	}

	var cfg GlobalConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return &GlobalConfig{RecentProjects: []string{}, Language: "de"}, nil
	}

	if cfg.Language == "" {
		cfg.Language = "de"
	}

	return &cfg, nil
}

// SaveGlobalConfig speichert die globale Konfiguration.
func (c *GlobalConfig) Save() error {
	path, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddProject fügt einen Projektpfad zur Historie hinzu und schiebt ihn an die Spitze.
func (c *GlobalConfig) AddProject(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// Entferne Duplikate
	newRecent := []string{absPath}
	for _, p := range c.RecentProjects {
		if p != absPath {
			newRecent = append(newRecent, p)
		}
	}

	// Limitiere auf die letzten 10 Projekte
	if len(newRecent) > 10 {
		newRecent = newRecent[:10]
	}

	c.RecentProjects = newRecent
}
