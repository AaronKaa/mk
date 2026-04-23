package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AaronKaa/mk/internal/config"
)

func convertConfig(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: mk --convert <json|yaml>")
	}

	targetFormat := strings.ToLower(args[0])
	if targetFormat != "json" && targetFormat != "yaml" && targetFormat != "yml" {
		return fmt.Errorf("unknown convert format %q; use json or yaml", args[0])
	}
	if targetFormat == "yml" {
		targetFormat = "yaml"
	}

	sourcePath, err := config.FindForConversion(".", targetFormat)
	if err != nil {
		return err
	}
	cfg, err := config.Load(sourcePath)
	if err != nil {
		return err
	}

	targetPath := config.ConvertPath(sourcePath, targetFormat)
	if targetPath == sourcePath {
		return fmt.Errorf("%s is already %s", filepath.Base(sourcePath), targetFormat)
	}
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("%s already exists; remove it before converting", targetPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := config.Save(targetPath, cfg); err != nil {
		return err
	}
	converted, err := config.Load(targetPath)
	if err != nil {
		return fmt.Errorf("converted file failed to load: %w", err)
	}
	if !config.Equal(cfg, converted) {
		return fmt.Errorf("converted file does not match source config")
	}
	if err := os.Remove(sourcePath); err != nil {
		return fmt.Errorf("converted %s -> %s, but failed to remove source: %w", sourcePath, targetPath, err)
	}
	fmt.Fprintf(stdout, "converted %s -> %s\n", sourcePath, targetPath)
	return nil
}
