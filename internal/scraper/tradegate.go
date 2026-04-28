/*
 *
 * Module:    GetRates
 * Package:   Scraper
 * Component: Tradegate
 *
 * Scrapes the current price for a fund from the Tradegate exchange order book.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package scraper

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ScrapeTradegate fetches the last price for the given ISIN from Tradegate,
// falling back to the bid price if no last price is available.
func ScrapeTradegate(isin string) (string, error) {
	url := fmt.Sprintf("https://www.tradegate.de/orderbuch.php?lang=en&isin=%s", isin)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parsing page for %s: %w", isin, err)
	}

	// Try last price first, fall back to bid.
	// Tradegate renders "./." when no trades have occurred today — treat as absent.
	price := tradegatePrice(doc, "last")
	if !hasDigit(price) {
		price = tradegatePrice(doc, "bid")
	}
	if !hasDigit(price) {
		return "", fmt.Errorf("no valid price found on Tradegate for %s", isin)
	}
	return normaliseDecimal(price), nil
}

// tradegatePrice finds the price for the strong element with the given id (last or bid).
func tradegatePrice(doc *goquery.Document, id string) string {
	var price string
	doc.Find("td.longprice strong#" + id).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		price = strings.TrimSpace(s.Text())
		return false
	})
	return price
}
