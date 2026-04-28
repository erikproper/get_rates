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
	"strconv"
	"time"

	"get_rates/config"
	"get_rates/internal/cache"
	"get_rates/internal/scraper"
	"get_rates/internal/sheets"
)

func main() {
	credFile  := flag.String("credentials", "credentials.json", "path to service account credentials JSON")
	cacheFile := flag.String("cache",       "cache/rates.json",  "path to rate cache file")
	fundsFile := flag.String("funds",       "funds.csv",         "path to funds CSV file")
	flag.Parse()

	funds, err := config.LoadFunds(*fundsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: loading funds: %v\n", err)
		os.Exit(1)
	}

	c, err := cache.Load(*cacheFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: loading cache: %v\n", err)
		os.Exit(1)
	}

	rows := [][]interface{}{{"ISIN", "Name", "Value", "Last update"}}

	for _, fund := range funds {
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
		var priceVal interface{} = price
		if f, err := strconv.ParseFloat(price, 64); err == nil {
			priceVal = f
		}
		rows = append(rows, []interface{}{fund.ISIN, fund.Name, priceVal, dateStr})
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
