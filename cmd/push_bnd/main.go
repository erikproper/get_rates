/*
 *
 * Module:    GetRates
 * Package:   Main
 * Component: PushBND
 *
 * Reads a clipboard copy of the BND portfolio page, pushes it to Google Sheets,
 * and refreshes the PDT Transactions tab.
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
	"io"
	"os"
	"os/exec"
	"time"

	"get_rates/config"
	"get_rates/internal/bnd"
	"get_rates/internal/cache"
	"get_rates/internal/model"
	"get_rates/internal/pdt"
	"get_rates/internal/sheets"
)

// readInput returns clipboard contents via pbpaste, or reads stdin if it is piped.
func readInput() ([]byte, error) {
	fi, err := os.Stdin.Stat()
	if err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		return io.ReadAll(os.Stdin)
	}
	return exec.Command("pbpaste").Output()
}

func main() {
	configFile := flag.String("config",      "config.ini",       "path to config file")
	credFile   := flag.String("credentials", "credentials.json", "path to service account credentials JSON")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: loading config: %v\n", err)
		os.Exit(1)
	}

	data, err := readInput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: reading input: %v\n", err)
		os.Exit(1)
	}

	rows, err := bnd.Parse(string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	sheetRows := make([][]interface{}, 0, len(rows)+1)
	for _, row := range rows {
		if row.Name != "" {
			fmt.Printf("Added %s  %s\n", row.Name, row.Waarde)
		}
		sheetRows = append(sheetRows, row.ToSlice())
	}
	sheetRows = append(sheetRows, []interface{}{"Last update", time.Now().Format("02-01-2006 15:04")})

	client, err := sheets.New(*credFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: connecting to Google Sheets: %v\n", err)
		os.Exit(1)
	}

	if err := client.WriteBND(sheetRows); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: writing to sheet: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK: pushed %d BND rows to sheet\n", len(rows))

	// Refresh PDT Transactions tab
	funds, err := config.LoadFunds(cfg.FundsFile, cfg.Separator)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: loading funds for PDT: %v\n", err)
		return
	}

	breakdown, err := config.LoadBreakdown(cfg.BreakdownFile, cfg.Separator)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: loading breakdown: %v\n", err)
		return
	}
	config.ValidateBreakdown(breakdown)

	c, err := cache.Load(cfg.CacheFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: loading rate cache for PDT: %v\n", err)
		return
	}

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
}
