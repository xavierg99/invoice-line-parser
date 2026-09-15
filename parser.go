package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Options controls how permissive parsing is. The zero value is strict.
type Options struct {
	Lenient bool
}

// ParseError is one rejected line. ParseDocument keeps going after a
// ParseError so a caller sees every problem in a file at once, the way
// a linter does, instead of stopping at the first bad line.
type ParseError struct {
	Line int
	Msg  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}

var (
	// Strict mode rejects leading zeros, missing decimals, and stray
	// whitespace, on the theory that an invoice feed that's malformed
	// today is worth failing loudly rather than silently misparsing.
	strictQuantityRe = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]{1,3})?$`)
	strictPriceRe    = regexp.MustCompile(`^(0|[1-9][0-9]*)\.[0-9]{2}$`)
	strictTaxRe      = regexp.MustCompile(`^(0(\.[0-9]{1,4})?|1(\.0{1,4})?)$`)

	lenientNumberRe = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
)

// ParseDocument reads one line item per line in the form:
//
//	quantity | description | unit_price | tax_rate
//
// tax_rate is optional and defaults to 0. Blank lines and lines
// starting with '#' are treated as comments and skipped.
func ParseDocument(r io.Reader, opts Options) ([]LineItem, []error) {
	var items []LineItem
	var errs []error

	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if opts.Lenient {
			if trimmed := strings.TrimSpace(line); trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
		} else {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
		}

		item, err := parseLine(line, lineNo, opts)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		items = append(items, item)
	}
	if err := scanner.Err(); err != nil {
		errs = append(errs, fmt.Errorf("reading input: %w", err))
	}
	return items, errs
}

func parseLine(line string, lineNo int, opts Options) (LineItem, error) {
	fields := strings.Split(line, "|")
	if len(fields) < 3 || len(fields) > 4 {
		return LineItem{}, &ParseError{lineNo, fmt.Sprintf(
			"expected 3 or 4 fields separated by '|' (quantity, description, unit_price, [tax_rate]), got %d", len(fields))}
	}

	if !opts.Lenient {
		for i, f := range fields {
			if f != strings.TrimSpace(f) {
				return LineItem{}, &ParseError{lineNo, fmt.Sprintf(
					"field %d has leading or trailing whitespace (use --lenient to allow)", i+1)}
			}
		}
	} else {
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
	}

	qtyField, descField, priceField := fields[0], fields[1], fields[2]
	taxField := ""
	if len(fields) == 4 {
		taxField = fields[3]
	}

	if descField == "" {
		return LineItem{}, &ParseError{lineNo, "description is empty"}
	}

	qty, err := parseQuantity(qtyField, opts)
	if err != nil {
		return LineItem{}, &ParseError{lineNo, err.Error()}
	}

	priceCents, err := parsePriceCents(priceField, opts)
	if err != nil {
		return LineItem{}, &ParseError{lineNo, err.Error()}
	}

	taxRate := 0.0
	if taxField != "" {
		taxRate, err = parseTaxRate(taxField, opts)
		if err != nil {
			return LineItem{}, &ParseError{lineNo, err.Error()}
		}
	} else if !opts.Lenient && len(fields) == 4 {
		return LineItem{}, &ParseError{lineNo, "tax rate field is present but empty (use --lenient to default it to 0)"}
	}

	return LineItem{
		Quantity:       qty,
		Description:    descField,
		UnitPriceCents: priceCents,
		TaxRate:        taxRate,
		Line:           lineNo,
	}, nil
}

func parseQuantity(s string, opts Options) (float64, error) {
	if opts.Lenient {
		if !lenientNumberRe.MatchString(s) {
			return 0, fmt.Errorf("quantity %q is not a valid non-negative number", s)
		}
	} else if !strictQuantityRe.MatchString(s) {
		return 0, fmt.Errorf("quantity %q must be a plain number with no leading zeros and at most 3 decimal places", s)
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("quantity %q is not a valid number", s)
	}
	if v <= 0 {
		return 0, fmt.Errorf("quantity must be greater than zero, got %v", v)
	}
	return v, nil
}

func parsePriceCents(s string, opts Options) (int64, error) {
	if opts.Lenient {
		if !lenientNumberRe.MatchString(s) {
			return 0, fmt.Errorf("unit price %q is not a valid number", s)
		}
	} else if !strictPriceRe.MatchString(s) {
		return 0, fmt.Errorf("unit price %q must look like 12.34, exactly two decimal places, no leading zeros", s)
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("unit price %q is not a valid number", s)
	}
	if v < 0 {
		return 0, fmt.Errorf("unit price cannot be negative, got %v", v)
	}
	return int64(math.Round(v * 100)), nil
}

func parseTaxRate(s string, opts Options) (float64, error) {
	if opts.Lenient {
		if !lenientNumberRe.MatchString(s) {
			return 0, fmt.Errorf("tax rate %q is not a valid number", s)
		}
	} else if !strictTaxRe.MatchString(s) {
		return 0, fmt.Errorf("tax rate %q must be a fraction between 0 and 1, e.g. 0.0825 for 8.25%%", s)
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("tax rate %q is not a valid number", s)
	}
	if v < 0 {
		return 0, fmt.Errorf("tax rate cannot be negative, got %v", v)
	}
	if !opts.Lenient && v > 1 {
		return 0, fmt.Errorf("tax rate %q looks like a percentage; expected a fraction (0.08, not 8)", s)
	}
	return v, nil
}
