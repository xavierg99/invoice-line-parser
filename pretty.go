package main

import (
	"fmt"
	"strconv"
	"strings"
)

const descColWidth = 32

// FormatTable renders parsed line items as an aligned, fixed-width
// table with a grand total row, suitable for printing to a terminal.
func FormatTable(items []LineItem) string {
	if len(items) == 0 {
		return "(no line items)\n"
	}

	var b strings.Builder
	writeRow(&b, "QTY", "DESCRIPTION", "UNIT", "TAX", "TOTAL")
	rule := strings.Repeat("-", 6+1+descColWidth+1+10+1+8+1+10)
	b.WriteString(rule + "\n")

	var grandTotal int64
	for _, li := range items {
		writeRow(&b,
			formatQty(li.Quantity),
			truncate(li.Description, descColWidth),
			formatCents(li.UnitPriceCents),
			fmt.Sprintf("%.2f%%", li.TaxRate*100),
			formatCents(li.TotalCents()),
		)
		grandTotal += li.TotalCents()
	}

	b.WriteString(rule + "\n")
	writeRow(&b, "", "GRAND TOTAL", "", "", formatCents(grandTotal))
	return b.String()
}

func writeRow(b *strings.Builder, qty, desc, unit, tax, total string) {
	fmt.Fprintf(b, "%-6s %-*s %10s %8s %10s\n", qty, descColWidth, desc, unit, tax, total)
}

func truncate(s string, width int) string {
	if len(s) <= width {
		return s
	}
	return s[:width-1] + "…"
}

func formatCents(c int64) string {
	neg := c < 0
	if neg {
		c = -c
	}
	s := fmt.Sprintf("$%d.%02d", c/100, c%100)
	if neg {
		s = "-" + s
	}
	return s
}

func formatQty(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
