/*
 *
 * Module:    GetRates
 * Package:   Scraper
 * Component: FT
 *
 * Scrapes the current price for a fund from the FT Markets summary page.
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

// ScrapeFT fetches the current price for the given ISIN from FT Markets.
func ScrapeFT(isin string) (string, error) {
	url := fmt.Sprintf("https://markets.ft.com/data/currencies/tearsheet/summary?s=%s", isin)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("building request for %s: %w", isin, err)
	}
	// FT requires a browser-like User-Agent to serve HTML rather than a redirect.
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; get_rates/1.0)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parsing page for %s: %w", isin, err)
	}

	price := ftPrice(doc)
	if price == "" {
		return "", fmt.Errorf("no price found on FT Markets for %s", isin)
	}
	return normaliseDecimal(price), nil
}

// ftPrice locates the price value on the FT summary page.
// The page has a label matching "Price (...)" and an associated value in the next element.
func ftPrice(doc *goquery.Document) string {
	var price string
	doc.Find("span.mod-ui-data-list__label").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if strings.HasPrefix(strings.TrimSpace(s.Text()), "Price") {
			price = strings.TrimSpace(s.Next().Text())
			return false
		}
		return true
	})
	return price
}
