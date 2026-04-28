/*
 *
 * Module:    GetRates
 * Package:   Config
 * Component: Config
 *
 * Loads tool configuration from a simple INI file.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// TConfig holds configurable file paths and format settings for the get_rates tools.
type TConfig struct {
	FundsFile     string
	BreakdownFile string
	CacheFile     string
	Separator     rune
}

// --- defaults ---

func defaultConfig() TConfig {
	return TConfig{
		FundsFile:     "funds.csv",
		BreakdownFile: "breakdown.csv",
		CacheFile:     "cache/rates.json",
		Separator:     ';',
	}
}

// --- load ---

// LoadConfig reads a simple INI file and returns a TConfig.
// Only the [paths] section is recognised; unknown keys are silently ignored.
// If the file does not exist, sensible defaults are returned without error.
func LoadConfig(path string) (TConfig, error) {
	cfg := defaultConfig()

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	inPaths := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == ';' || line[0] == '#' {
			continue
		}
		if line == "[paths]" {
			inPaths = true
			continue
		}
		if line[0] == '[' {
			inPaths = false
			continue
		}
		if !inPaths {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "funds":
			cfg.FundsFile = strings.TrimSpace(val)
		case "breakdown":
			cfg.BreakdownFile = strings.TrimSpace(val)
		case "cache":
			cfg.CacheFile = strings.TrimSpace(val)
		case "separator":
			if r := []rune(strings.TrimSpace(val)); len(r) == 1 {
				cfg.Separator = r[0]
			}
		}
	}
	return cfg, scanner.Err()
}
