package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	lenient := flag.Bool("lenient", false, "relax parsing: trim whitespace, default missing tax rate to 0, accept any decimal precision")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: %s [--lenient] [file]

Parses a pipe-delimited invoice line item file and prints a formatted
invoice table. Reads from stdin if no file is given.

Line format: quantity | description | unit_price | tax_rate
tax_rate is optional and defaults to 0.

`, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	var r io.Reader = os.Stdin
	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		r = f
	}

	items, errs := ParseDocument(r, Options{Lenient: *lenient})
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		if !*lenient {
			fmt.Fprintln(os.Stderr, "\nrerun with --lenient to relax formatting rules, or fix the input")
		}
		os.Exit(1)
	}

	fmt.Print(FormatTable(items))
}
