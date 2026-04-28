/*
 *
 * Module:    GetRates
 * Package:   Scraper
 * Component: Scrape
 *
 * Dispatches scrape requests to the appropriate source scraper and normalises decimal separators.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package scraper

import (
	"fmt"
	"strings"
	"time"

	"get_rates/internal/model"
)

// --- dispatch ---

func Scrape(fund model.TFund) (string, error) {
	var err error
	for range 3 {
		var price string
		price, err = scrapeOnce(fund)
		if err == nil {
			return price, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", err
}

func scrapeOnce(fund model.TFund) (string, error) {
	switch fund.Source {
	case model.SourceTradegate:
		return ScrapeTradegate(fund.ISIN)
	case model.SourceFTMarkets:
		return ScrapeFT(fund.ISIN)
	case model.SourceMorningstar:
		return ScrapeMorningstar(fund.ISIN)
	default:
		return "", fmt.Errorf("unknown source for ISIN %s", fund.ISIN)
	}
}

// --- price validation ---

// hasDigit reports whether s contains at least one ASCII digit.
// Used to reject Tradegate's "./." placeholder for no-trade-today prices.
func hasDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// --- decimal normalisation ---

// normaliseDecimal converts European decimal notation (comma) to dot notation,
// stripping any thousands separators in the process.
func normaliseDecimal(price string) string {
	price = strings.TrimSpace(price)
	if strings.Contains(price, ".") && strings.Contains(price, ",") {
		// European format: 1.234,56 → 1234.56
		price = strings.ReplaceAll(price, ".", "")
		price = strings.ReplaceAll(price, ",", ".")
	} else {
		price = strings.ReplaceAll(price, ",", ".")
	}
	return price
}
