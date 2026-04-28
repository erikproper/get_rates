/*
 *
 * Module:    GetRates
 * Package:   BND
 * Component: Parser
 *
 * Parses a clipboard copy of the BND portfolio page and extracts normalised fund rows.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package bnd

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	tableHeader = "Fondsnaam"
	tableEnd    = "Meer informatie"
)

// TRow holds one parsed and normalised row from the BND portfolio table.
type TRow struct {
	Name   string
	Aantal string
	Koers  string
	Datum  string
	Weging string
	Waarde string
}

// --- parse ---

// Parse extracts fund rows from a clipboard copy of the BND portfolio page.
// Returns an error if the input does not look like a valid BND page copy.
func Parse(input string) ([]TRow, error) {
	if !strings.Contains(input, tableHeader) {
		return nil, fmt.Errorf("input does not look like a copy from the BND portfolio page (table header not found)")
	}

	var rows []TRow
	inTable := false

	for _, raw := range strings.Split(input, "\n") {
		line := strings.TrimRight(raw, " \t\r")

		if strings.HasPrefix(line, tableHeader) {
			inTable = true
			continue
		}
		if inTable && strings.HasPrefix(line, tableEnd) {
			break
		}
		if !inTable || line == "" {
			continue
		}

		rows = append(rows, parseRow(strings.Split(line, "\t")))
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no fund rows found in input")
	}
	return rows, nil
}

func parseRow(fields []string) TRow {
	get := func(i int) string {
		if i < len(fields) {
			return strings.TrimSpace(fields[i])
		}
		return ""
	}
	return TRow{
		Name:   get(0),
		Aantal: normaliseField(get(1)),
		Koers:  normaliseField(get(2)),
		Datum:  get(3),
		Weging: normaliseField(get(4)),
		Waarde: normaliseField(get(5)),
	}
}

// ToSlice returns the row as a slice suitable for the Sheets API.
func (r TRow) ToSlice() []interface{} {
	return []interface{}{r.Name, r.Aantal, r.Koers, r.Datum, r.Weging, r.Waarde}
}

// --- normalisation ---

// normaliseField converts a European-formatted value to dot-decimal notation.
// It handles € prefixes, thousands-dot separators, decimal commas, and trailing % signs.
func normaliseField(s string) string {
	if s == "" {
		return ""
	}
	euroPrefix := strings.HasPrefix(s, "€")
	if euroPrefix {
		s = strings.TrimSpace(s[len("€"):])
	}
	s = strings.TrimSuffix(s, "%")
	s = strings.TrimSpace(s)
	if strings.Contains(s, ".") && strings.Contains(s, ",") {
		// European thousands format: 1.234,56 → 1234.56
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else {
		s = strings.ReplaceAll(s, ",", ".")
	}
	if euroPrefix {
		return "€ " + s
	}
	return s
}
