/*
 *
 * Module:    GetRates
 * Package:   Model
 * Component: Position
 *
 * Defines the portfolio position type read from the Portfolio tab.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package model

// TPosition holds a single fund position as read from the Portfolio tab.
type TPosition struct {
	ISIN     string
	Broker   string
	Name     string
	Position float64
	Price    float64
	NetPct   float64 // decimal fraction, e.g. 0.81 or 1.0
}

// NetValue returns the after-tax net value of the position.
func (p TPosition) NetValue() float64 {
	return p.Position * p.Price * p.NetPct
}
