package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadEnvFiles(baseDir string, files EnvFiles) (map[string]string, error) {
	env := map[string]string{}
	for _, file := range files {
		path := file
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}
		values, err := LoadEnvFile(path)
		if err != nil {
			return nil, err
		}
		for key, value := range values {
			env[key] = value
		}
	}
	return env, nil
}

func LoadEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	env := map[string]string{}
	scanner := bufio.NewScanner(f)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("%s:%d: expected KEY=value", path, lineNumber)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("%s:%d: empty env key", path, lineNumber)
		}
		env[key] = cleanEnvValue(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return env, nil
}

func cleanEnvValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		first := value[0]
		last := value[len(value)-1]
		if first == last && (first == '"' || first == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}
