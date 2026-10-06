/*
 *
 * Module:    GetRates
 * Package:   Scraper
 * Component: Morningstar
 *
 * Fetches the latest closing price for a fund from the Morningstar screener API.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 05.10.2026
 *
 */

package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// morningstarAPI is the public screener endpoint used by Morningstar's own fund
// tools. The universe FONLD$$ALL covers all funds available in the Netherlands.
const morningstarAPI = "https://lt.morningstar.com/api/rest.svc/klr5zyak8x/security/screener"

// --- JSON response types ---

type TMorningstarRow struct {
	ISIN           string  `json:"isin"`
	ClosePrice     float64 `json:"ClosePrice"`
	ClosePriceDate string  `json:"ClosePriceDate"`
}

type TMorningstarResult struct {
	Rows []TMorningstarRow `json:"rows"`
}

// --- scraper ---

// ScrapeMorningstar fetches the latest closing price for the given ISIN from Morningstar.
func ScrapeMorningstar(isin string) (string, error) {
	q := url.Values{}
	q.Set("page", "1")
	q.Set("pageSize", "10")
	q.Set("outputType", "json")
	q.Set("version", "1")
	q.Set("languageId", "nl-NL")
	q.Set("currencyId", "EUR")
	q.Set("universeIds", "FONLD$$ALL")
	q.Set("securityDataPoints", "isin|ClosePrice|ClosePriceDate")
	q.Set("term", isin)

	resp, err := http.Get(morningstarAPI + "?" + q.Encode()) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("fetching Morningstar data for %s: %w", isin, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d fetching Morningstar data for %s", resp.StatusCode, isin)
	}

	var result TMorningstarResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parsing Morningstar data for %s: %w", isin, err)
	}

	// The search term may match more than one security; only accept an exact ISIN hit.
	for _, row := range result.Rows {
		if row.ISIN == isin && row.ClosePrice > 0 {
			return strconv.FormatFloat(row.ClosePrice, 'f', -1, 64), nil
		}
	}
	return "", fmt.Errorf("no price found on Morningstar for %s", isin)
}
