/*
 *
 * Module:    GetRates
 * Package:   Scraper
 * Component: Morningstar
 *
 * Scrapes the current price for a fund from the Morningstar global quote page.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package scraper

import (
	"fmt"
)

// ScrapeMorningstar fetches the current price for the given ISIN from Morningstar.
// Note: Morningstar loads prices via JavaScript; this implementation is a stub pending
// verification of whether a plain HTTP fetch returns usable data.
func ScrapeMorningstar(isin string) (string, error) {
	return "", fmt.Errorf("Morningstar scraper not yet implemented for %s", isin)
}
