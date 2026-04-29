/*
 *
 * Module:    GetRates
 * Package:   Sheets
 * Component: Client
 *
 * Authenticates with the Google Sheets API and reads/writes data across the
 * Portfolio and PDT template spreadsheets.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 29.04.2026
 *
 */

package sheets

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"get_rates/internal/model"
)

const (
	portfolioSheetID = "10usrsYZ47thNRCMTO9pHpIlqfR8nJlH8A37h1GGAS1o"

	ratesWriteRange = "'Import rates'!A:D"
	ratesSheetTitle = "Import rates"
	ratesNumCols    = int64(4)

	bndWriteRange = "'Import from BND'!A:F"
	bndSheetTitle = "Import from BND"
	bndNumCols    = int64(6)

	portfolioReadRange = "'Portfolio'!A:P"

	pdtSheetID           = "1x7U-ieHotiuE6VcVBlDgw8SJC5m36XuNRHzpv4TczjE"
	pdtTransactionsRange = "'Transactions'!A4:K"
	pdtBookingsRange     = "'Bookings'!A4:F"
)

// TClient wraps the Google Sheets service for reading and writing spreadsheet data.
type TClient struct {
	service *sheets.Service
}

// --- constructor ---

func New(credentialsFile string) (*TClient, error) {
	ctx := context.Background()
	svc, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}
	return &TClient{service: svc}, nil
}

// --- portfolio read ---

// ReadPortfolio reads the Portfolio tab of the portfolio spreadsheet and returns
// all fund positions found in the fund table (identified by an ISIN in the Code column).
func (c *TClient) ReadPortfolio() ([]model.TPosition, error) {
	resp, err := c.service.Spreadsheets.Values.
		Get(portfolioSheetID, portfolioReadRange).
		ValueRenderOption("UNFORMATTED_VALUE").
		Do()
	if err != nil {
		return nil, fmt.Errorf("reading Portfolio tab: %w", err)
	}
	return parsePortfolioRows(resp.Values)
}

func parsePortfolioRows(rows [][]interface{}) ([]model.TPosition, error) {
	codeCol, brokerCol, nameCol, posCol, priceCol, netPctCol := -1, -1, -1, -1, -1, -1
	dataStart := -1

	for i, row := range rows {
		for j, cell := range row {
			switch fmt.Sprint(cell) {
			case "Code":
				codeCol = j
			case "Broker":
				brokerCol = j
			case "Name":
				nameCol = j
			case "Position":
				posCol = j
			case "Price":
				priceCol = j
			case "Net %":
				netPctCol = j
			}
		}
		if codeCol >= 0 && posCol >= 0 && priceCol >= 0 {
			dataStart = i + 1
			break
		}
	}
	if dataStart < 0 {
		return nil, fmt.Errorf("fund table header not found in Portfolio tab")
	}

	get := func(row []interface{}, col int) interface{} {
		if col >= 0 && col < len(row) {
			return row[col]
		}
		return nil
	}

	var positions []model.TPosition
	for _, row := range rows[dataStart:] {
		isin := fmt.Sprint(get(row, codeCol))
		if !isISIN(isin) {
			continue
		}
		pos, ok1 := cellFloat(get(row, posCol))
		price, ok2 := cellFloat(get(row, priceCol))
		if !ok1 || !ok2 || pos == 0 {
			continue
		}
		netPct := 1.0
		if n, ok := cellFloat(get(row, netPctCol)); ok && n > 0 {
			netPct = n
		}
		positions = append(positions, model.TPosition{
			ISIN:     isin,
			Broker:   fmt.Sprint(get(row, brokerCol)),
			Name:     fmt.Sprint(get(row, nameCol)),
			Position: pos,
			Price:    price,
			NetPct:   netPct,
		})
	}
	return positions, nil
}

