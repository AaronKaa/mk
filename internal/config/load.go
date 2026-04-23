package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ErrNotFound = errors.New("no mk.json or mk.yaml found")

func FindAndLoad(start string) (Config, string, error) {
	source, err := FindAndLoadSource(start)
	if err != nil {
		return Config{}, "", err
	}
	return source.Config, source.Path, nil
}

func FindAndLoadSource(start string) (Source, error) {
	path, err := Find(start)
	if err != nil {
		if env := os.Getenv(EnvCommandsVar); env != "" {
			cfg, err := LoadEnvCommands(env)
			if err != nil {
				return Source{}, err
			}
			baseDir, err := filepath.Abs(start)
			if err != nil {
				return Source{}, err
			}
			return Source{Config: cfg, Path: EnvCommandsPath, BaseDir: baseDir, Inherited: map[string]bool{}, Hidden: map[string]bool{}}, nil
		}
		return Source{}, err
	}

	cfg, err := Load(path)
	if err != nil {
		return Source{}, err
	}
	if env := os.Getenv(EnvCommandsVar); env != "" {
		prefix := os.Getenv(EnvCommandsPrefixVar)
		envCfg, err := LoadEnvCommands(env)
		if err != nil {
			return Source{}, err
		}
		merged, err := MergePrefixed(&cfg, envCfg, prefix)
		if err != nil {
			return Source{}, err
		}
		return Source{Config: cfg, Path: path, BaseDir: filepath.Dir(path), Inherited: merged.Inherited, Hidden: merged.Hidden}, nil
	}
	return Source{Config: cfg, Path: path, BaseDir: filepath.Dir(path), Inherited: map[string]bool{}, Hidden: map[string]bool{}}, nil
}

func Find(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for {
		for _, name := range []string{"mk.json", "mk.yaml", "mk.yml"} {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%w; run `mk init` to create one or set MK_COMMANDS", ErrNotFound)
		}
		dir = parent
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg, err := loadBytes(data, filepath.Ext(path), path)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func LoadEnvCommands(value string) (Config, error) {
	cfg, err := loadBytes([]byte(value), "", EnvCommandsVar)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func loadBytes(data []byte, ext, source string) (Config, error) {
	var cfg Config
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("load %s: %w", source, err)
		}
	default:
		if err := json.Unmarshal(data, &cfg); err != nil {
			if yamlErr := yaml.Unmarshal(data, &cfg); yamlErr != nil {
				return Config{}, fmt.Errorf("load %s as json: %w", source, err)
			}
		}
	}

	if cfg.Commands == nil {
		cfg.Commands = map[string]Command{}
	}
	if err := NormalizeAliases(&cfg); err != nil {
		return Config{}, err
	}
	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
