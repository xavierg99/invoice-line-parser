package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files with current output")

// checkGolden compares got against testdata/<name>.golden. Run the tests
// with -update after an intentional layout change and review the diff.
func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("output differs from %s\n--- got ---\n%s--- want ---\n%s", path, got, want)
	}
}

func TestFormatTableGolden(t *testing.T) {
	items := []LineItem{
		{Quantity: 2, Description: "Widget", UnitPriceCents: 1000, TaxRate: 0.0825, Line: 1},
		{Quantity: 2.5, Description: "Consulting hours", UnitPriceCents: 12000, TaxRate: 0, Line: 2},
		{Quantity: 1, Description: "A very long description that exceeds the column width", UnitPriceCents: 123456, TaxRate: 0.1, Line: 3},
	}
	checkGolden(t, "basic", FormatTable(items))
}

func TestFormatTableEmpty(t *testing.T) {
	if got, want := FormatTable(nil), "(no line items)\n"; got != want {
		t.Fatalf("FormatTable(nil) = %q, want %q", got, want)
	}
}

func TestFormatCents(t *testing.T) {
	cases := map[int64]string{
		0:       "$0.00",
		5:       "$0.05",
		100:     "$1.00",
		123456:  "$1234.56",
		-5:      "-$0.05",
		-123456: "-$1234.56",
	}
	for in, want := range cases {
		if got := formatCents(in); got != want {
			t.Errorf("formatCents(%d) = %q, want %q", in, got, want)
		}
	}
}
