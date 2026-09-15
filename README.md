# invoice-line-parser

A parser and pretty printer for invoice line items in a plain pipe-delimited
text format. I keep getting invoice exports from small accounting tools that
are *almost* well-formed — a stray space, a percentage where a fraction was
expected, three decimal places on a price that should have two — and I
wanted something that would reject that stuff loudly by default instead of
quietly doing the wrong math.

## The format

One line item per line:

```
quantity | description | unit_price | tax_rate
```

`tax_rate` is optional and defaults to `0`. `tax_rate` is a fraction, not a
percentage (`0.0825`, not `8.25`). Lines that are blank or start with `#`
are comments.

Example file:

```
2 | Widescreen Monitor 27in | 249.99 | 0.0825
1 | Setup and configuration | 120.00
5 | USB-C Cable 2m | 8.50 | 0.0825
```

## Usage

```
go run . invoice.txt
```

```
QTY    DESCRIPTION                            UNIT       TAX      TOTAL
------------------------------------------------------------------------
2      Widescreen Monitor 27in               $249.99    8.25%    $541.23
1      Setup and configuration               $120.00    0.00%    $120.00
5      USB-C Cable 2m                          $8.50    8.25%     $46.00
------------------------------------------------------------------------
       GRAND TOTAL                                                $707.23
```

With no file argument it reads from stdin, so it composes with other tools:

```
cat invoice.txt | go run .
```

## Strict by default

By default the parser is strict about the things that usually mean the
upstream data is broken rather than merely differently styled:

- fields cannot have leading or trailing whitespace
- quantities cannot have leading zeros, and allow at most 3 decimal places
- unit prices must have exactly two decimal places (`12.30`, not `12.3`)
- tax rate must be a fraction in `[0, 1]`, not a percentage
- a trailing `| tax_rate` field cannot be present but empty

Any violation is reported with a line number and stops the run (exit code
1), with every bad line reported at once rather than stopping at the first.

Pass `--lenient` to relax all of the above: whitespace is trimmed, numbers
of any precision are accepted, and a missing or empty tax rate defaults to
`0`. Reach for it when you trust the source but not its formatting; leave it
off when you're validating a feed you don't control.

```
go run . --lenient sloppy_export.txt
```

## Status

Early. The parser, decimal handling, and table renderer work end to end,
but there's no test suite yet and the column widths in the pretty printer
are fixed rather than content-aware.

## License

MIT, see LICENSE.
