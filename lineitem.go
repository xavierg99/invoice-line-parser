package main

import "math"

// LineItem is a single priced entry on an invoice. Money is kept in
// integer cents throughout so totals don't drift from float rounding;
// quantity is the one field where fractional values (2.5 kg, 0.5 hr)
// are the normal case, so it stays a float64.
type LineItem struct {
	Quantity       float64
	Description    string
	UnitPriceCents int64
	TaxRate        float64 // fraction, e.g. 0.0825 for 8.25%
	Line           int     // source line number, for traceability
}

func (li LineItem) SubtotalCents() int64 {
	return int64(math.Round(li.Quantity * float64(li.UnitPriceCents)))
}

func (li LineItem) TaxCents() int64 {
	return int64(math.Round(float64(li.SubtotalCents()) * li.TaxRate))
}

func (li LineItem) TotalCents() int64 {
	return li.SubtotalCents() + li.TaxCents()
}
