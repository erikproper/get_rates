/*
 *
 * Module:    GetRates
 * Package:   Model
 * Component: Fund
 *
 * Defines the fund and rate data types used throughout the programme.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package model

// TSource identifies which scraping source handles a fund.
type TSource int

const (
	SourceTradegate  TSource = iota
	SourceFTMarkets
	SourceMorningstar
)

// TFund describes a single fund to be scraped.
type TFund struct {
	ISIN   string
	Name   string
	Source TSource
}
