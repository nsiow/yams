package cli

import (
	"log/slog"
	"os"
	"path/filepath"

	json "github.com/bytedance/sonic"
)

// Config represents the yams configuration file structure
type Config struct {
	Server string `json:"server"`
	Format string `json:"format"`
}

// ConfigPaths returns the list of paths to check for config files (in priority order)
func ConfigPaths() []string {
	var paths []string

	// XDG config dir
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		paths = append(paths, filepath.Join(xdg, "yams", "config.json"))
	}

	// ~/.config/yams/config.json
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "yams", "config.json"))
	}

	return paths
}

// LoadConfig loads the configuration from the first available config file
func LoadConfig() *Config {
	for _, path := range ConfigPaths() {
		cfg, err := loadConfigFile(path)
		if err == nil {
			slog.Debug("loaded config file", "path", path)
			return cfg
		}
		if !os.IsNotExist(err) {
			slog.Debug("error loading config file", "path", path, "error", err)
		}
	}
	return nil
}

func loadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
