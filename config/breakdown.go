/*
 *
 * Module:    GetRates
 * Package:   Config
 * Component: Breakdown
 *
 * Loads the BND fund breakdown table from a CSV file.
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
	"math"
	"os"
	"strconv"
	"strings"
)

// TBreakdownEntry maps a fraction of a source fund's net value to a target fund.
type TBreakdownEntry struct {
	SourceISIN string
	Division   float64 // fraction, 0–1
	TargetISIN string
}

// --- load ---

// LoadBreakdown reads the breakdown CSV (columns: Source ISIN, Division, Target ISIN)
// and returns a map from source ISIN to its breakdown entries.
// Division values are expected as decimal fractions using the locale's decimal separator.
// If the file does not exist, nil is returned without error.
func LoadBreakdown(path string, sep rune) (map[string][]TBreakdownEntry, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("opening breakdown file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = sep
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing breakdown CSV: %w", err)
	}

	result := make(map[string][]TBreakdownEntry)
	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			return nil, fmt.Errorf("line %d: expected 3 columns, got %d", i+1, len(row))
		}
		divStr := strings.ReplaceAll(strings.TrimSpace(row[1]), ",", ".")
		div, err := strconv.ParseFloat(divStr, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid division %q: %w", i+1, row[1], err)
		}
		src := strings.TrimSpace(row[0])
		result[src] = append(result[src], TBreakdownEntry{
			SourceISIN: src,
			Division:   div,
			TargetISIN: strings.TrimSpace(row[2]),
		})
	}
	return result, nil
}

// --- validation ---

// ValidateBreakdown warns to stderr for any source ISIN whose divisions do not sum to 1.
func ValidateBreakdown(breakdown map[string][]TBreakdownEntry) {
	for isin, entries := range breakdown {
		var sum float64
		for _, e := range entries {
			sum += e.Division
		}
		if math.Abs(sum-1.0) > 0.0001 {
			fmt.Fprintf(os.Stderr,
				"WARN: breakdown for %s sums to %.4f, expected 1.0\n", isin, sum)
		}
	}
}
