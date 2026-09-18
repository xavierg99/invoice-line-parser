package main

import (
	"strings"
	"testing"
)

// wantResult describes the expected outcome of parsing a line under one
// mode (strict or lenient). An empty errContains means the parse must
// succeed and match item exactly; a non-empty one means it must fail
// with an error containing that substring.
type wantResult struct {
	errContains string
	item        LineItem
}

func checkParseLine(t *testing.T, line string, opts Options, want wantResult) {
	t.Helper()
	got, err := parseLine(line, 1, opts)
	if want.errContains != "" {
		if err == nil {
			t.Fatalf("parseLine(%q, lenient=%v) = %+v, want error containing %q", line, opts.Lenient, got, want.errContains)
		}
		if !strings.Contains(err.Error(), want.errContains) {
			t.Fatalf("parseLine(%q, lenient=%v) error = %q, want it to contain %q", line, opts.Lenient, err.Error(), want.errContains)
		}
		return
	}
	if err != nil {
		t.Fatalf("parseLine(%q, lenient=%v) unexpected error: %v", line, opts.Lenient, err)
	}
	if got != want.item {
		t.Fatalf("parseLine(%q, lenient=%v) = %+v, want %+v", line, opts.Lenient, got, want.item)
	}
}

func TestParseLine(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		strict  wantResult
		lenient wantResult
	}{
		{
			name: "well formed three fields",
			line: "2|Widget|10.00",
			strict: wantResult{
				item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0, Line: 1},
			},
			lenient: wantResult{
				item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0, Line: 1},
			},
		},
		{
			name: "well formed four fields with tax",
			line: "2|Widget|10.00|0.0825",
			strict: wantResult{
				item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0.0825, Line: 1},
			},
			lenient: wantResult{
				item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0.0825, Line: 1},
			},
		},
		{
			name:    "surrounding whitespace on fields",
			line:    "2 | Widget | 10.00",
			strict:  wantResult{errContains: "whitespace"},
			lenient: wantResult{item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0, Line: 1}},
		},
		{
			name:    "leading zero on quantity",
			line:    "02|Widget|10.00",
			strict:  wantResult{errContains: "leading zeros"},
			lenient: wantResult{item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0, Line: 1}},
		},
		{
			name: "quantity with three decimals is fine in both modes",
			line: "2.125|Widget|10.00",
			strict: wantResult{
				item: LineItem{Quantity: 2.125, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0, Line: 1},
			},
			lenient: wantResult{
				item: LineItem{Quantity: 2.125, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0, Line: 1},
			},
		},
		{
			name:    "quantity with four decimals rejected strict, allowed lenient",
			line:    "2.1234|Widget|10.00",
			strict:  wantResult{errContains: "decimal places"},
			lenient: wantResult{item: LineItem{Quantity: 2.1234, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0, Line: 1}},
		},
		{
			name:    "zero quantity rejected in both modes",
			line:    "0|Widget|10.00",
			strict:  wantResult{errContains: "greater than zero"},
			lenient: wantResult{errContains: "greater than zero"},
		},
		{
			name:    "negative quantity rejected in both modes",
			line:    "-1|Widget|10.00",
			strict:  wantResult{errContains: "quantity"},
			lenient: wantResult{errContains: "quantity"},
		},
		{
			name:    "price with one decimal place rejected strict, allowed lenient",
			line:    "2|Widget|10.5",
			strict:  wantResult{errContains: "two decimal places"},
			lenient: wantResult{item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1050, TaxRate: 0, Line: 1}},
		},
		{
			name:    "negative price rejected in both modes",
			line:    "2|Widget|-5.00",
			strict:  wantResult{errContains: "unit price"},
			lenient: wantResult{errContains: "unit price"},
		},
		{
			name:    "empty description rejected in both modes",
			line:    "2||10.00",
			strict:  wantResult{errContains: "description is empty"},
			lenient: wantResult{errContains: "description is empty"},
		},
		{
			name:    "too few fields rejected in both modes",
			line:    "2|Widget",
			strict:  wantResult{errContains: "expected 3 or 4 fields"},
			lenient: wantResult{errContains: "expected 3 or 4 fields"},
		},
		{
			name:    "too many fields rejected in both modes",
			line:    "2|Widget|10.00|0.05|extra",
			strict:  wantResult{errContains: "expected 3 or 4 fields"},
			lenient: wantResult{errContains: "expected 3 or 4 fields"},
		},
		{
			name:    "tax rate given as a percentage rejected strict, allowed lenient",
			line:    "2|Widget|10.00|8.25",
			strict:  wantResult{errContains: "fraction between 0 and 1"},
			lenient: wantResult{item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 8.25, Line: 1}},
		},
		{
			name: "tax rate of exactly one is a valid fraction",
			line: "2|Widget|10.00|1.0",
			strict: wantResult{
				item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 1.0, Line: 1},
			},
			lenient: wantResult{
				item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 1.0, Line: 1},
			},
		},
		{
			name:    "present but empty tax field rejected strict, defaults to zero lenient",
			line:    "2|Widget|10.00|",
			strict:  wantResult{errContains: "present but empty"},
			lenient: wantResult{item: LineItem{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0, Line: 1}},
		},
		{
			name:    "negative tax rate rejected in both modes",
			line:    "2|Widget|10.00|-0.5",
			strict:  wantResult{errContains: "tax rate"},
			lenient: wantResult{errContains: "tax rate"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/strict", func(t *testing.T) {
			checkParseLine(t, tc.line, Options{Lenient: false}, tc.strict)
		})
		t.Run(tc.name+"/lenient", func(t *testing.T) {
			checkParseLine(t, tc.line, Options{Lenient: true}, tc.lenient)
		})
	}
}

func TestParseDocumentSkipsCommentsAndBlankLines(t *testing.T) {
	input := "" +
		"# this is a comment\n" +
		"\n" +
		"2|Widget|10.00\n" +
		"# another comment\n" +
		"3|Gadget|5.00\n"

	items, errs := ParseDocument(strings.NewReader(input), Options{})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2: %+v", len(items), items)
	}
	if items[0].Description != "Widget" || items[1].Description != "Gadget" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestParseDocumentCollectsAllErrors(t *testing.T) {
	input := "" +
		"2|Widget|10.00\n" +
		"bad|Broken\n" +
		"0|Zero Qty|5.00\n" +
		"3|Gadget|5.00\n"

	items, errs := ParseDocument(strings.NewReader(input), Options{})
	if len(items) != 2 {
		t.Fatalf("got %d valid items, want 2: %+v", len(items), items)
	}
	if len(errs) != 2 {
		t.Fatalf("got %d errors, want 2: %v", len(errs), errs)
	}

	pe, ok := errs[0].(*ParseError)
	if !ok {
		t.Fatalf("errs[0] is %T, want *ParseError", errs[0])
	}
	if pe.Line != 2 {
		t.Fatalf("errs[0].Line = %d, want 2", pe.Line)
	}

	pe, ok = errs[1].(*ParseError)
	if !ok {
		t.Fatalf("errs[1] is %T, want *ParseError", errs[1])
	}
	if pe.Line != 3 {
		t.Fatalf("errs[1].Line = %d, want 3", pe.Line)
	}
}

func TestParseDocumentLenientTrimsWhitespaceOnlyLines(t *testing.T) {
	input := "2|Widget|10.00\n   \n3|Gadget|5.00\n"

	items, errs := ParseDocument(strings.NewReader(input), Options{Lenient: true})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2: %+v", len(items), items)
	}
}
