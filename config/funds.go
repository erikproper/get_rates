/*
 *
 * Module:    GetRates
 * Package:   Config
 * Component: Funds
 *
 * Loads the fund list from a CSV file (columns: ISIN, Name, Source).
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package config

import (
	"encoding/csv"
	"fmt"
	"os"

	"get_rates/internal/model"
)

var sourceByCode = map[string]model.TSource{
	"TGD": model.SourceTradegate,
	"FTD": model.SourceFTMarkets,
	"MSD": model.SourceMorningstar,
}

// LoadFunds reads the fund list from a CSV file with columns ISIN, Name, Source.
// The first row is treated as a header and skipped.
func LoadFunds(path string, separator rune) ([]model.TFund, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening funds file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = separator
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing funds CSV: %w", err)
	}

	var funds []model.TFund
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 3 {
			return nil, fmt.Errorf("line %d: expected 3 columns, got %d", i+1, len(row))
		}
		src, ok := sourceByCode[row[2]]
		if !ok {
			return nil, fmt.Errorf("line %d: unknown source %q", i+1, row[2])
		}
		funds = append(funds, model.TFund{ISIN: row[0], Name: row[1], Source: src})
	}

	if len(funds) == 0 {
		return nil, fmt.Errorf("no funds found in %s", path)
	}
	return funds, nil
}
