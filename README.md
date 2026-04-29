# get_rates

Two Go tools that replace a manual clipboard-based portfolio workflow:

- **`get_rates`** — scrapes current prices for a fixed list of funds/ETFs from
  [Tradegate](https://www.tradegate.de) and [FT Markets](https://markets.ft.com),
  writes the results to the `Import rates` tab of the Portfolio sheet, and refreshes
  the Transactions tab of the Portfolio Dividend Tracker template.
- **`push_bnd`** — reads a clipboard copy of the BND pension-account portfolio page,
  normalises the data, writes it to the `Import from BND` tab, and likewise refreshes
  the PDT Transactions tab.

Both tools apply a configurable BND breakdown (`breakdown.csv`) to convert BND pension
positions into their underlying fund positions before writing to the PDT template.

---

## Prerequisites

- Go 1.22+
- A Google Cloud project with the **Google Sheets API** enabled
- A **service account** with Editor access to both the Portfolio sheet and the
  Portfolio Dividend Tracker template sheet
- The service account's JSON credentials file

---

## Installation

### 1. Clone and build

```bash
git clone https://github.com/erikproper/get_rates.git
cd get_rates
go install ./cmd/get_rates/
go install ./cmd/push_bnd/
```

Binaries are placed in `$GOPATH/bin`. Make sure that directory is on your `PATH`
(add `export PATH="$PATH:$(go env GOPATH)/bin"` to your shell profile if needed).

### 2. Set up Google Cloud credentials

1. Create a project in [Google Cloud Console](https://console.cloud.google.com).
2. Enable the **Google Sheets API**.
3. Create a **service account** and download its JSON key file.
4. Share both the Portfolio sheet and the PDT template sheet with the service
   account's email address (Editor access).
5. Save the key file as `credentials.json` in your working directory.

### 3. Create the working directory

Run the tools from a dedicated directory that contains:

```
credentials.json     ← service account key (never commit this)
config.ini           ← tool configuration (see below)
funds.csv            ← list of funds to scrape
breakdown.csv        ← BND breakdown table
cache/               ← created automatically on first run
```

A sample `config.ini` and both CSV files are included in the repository.

### 4. Verify the sheet IDs

The sheet IDs are compiled into the binary (`internal/sheets/client.go`).
If you use your own spreadsheets, update the two constants before building:

```go
portfolioSheetID = "..."   // Portfolio sheet
pdtSheetID       = "..."   // Portfolio Dividend Tracker template
```

---

## Usage

### get_rates

Run from the directory containing `credentials.json` and `config.ini`:

```bash
get_rates
```

This scrapes all 41 funds (25 from Tradegate, 16 from FT Markets), writes prices to
the `Import rates` tab, then rebuilds the PDT Transactions tab. Failed scrapes are
retried up to 3 times; if all retries fail the last cached price is used with a
warning on stderr.

Override config values on the command line if needed:

```bash
get_rates -credentials /path/to/credentials.json
get_rates -cache /path/to/cache.json
get_rates -funds /path/to/funds.csv
```

### push_bnd

Copy your BND portfolio page (select all with `Cmd-A`, then `Cmd-C`), then run:

```bash
push_bnd
```

Or pipe input for testing:

```bash
cat sample.page | push_bnd
```

The tool validates that the clipboard contains a recognisable BND portfolio page,
normalises all numeric fields (European decimal format → dot notation), writes the
data to the `Import from BND` tab with a `Last update` timestamp, and then rebuilds
the PDT Transactions tab using the updated BND positions.

---

## Configuration

### config.ini

```ini
[paths]
funds     = funds.csv
breakdown = breakdown.csv
cache     = cache/rates.json
separator = ;
```

| Key | Default | Description |
|---|---|---|
| `funds` | `funds.csv` | Fund list for `get_rates` |
| `breakdown` | `breakdown.csv` | BND breakdown table |
| `cache` | `cache/rates.json` | Rate cache file |
| `separator` | `;` | CSV field separator |

### Command-line flags

| Flag | Default | Description |
|---|---|---|
| `-config` | `config.ini` | Path to config file |
| `-credentials` | `credentials.json` | Path to service account JSON |
| `-cache` | *(from config)* | Path to rate cache file (`get_rates` only) |
| `-funds` | *(from config)* | Path to funds CSV (`get_rates` only) |

### funds.csv

Semicolon-separated, one fund per row. Columns: `ISIN`, `Name`, `Source`,
`Asset Kind`, `Exchange`.

| Source code | Scraper |
|---|---|
| `TGD` | Tradegate |
| `FTD` | FT Markets |

### breakdown.csv

Maps BND pension fund ISINs to the underlying funds used in the PDT simulation.
Semicolon-separated, columns: `Source ISIN`, `Division`, `Target ISIN`.
`Division` is a decimal fraction (0–1) in European notation; fractions for a single
source ISIN must sum to 1. A warning is printed if they do not.

---

## Sheet structure

### Portfolio sheet

| Tab | Written by | Columns |
|---|---|---|
| `Import rates` | `get_rates` | A=ISIN, B=Name, C=Value, D=Last update |
| `Import from BND` | `push_bnd` | A=Fondsnaam, B=Aantal, C=Koers, D=Datum, E=Weging, F=Waarde |

The named range `Rates` (used by VLOOKUP in a third tab) should cover
`'Import rates'!A:D`.

### Portfolio Dividend Tracker template

| Tab | Written by | Notes |
|---|---|---|
| `Transactions` | both tools | Rows 1–3 are system headers; data starts at row 4 |
| `Bookings` | both tools | Row 4 holds one simulation booking; rows 1–3 are system headers |

**Transactions** — on every run, rows 4 and below are cleared and fully rewritten.
Direct fund positions are written as-is; BND pension positions are first expanded into
their underlying funds via `breakdown.csv`, with positions computed as:

```
target_amount = BND_net_value × division / target_fund_rate
```

where `BND_net_value = position × price × net_pct` (net_pct accounts for the BND
tax reservation, typically 0.81).

**Bookings** — on every run, row 4 is cleared and a single simulation booking is
written with the current total portfolio value. The value is read from the named range
`Total` in the Portfolio sheet (the sheet-computed total, not a recomputed sum).
Columns: Broker, Date, Time, Action (`Deposit`), Value, Currency (`EUR`).

---

## Rate cache

`get_rates` maintains a local cache at `cache/rates.json`. On a failed scrape the
cached price and its original fetch timestamp are used, and a warning is printed to
stderr. `push_bnd` reads the same cache to obtain current rates when rebuilding the
PDT Transactions tab.
