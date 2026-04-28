/*
 *
 * Module:    GetRates
 * Package:   PDT
 * Component: PDT
 *
 * Builds simulated PDT transactions from current portfolio positions and rate cache,
 * then writes them to the Transactions tab of the Portfolio Dividend Tracker template.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package pdt

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"get_rates/config"
	"get_rates/internal/cache"
	"get_rates/internal/model"
	"get_rates/internal/sheets"
)

// --- public ---

// BuildAndPush computes PDT transactions for all positions in the Portfolio tab,
// applying breakdown.csv entries for BND funds, then writes them to the PDT template.
func BuildAndPush(
	client *sheets.TClient,
	positions []model.TPosition,
	fundsMap map[string]model.TFund,
	breakdown map[string][]config.TBreakdownEntry,
	c cache.TCache,
) error {
	rates := cacheRates(c)
	rows := buildRows(positions, fundsMap, breakdown, rates)
	if err := client.WriteTransactions(rows); err != nil {
		return fmt.Errorf("writing PDT transactions: %w", err)
	}
	fmt.Printf("OK: wrote %d PDT transactions\n", len(rows))
	return nil
}

// --- helpers ---

func cacheRates(c cache.TCache) map[string]float64 {
	rates := make(map[string]float64, len(c))
	for isin, entry := range c {
		if f, err := strconv.ParseFloat(entry.Price, 64); err == nil {
			rates[isin] = f
		}
	}
	return rates
}

func buildRows(
	positions []model.TPosition,
	fundsMap map[string]model.TFund,
	breakdown map[string][]config.TBreakdownEntry,
	rates map[string]float64,
) [][]interface{} {
	today := formatDate(time.Now())
	var rows [][]interface{}

	for _, pos := range positions {
		if entries, ok := breakdown[pos.ISIN]; ok {
			// BND fund: expand into underlying target funds
			netVal := pos.NetValue()
			for _, e := range entries {
				targetRate, ok := rates[e.TargetISIN]
				if !ok || targetRate == 0 {
					fmt.Fprintf(os.Stderr,
						"WARN: no rate for breakdown target %s (from %s) — skipped\n",
						e.TargetISIN, pos.ISIN)
					continue
				}
				targetAmt := netVal * e.Division / targetRate
				fund := fundsMap[e.TargetISIN]
				rows = append(rows, makeRow(
					"FlatEx", fund.Name, fund.AssetKind, e.TargetISIN, fund.Exchange,
					today, targetAmt, targetRate,
				))
			}
		} else {
			// Direct position: use current cached rate if available, else Portfolio price
			rate, ok := rates[pos.ISIN]
			if !ok || rate == 0 {
				rate = pos.Price
			}
			fund := fundsMap[pos.ISIN]
			assetKind, exchange := fund.AssetKind, fund.Exchange
			rows = append(rows, makeRow(
				"FlatEx", pos.Name, assetKind, pos.ISIN, exchange,
				today, pos.Position, rate,
			))
		}
	}
	return rows
}

func makeRow(broker, name, assetKind, isin, exchange, date string, amount, price float64) []interface{} {
	return []interface{}{broker, name, assetKind, isin, exchange, date, "13:00", "Buy", amount, price, "EUR"}
}

// formatDate produces the D-M-YYYY format that PDT expects (no leading zeros).
func formatDate(t time.Time) string {
	return fmt.Sprintf("%d-%d-%d", t.Day(), int(t.Month()), t.Year())
}
