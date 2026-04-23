package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

var (
	ErrNoYAMLForConversion = errors.New("no mk.yaml or mk.yml found to convert")
	ErrNoJSONForConversion = errors.New("no mk.json found to convert")
)

func FindForConversion(start, targetFormat string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	names := []string{"mk.json"}
	if targetFormat == "json" {
		names = []string{"mk.yaml", "mk.yml"}
	}

	for {
		for _, name := range names {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			if targetFormat == "json" {
				return "", ErrNoYAMLForConversion
			}
			return "", ErrNoJSONForConversion
		}
		dir = parent
	}
}

func ConvertPath(sourcePath, targetFormat string) string {
	ext := ".yaml"
	if targetFormat == "json" {
		ext = ".json"
	}
	return strings.TrimSuffix(sourcePath, filepath.Ext(sourcePath)) + ext
}

func SortedNames(cfg Config) []string {
	names := make([]string, 0, len(cfg.Commands))
	for name := range cfg.Commands {
		names = append(names, name)
	}
	return sortStrings(names)
}

func sortStrings(values []string) []string {
	sort.Strings(values)
	return values
}

func Equal(a, b Config) bool {
	return reflect.DeepEqual(a, b)
}
