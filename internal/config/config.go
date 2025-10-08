package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config encapsulates all config required by the application.
type Config struct {
	HttpServer struct {
		Addr string `json:"addr"`
	} `json:"httpServer"`

	Logger struct {
		Level  string `json:"level"`
		Pretty bool   `json:"pretty"`
	} `json:"logger"`
}

// Load config from the given JSON file.
func Load(jsonPath string) (Config, error) {
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file at %s because: %w", jsonPath, err)
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config file at %s because: %w", jsonPath, err)
	}

	return config, nil
}
