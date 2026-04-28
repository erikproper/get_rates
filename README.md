# get_rates

Two small Go tools that replace a manual clipboard-based portfolio workflow:

- **`get_rates`** — scrapes current prices for a fixed list of funds/ETFs from
  [Tradegate](https://www.tradegate.de) and [FT Markets](https://markets.ft.com),
  then writes the results directly to a Google Sheet.
- **`push_bnd`** — reads a clipboard copy of the BND pension-account portfolio page,
  normalises the data, and writes it to a second tab in the same sheet.

---

## Prerequisites

- Go 1.22+
- A Google Cloud project with the **Google Sheets API** enabled
- A **service account** with Editor access to the target spreadsheet
- The service account's JSON credentials file saved as `credentials.json`

---

## Installation

```bash
go install ./cmd/get_rates/
go install ./cmd/push_bnd/
```

Binaries are installed to `$GOPATH/bin` (ensure that is on your `PATH`).

---

## Usage

### get_rates

Run from the directory containing `credentials.json`:

```bash
get_rates
```

Or point to credentials explicitly:

```bash
get_rates -credentials /path/to/credentials.json
```

Scrapes all 41 funds (25 from Tradegate, 16 from FT Markets), prints a progress line
per fund, and writes the results to the `Import rates` tab. Failed scrapes are retried
up to 3 times; if all retries fail the last cached price is used with a warning on
stderr.

### push_bnd

Copy your BND portfolio page (the page listing fund holdings, prices and values) with
`Cmd-A` / `Cmd-C`, then run:

```bash
push_bnd
```

Or pipe input manually for testing:

```bash
cat sample.page | push_bnd
```

The tool validates that the clipboard contains a recognisable BND portfolio page,
normalises all numeric fields (European decimal format → dot notation), and writes the
data to the `Import from BND` tab with a `Last update` timestamp appended.

---

## Sheet structure

| Tab | Written by | Columns |
|---|---|---|
| `Import rates` | `get_rates` | A=ISIN, B=Name, C=Value, D=Last update |
| `Import from BND` | `push_bnd` | A=Fondsnaam, B=Aantal, C=Koers, D=Datum, E=Weging, F=Waarde |

The named range `Rates` (referenced by VLOOKUP in a third tab) should cover
`'Import rates'!A:D`.

---

## Rate cache

`get_rates` maintains a local cache at `cache/rates.json` (relative to the working
directory, configurable with `-cache`). On a failed scrape the cached price and its
original fetch timestamp are used, and a warning is printed to stderr.

---

## Configuration

| Flag | Default | Description |
|---|---|---|
| `-credentials` | `credentials.json` | Path to service account JSON |
| `-cache` | `cache/rates.json` | Path to rate cache file (`get_rates` only) |
