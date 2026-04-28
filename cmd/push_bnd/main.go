/*
 *
 * Module:    GetRates
 * Package:   Main
 * Component: PushBND
 *
 * Reads a clipboard copy of the BND portfolio page and pushes it to Google Sheets.
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

	if _, err := config.LoadConfig(*configFile); err != nil {
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
}