// isISIN reports whether s looks like a valid ISIN: 2 uppercase letters + 10 alphanumerics.
func isISIN(s string) bool {
	if len(s) != 12 {
		return false
	}
	for i, r := range s {
		if i < 2 {
			if r < 'A' || r > 'Z' {
				return false
			}
		} else if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func cellFloat(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int64:
		return float64(x), true
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	}
	return 0, false
}

// --- rates write ---

// WriteRates overwrites the Import rates tab with the given rows, then left-aligns
// all cells and italicises the header row.
func (c *TClient) WriteRates(rows [][]interface{}) error {
	if err := c.writeValues(portfolioSheetID, ratesWriteRange, "RAW", rows); err != nil {
		return err
	}
	sheetID, err := c.sheetID(portfolioSheetID, ratesSheetTitle)
	if err != nil {
		return fmt.Errorf("looking up sheet: %w", err)
	}
	return c.applyFormatting(portfolioSheetID, sheetID, ratesNumCols, true)
}

// --- BND write ---

// WriteBND clears and overwrites the Import from BND tab with the given rows,
// then left-aligns all cells.
func (c *TClient) WriteBND(rows [][]interface{}) error {
	if err := c.clearRange(portfolioSheetID, bndWriteRange); err != nil {
		return err
	}
	if err := c.writeValues(portfolioSheetID, bndWriteRange, "RAW", rows); err != nil {
		return err
	}
	sheetID, err := c.sheetID(portfolioSheetID, bndSheetTitle)
	if err != nil {
		return fmt.Errorf("looking up sheet: %w", err)
	}
	return c.applyFormatting(portfolioSheetID, sheetID, bndNumCols, false)
}

// --- PDT write ---

// WriteTransactions clears row 4 onwards in the PDT Transactions tab and writes
// the given rows. Date strings are parsed by Sheets (USER_ENTERED) so that PDT
// recognises them as dates.
func (c *TClient) WriteTransactions(rows [][]interface{}) error {
	if err := c.clearRange(pdtSheetID, pdtTransactionsRange); err != nil {
		return err
	}
	return c.writeValues(pdtSheetID, pdtTransactionsRange, "USER_ENTERED", rows)
}

// ReadPortfolioTotal reads the named range "Total" from the portfolio spreadsheet
// and returns its value.
func (c *TClient) ReadPortfolioTotal() (float64, error) {
	resp, err := c.service.Spreadsheets.Values.
		Get(portfolioSheetID, "Total").
		ValueRenderOption("UNFORMATTED_VALUE").
		Do()
	if err != nil {
		return 0, fmt.Errorf("reading named range Total: %w", err)
	}
	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		return 0, fmt.Errorf("named range Total is empty")
	}
	f, ok := cellFloat(resp.Values[0][0])
	if !ok {
		return 0, fmt.Errorf("named range Total is not numeric: %v", resp.Values[0][0])
	}
	return f, nil
}

// WriteBooking clears the PDT Bookings tab from row 4 and writes a single booking row.
func (c *TClient) WriteBooking(row []interface{}) error {
	if err := c.clearRange(pdtSheetID, pdtBookingsRange); err != nil {
		return err
	}
	return c.writeValues(pdtSheetID, pdtBookingsRange, "USER_ENTERED", [][]interface{}{row})
}

// --- helpers ---

func (c *TClient) writeValues(spreadsheetID, rangeStr, valueOption string, rows [][]interface{}) error {
	_, err := c.service.Spreadsheets.Values.
		Update(spreadsheetID, rangeStr, &sheets.ValueRange{Values: rows}).
		ValueInputOption(valueOption).
		Do()
	if err != nil {
		return fmt.Errorf("writing to sheet: %w", err)
	}
	return nil
}

func (c *TClient) clearRange(spreadsheetID, rangeStr string) error {
	_, err := c.service.Spreadsheets.Values.
		Clear(spreadsheetID, rangeStr, &sheets.ClearValuesRequest{}).
		Do()
	if err != nil {
		return fmt.Errorf("clearing range: %w", err)
	}
	return nil
}

func (c *TClient) sheetID(spreadsheetID, title string) (int64, error) {
	ss, err := c.service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return 0, fmt.Errorf("getting spreadsheet: %w", err)
	}
	for _, s := range ss.Sheets {
		if s.Properties.Title == title {
			return s.Properties.SheetId, nil
		}
	}
	return 0, fmt.Errorf("sheet %q not found", title)
}

// applyFormatting left-aligns all cells in the first numCols columns, and optionally
// italicises the header row (row 1).
func (c *TClient) applyFormatting(spreadsheetID string, sheetID, numCols int64, italicHeader bool) error {
	requests := []*sheets.Request{
		{
			RepeatCell: &sheets.RepeatCellRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartColumnIndex: 0,
					EndColumnIndex:   numCols,
				},
				Cell: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{
						HorizontalAlignment: "LEFT",
					},
				},
				Fields: "userEnteredFormat.horizontalAlignment",
			},
		},
	}

	if italicHeader {
		requests = append(requests, &sheets.Request{
			RepeatCell: &sheets.RepeatCellRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    0,
					EndRowIndex:      1,
					StartColumnIndex: 0,
					EndColumnIndex:   numCols,
				},
				Cell: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{
						TextFormat: &sheets.TextFormat{Italic: true},
					},
				},
				Fields: "userEnteredFormat.textFormat.italic",
			},
		})
	}

	_, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return fmt.Errorf("formatting sheet: %w", err)
	}
	return nil
}
