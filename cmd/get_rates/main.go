/*
 *
 * Module:    GetRates
 * Package:   Main
 * Component: Scrape
 *
 * Orchestrates scraping of fund prices from all sources and pushes results to Google Sheets.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"get_rates/config"
	"get_rates/internal/cache"
	"get_rates/internal/scraper"
	"get_rates/internal/sheets"
)

func main() {
	credFile := flag.String("credentials", "credentials.json", "path to service account credentials JSON")
	cacheFile := flag.String("cache", "cache/rates.json", "path to rate cache file")
	flag.Parse()

	c, err := cache.Load(*cacheFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: loading cache: %v\n", err)
		os.Exit(1)
	}

	rows := [][]interface{}{{"ISIN", "Name", "Value", "Last update"}}

	for _, fund := range config.Funds {
		time.Sleep(250 * time.Millisecond)

		price, scrapeErr := scraper.Scrape(fund)
		fetchedAt := time.Now()

		if scrapeErr != nil {
			if entry, ok := c[fund.ISIN]; ok {
				fmt.Fprintf(os.Stderr,
					"WARN: %s (%s): scrape failed, using cached price %s from %s\n",
					fund.ISIN, fund.Name, entry.Price, entry.FetchedAt.Format(time.RFC3339))
				price = entry.Price
				fetchedAt = entry.FetchedAt
			} else {
				fmt.Fprintf(os.Stderr,
					"WARN: %s (%s): scrape failed, no cache available — price left empty\n",
					fund.ISIN, fund.Name)
				price = ""
				fetchedAt = time.Time{}
			}
		} else {
			c[fund.ISIN] = cache.TEntry{Price: price, FetchedAt: fetchedAt}
		}

		dateStr := ""
		if !fetchedAt.IsZero() {
			dateStr = fetchedAt.Format("02-01-2006 15:04")
		}
		fmt.Printf("Added %s  %s  %s\n", fund.ISIN, fund.Name, price)
		rows = append(rows, []interface{}{fund.ISIN, fund.Name, price, dateStr})
	}

	if err := cache.Save(*cacheFile, c); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: could not save cache: %v\n", err)
	}

	client, err := sheets.New(*credFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: connecting to Google Sheets: %v\n", err)
		os.Exit(1)
	}

	if err := client.WriteRates(rows); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: writing to sheet: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("OK: wrote %d rates to sheet\n", len(rows)-1)
}
