/*
 *
 * Module:    GetRates
 * Package:   Scraper
 * Component: BND
 *
 * Scrapes the current NAV for a fund from the Brand New Day robot API.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 29.05.2026
 *
 */

package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

const bndAPI = "https://devrobotapi.azurewebsites.net/v1"

// bndISINToID caches the ISIN→fund-ID mapping built from the BND API on first use.
var (
	bndISINToID map[string]int
	bndMu       sync.Mutex
)

// --- JSON response types ---

type TBNDFundEntry struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TBNDFundList struct {
	Data []TBNDFundEntry `json:"data"`
}

type TBNDFactsheet struct {
	BasicCharacteristics struct {
		ISIN string `json:"isin"`
	} `json:"basicCharacteristics"`
}

type TBNDRateEntry struct {
	NAV float64 `json:"nav"`
}

type TBNDFundRates struct {
	Rates []TBNDRateEntry `json:"rates"`
}

// --- API helpers ---

func bndGet(url string, v any) error {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// buildBNDISINMap fetches all fund IDs from /v1/funds, then resolves each to an
// ISIN via /v1/fund/<id>/factsheetdata, returning a complete ISIN→ID map.
func buildBNDISINMap() (map[string]int, error) {
	var list TBNDFundList
	if err := bndGet(bndAPI+"/funds", &list); err != nil {
		return nil, fmt.Errorf("fetching BND fund list: %w", err)
	}

	m := make(map[string]int, len(list.Data))
	for _, fund := range list.Data {
		var fs TBNDFactsheet
		url := fmt.Sprintf("%s/fund/%d/factsheetdata", bndAPI, fund.ID)
		if err := bndGet(url, &fs); err != nil {
			continue // skip funds whose factsheet is unavailable
		}
		if isin := fs.BasicCharacteristics.ISIN; isin != "" {
			m[isin] = fund.ID
		}
	}
	return m, nil
}

// --- scraper ---

// ScrapeBND fetches the most recent NAV for the given ISIN from the BND robot API.
func ScrapeBND(isin string) (string, error) {
	bndMu.Lock()
	if bndISINToID == nil {
		m, err := buildBNDISINMap()
		if err != nil {
			bndMu.Unlock()
			return "", fmt.Errorf("building BND ISIN map: %w", err)
		}
		bndISINToID = m
	}
	id, ok := bndISINToID[isin]
	bndMu.Unlock()

	if !ok {
		return "", fmt.Errorf("ISIN %s not found in BND fund list", isin)
	}

	var rates TBNDFundRates
	if err := bndGet(fmt.Sprintf("%s/fundrates?id=%d", bndAPI, id), &rates); err != nil {
		return "", fmt.Errorf("fetching BND rates for ISIN %s: %w", isin, err)
	}
	if len(rates.Rates) == 0 {
		return "", fmt.Errorf("no rates returned for BND ISIN %s", isin)
	}
	return strconv.FormatFloat(rates.Rates[0].NAV, 'f', -1, 64), nil
}
