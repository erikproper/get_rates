/*
 *
 * Module:    GetRates
 * Package:   Sheets
 * Component: Client
 *
 * Authenticates with the Google Sheets API and writes rate data to the target spreadsheet.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package sheets

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	spreadsheetID = "10usrsYZ47thNRCMTO9pHpIlqfR8nJlH8A37h1GGAS1o"

	ratesWriteRange = "'Import rates'!A:D"
	ratesSheetTitle = "Import rates"
	ratesNumCols    = int64(4)

	bndWriteRange = "'Import from BND'!A:F"
	bndSheetTitle = "Import from BND"
	bndNumCols    = int64(6)
)

// TClient wraps the Google Sheets service for writing rate data.
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

// --- write ---

// WriteRates overwrites the Import rates tab with the given rows, then left-aligns
// all cells and italicises the header row.
func (c *TClient) WriteRates(rows [][]interface{}) error {
	if err := c.writeValues(ratesWriteRange, rows); err != nil {
		return err
	}
	sheetID, err := c.sheetID(ratesSheetTitle)
	if err != nil {
		return fmt.Errorf("looking up sheet: %w", err)
	}
	return c.applyFormatting(sheetID, ratesNumCols, true)
}

// WriteBND clears and overwrites the Import from BND tab with the given rows,
// then left-aligns all cells.
func (c *TClient) WriteBND(rows [][]interface{}) error {
	if err := c.clearRange(bndWriteRange); err != nil {
		return err
	}
	if err := c.writeValues(bndWriteRange, rows); err != nil {
		return err
	}
	sheetID, err := c.sheetID(bndSheetTitle)
	if err != nil {
		return fmt.Errorf("looking up sheet: %w", err)
	}
	return c.applyFormatting(sheetID, bndNumCols, false)
}

// --- helpers ---

func (c *TClient) writeValues(rangeStr string, rows [][]interface{}) error {
	_, err := c.service.Spreadsheets.Values.
		Update(spreadsheetID, rangeStr, &sheets.ValueRange{Values: rows}).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		return fmt.Errorf("writing to sheet: %w", err)
	}
	return nil
}

func (c *TClient) clearRange(rangeStr string) error {
	_, err := c.service.Spreadsheets.Values.
		Clear(spreadsheetID, rangeStr, &sheets.ClearValuesRequest{}).
		Do()
	if err != nil {
		return fmt.Errorf("clearing range: %w", err)
	}
	return nil
}

func (c *TClient) sheetID(title string) (int64, error) {
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
func (c *TClient) applyFormatting(sheetID, numCols int64, italicHeader bool) error {
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
