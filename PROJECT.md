# Fund Rate Scraper — Claude Code Project Brief

## Status

Completed (29-04-2026). Both `get_rates` and `push_bnd` are built, tested, and
installed. The original Bash workflow has been fully replaced.

Extended (29-04-2026): both tools now also write a simulation booking to the PDT
Bookings tab, recording the current total portfolio value (read from the Portfolio
sheet's named range `Total`) as a single `Deposit` row at row 4.

---

## Goal

Rewrite an existing Bash scraping script in Go. The programme scrapes current prices for a fixed list of funds/ETFs from three websites and pushes the results directly into a Google Sheet, replacing a manual clipboard-based workflow.

---

## Current Workflow (to be replaced)

1. Bash script (`rates`) scrapes prices from three sources and copies results to macOS clipboard (comma as decimal separator)
2. A wrapper script flips commas to dots: `pbpaste | sed -e "s/,/./g" | pbcopy`
3. User manually pastes into a Google Sheet tab called `Import rates`
4. A second tab uses `VLOOKUP` referencing a named range called `Rates`

---

## Target Workflow

1. Go programme scrapes all funds
2. Normalises decimal separators (comma → dot), applied **per price field only** — not bluntly across the whole row
3. Pushes data directly to Google Sheets via the Sheets API
4. No clipboard, no manual paste

---

## Google Sheet Structure

- **Target tab:** `Import rates`
- **Named range:** `Rates`, currently defined as `'Import rates'!B2:D80` — to be redefined as `'Import rates'!B:D` (full columns, unlimited rows)
- **Row 1:** Header row — `ISIN | Name | Price`
- **Row 2 onward:** Data rows, fully overwritten on each run
- **Write strategy:** Single `Values.Update` call with `ValueInputOption("RAW")`, overwriting the full column range each run
- A second tab references `Rates` by name via `VLOOKUP` — no changes needed there

---

## Data Sources & Scraping

### 1. Tradegate (`TGD`)

- URL: `https://www.tradegate.de/orderbuch.php?lang=en&isin={ISIN}`
- Field: element with class `longprice` and `last`; fall back to `bid` if last price is unavailable
- Parser: `goquery` recommended

### 2. FT Markets (`FTD`)

- URL: `https://markets.ft.com/data/currencies/tearsheet/summary?s={ISIN}`
- Field: price label extracted from `Price (...)` pattern, then price value under that label
- Parser: `goquery` recommended

### 3. Morningstar (`MSD`)

- URL: `https://global.morningstar.com/nl/beleggingen/fondsen/{ISIN}/quote?marktID=nl`
- Field: element containing `Koers`, value in EUR
- Note: Morningstar may load data via JavaScript — verify that raw HTML fetch is sufficient; if not, consider using a headless browser or switching to an alternative source

---

## Fund List

### Tradegate (TGD)

| ISIN | Name |
|---|---|
| IE00B4ND3602 | iShares Physical Gold ETC |
| IE00B4NCWG09 | iShares Physical Silver ETC |
| NL0000337319 | Koninklijke BAM Groep NV |
| IE000U9ODG19 | iShares Global Aerospace & Defence ETF |
| IE00B4L5Y983 | iShares Core MSCI World UCITS ETF |
| DE0007030009 | Rheinmetall AG |
| BMG3602E1084 | Flow Traders Ltd |
| IE000YYE6WK5 | VanEck Defense UCITS ETF A |
| IE00B5L8K969 | iShares MSCI EM Asia UCITS ETF |
| IE00BK5BC891 | L&G Clean Water UCITS ETF |
| JE00BQRFDY49 | WisdomTree Core Physical Silver ETC |
| JE00BN2CJ301 | WisdomTree Core Physical Gold ETC |
| DE000A0H08F7 | iShares STOXX Europe 600 Construction & Materials ETF |
| IE0006WW1TQ4 | Xtrackers MSCI World ex USA UCITS ETF 1C |
| IE00BD1F4M44 | iShares Edge MSCI USA Value Factor UCITS ETF |
| IE00B5BMR087 | iShares Core S&P 500 UCITS ETF |
| IE000CFH1JX2 | iShares Global Water UCITS ETF |
| IE00BYTRRB94 | SPDR MSCI World Health Care UCITS ETF |
| IE00BM67HK77 | Xtrackers MSCI World Health Care UCITS ETF 1C |
| IE00BMH5Y327 | Global X Data Center REITS & Digital Infrastructure ETF |
| IE000CK5G8J7 | iShares Global Infrastructure UCITS ETF |
| IE00BYZK4552 | iShares Automation & Robotics UCITS ETF |
| IE00B53SZB19 | iShares Nasdaq 100 UCITS ETF |
| IE000E7EI9P0 | Amundi S&P Global Information Technology ESG UCITS ETF |
| IE00BF5DXP42 | First Trust Index Innovative Transaction & Process UCITS ETF |

### FT Markets (FTD)

| ISIN | Name |
|---|---|
| NL0012706154 | BND Wereld Indexfonds Hedged |
| NL0012706196 | BND Wereld Obligatie Indexfonds |
| NL0012706238 | BND Wereld Indexfonds Unhedged |
| NL0014040164 | BND Emerging Markets Indexfonds |
| NL0013546955 | BND Small Cap Wereld Indexfonds Carb Screened |
| NL0012706220 | BND Euro Investment Grade Obligatie Indexfonds |
| NL0012706204 | BND Euro Staatsobligatie Indexfonds |
| NL0012706279 | BND Euro Staatsobligatie Indexfonds Inflatie |
| IE00BVVQ9K69 | Selected Screened FTSE Developed World II (B) Common Contractual Fund - EUR Hedged Acc |
| IE000FZ5BIF1 | Vanguard Global Government Bond Index EUR Hedged Acc |
| IE00BNDS0V25 | Vanguard ESG Global Corporate Bond Index Fund EUR Hedged Acc |
| IE00BVVQBD33 | Vanguard Selected Screened FTSE Developed II (B) World Common Contractual Fd Instl B EUR Acc |
| NL0013089147 | Northern Trust UCITS FGR Fund - Emerging Markets Screened Equity Index FGR Fund E EUR |
| NL0013552094 | NT World Small Cap Low Carbon Equity Index FGR Feeder Fund Class C EUR Units |
| IE00BYSX5D68 | Vanguard Selected Screened Euro Investment Grade Bond Index Fund EUR Acc |
| IE0007472990 | Vanguard Euro Government Bond Index Fund EUR Acc |

---

## Suggested Project Structure

```
cmd/
  scrape/
    main.go           ← entry point; orchestrates scraping and sheet push
internal/
  scraper/
    tradegate.go      ← TGD scraper
    ft.go             ← FTD scraper
    morningstar.go    ← MSD scraper (verify JS rendering situation first)
  model/
    fund.go           ← Fund struct: ISIN, Name, Price string
  sheets/
    client.go         ← Google Sheets auth and write logic
config/
  funds.go            ← fund list as structured data (not hardcoded in scrapers)
```

---

## Google Sheets Authentication

- Create a **service account** in Google Cloud Console
- Download the JSON credentials file
- Share the target Google Sheet with the service account's email address (Editor access)
- Use `option.WithCredentialsFile("credentials.json")` in the Sheets client

### Dependencies

```
go get google.golang.org/api/sheets/v4
go get golang.org/x/oauth2/google
go get github.com/PuerkitoBio/goquery
```

---

## Rate Caching (Stale Fallback)

If a scrape fails for any fund, the programme must fall back to the last known good price rather than writing an empty or error value to the sheet.

### Behaviour

- On each successful scrape, persist the result to a local cache
- On failure, load the cached value, use it, and emit a warning to stderr
- If no cache exists for a fund and the scrape fails, write an empty price and warn
- The cache should record the timestamp of the last successful fetch, so warnings can indicate how stale the value is

### Cache Format

A simple JSON file (e.g. `cache/rates.json`) keyed by ISIN:

```json
{
  "IE00B4ND3602": { "price": "38.45", "fetched_at": "2026-04-07T09:00:00Z" },
  "NL0012706154": { "price": "112.30", "fetched_at": "2026-04-06T18:30:00Z" }
}
```

### Warning Format

```
WARN: IE00B4ND3602 (iShares Physical Gold ETC): scrape failed, using cached price 38.45 from 2026-04-07T09:00:00Z
WARN: NL0012706154 (BND Wereld Indexfonds Hedged): scrape failed, no cache available — price left empty
```

### Cache Location

Store in a `cache/` subdirectory relative to the binary, or make it configurable via a flag/environment variable.

---

## Development Conventions

When working on this project, read:
- `../DEVELOPMENT_CONVENTIONS.md` — overall conventions strategy
- `../GO_CONVENTIONS.md` — Go coding, naming, and parser conventions

---

## Key Implementation Notes

- Decimal normalisation must target **only the price field** — not applied across the full row string
- The write is a **full overwrite** of `'Import rates'!B:D` on every run — no appending, no partial updates
- The named range `Rates` should be redefined (manually, once) from `'Import rates'!B2:D80` to `'Import rates'!B:D` before first use
- Row 1 (`B1:D1`) should contain headers: `ISIN`, `Name`, `Price`
- Morningstar: verify whether `net/http` fetch returns price data or whether JavaScript rendering is required; if the latter, consider replacing MSD with an alternative source or a different fetch strategy
- The original Bash script fell back from `last` to `bid` price on Tradegate — preserve this fallback logic

