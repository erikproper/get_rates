/*
 *
 * Module:    GetRates
 * Package:   Main
 * Component: Scrape
 *
 * Orchestrates scraping of fund prices from all sources, pushes results to Google Sheets,
 * and refreshes the PDT Transactions tab.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 29.04.2026
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
	"get_rates/internal/model"
	"get_rates/internal/pdt"
	"get_rates/internal/scraper"
	"get_rates/internal/sheets"
)

func main() {
	configFile := flag.String("config",      "config.ini",       "path to config file")
	credFile   := flag.String("credentials", "credentials.json", "path to service account credentials JSON")
	cacheFlag  := flag.String("cache",       "",                 "path to rate cache file (overrides config)")
	fundsFlag  := flag.String("funds",       "",                 "path to funds CSV (overrides config)")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: loading config: %v\n", err)
		os.Exit(1)
	}
	fundsFile := cfg.FundsFile
	if *fundsFlag != "" {
		fundsFile = *fundsFlag
	}
	cacheFile := cfg.CacheFile
	if *cacheFlag != "" {
		cacheFile = *cacheFlag
	}

	funds, err := config.LoadFunds(fundsFile, cfg.Separator)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: loading funds: %v\n", err)
		os.Exit(1)
	}

	c, err := cache.Load(cacheFile)
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

	if err := cache.Save(cacheFile, c); err != nil {
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

	// Refresh PDT Transactions tab
	breakdown, err := config.LoadBreakdown(cfg.BreakdownFile, cfg.Separator)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: loading breakdown: %v\n", err)
		return
	}
	config.ValidateBreakdown(breakdown)

	positions, err := client.ReadPortfolio()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: reading portfolio: %v\n", err)
		return
	}

	fundsMap := make(map[string]model.TFund, len(funds))
	for _, f := range funds {
		fundsMap[f.ISIN] = f
	}

	if err := pdt.BuildAndPush(client, positions, fundsMap, breakdown, c); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: updating PDT: %v\n", err)
	}

	if err := pdt.PushBooking(client); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: pushing portfolio booking: %v\n", err)
	}
}
