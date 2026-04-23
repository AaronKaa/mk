package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Save(path string, cfg Config) error {
	if cfg.Commands == nil {
		cfg.Commands = map[string]Command{}
	}

	var (
		data []byte
		err  error
	)
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(cfg)
	default:
		data, err = json.MarshalIndent(cfg, "", "  ")
		if err == nil {
			data = append(data, '\n')
		}
	}
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
