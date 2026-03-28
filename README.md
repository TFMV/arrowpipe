# ArrowPipe

**A high-performance CLI toolkit for data processing and analysis, built on Apache Arrow.**

---

[![Go version](https://img.shields.io/github/go-mod/go-version/TFMV/arrowpipe)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/TFMV/arrowpipe)](https://goreportcard.com/report/github.com/TFMV/arrowpipe)
[![Build Status](https://github.com/TFMV/arrowpipe/actions/workflows/go.yml/badge.svg)](https://github.com/TFMV/arrowpipe/actions)
[![License](https://img.shields.io/github/license/TFMV/arrowpipe)](./LICENSE)

`arrowpipe` brings the power of in-memory, columnar data processing to your terminal. It allows you to build complex, high-performance data pipelines with simple, chainable commands using Apache Arrow as its core engine.

Think of it as `sed`, `awk`, and `jq` for structured, tabular data, but supercharged.

## Core Concepts

- **Unix Philosophy**: `arrowpipe` reads from `stdin` and writes to `stdout`. This allows you to pipe commands together to create sophisticated data workflows right in your shell.
- **Apache Arrow**: All data flowing between `arrowpipe` commands is in the Arrow IPC format, eliminating the overhead of parsing and serialization at each step.
- **Rich Command Set**: Filter, select, aggregate, transform, and analyze your data with a comprehensive set of commands.

## Installation

Ensure you have Go installed (version 1.21 or newer). Then, install `arrowpipe` with:

```sh
go install github.com/TFMV/arrowpipe/cmd/arrowpipe@latest
```

Verify the installation:
```sh
arrowpipe --version
```

## Quick Start

Imagine you have a CSV file of sales data, `sales.csv`:

```csv
region,product,sales,quantity
north,widget,100.50,10
south,gadget,250.00,5
north,gadget,150.75,15
west,widget,50.25,8
south,widget,120.00,12
```

**Goal:** Find the total sales for the "widget" product in the "north" region.

You can do this in a single, readable pipeline:

```sh
cat sales.csv | \
  arrowpipe from-csv | \
  arrowpipe filter "region == 'north' && product == 'widget'" | \
  arrowpipe aggregate --metrics "sum(sales)" | \
  arrowpipe to-json
```

**Output:**
```json
{"sum_sales":"100.5"}
```

This pipeline:
1. Converts the CSV to Arrow format (`from-csv`).
2. Filters the data to keep only the relevant rows (`filter`).
3. Calculates the sum of the `sales` column (`aggregate`).
4. Formats the final result as JSON (`to-json`).

### Alternative: Single Pipeline Command

You can also combine multiple transformations in a single command:

```sh
cat sales.csv | \
  arrowpipe from-csv | \
  arrowpipe pipeline --filter "region == 'north' && product == 'widget'" --aggregate "sum(sales)" | \
  arrowpipe to-json
```

## Command Reference

### Data Conversion

| Command | Description | Example |
|---------|-------------|---------|
| `from-csv` | Converts CSV from stdin to Arrow IPC on stdout. | `cat data.csv \| arrowpipe from-csv` |
| `to-csv` | Converts Arrow IPC from stdin to CSV on stdout. | `cat data.arrow \| arrowpipe to-csv` |
| `from-json` | Converts JSON array from stdin to Arrow IPC on stdout. | `cat data.json \| arrowpipe from-json` |
| `to-json` | Converts Arrow IPC from stdin to JSON on stdout. | `cat data.arrow \| arrowpipe to-json` |

### Data Transformation & Analysis

| Command | Description | Example |
|---------|-------------|---------|
| `filter` | Filters rows based on an expression. Supports `&&` (AND) and `\|\|` (OR). | `... \| arrowpipe filter --expr "score > 80"` |
| `select` | Selects a subset of columns. | `... \| arrowpipe select --columns "id,name,score"` |
| `aggregate` | Performs aggregations (sum, avg, mean, min, max, count). | `... \| arrowpipe aggregate --metrics "avg(score),max(age)"` |
| `columns` | Rename, cast, or drop columns. Use `--list` to show columns. | `... \| arrowpipe columns --ops "rename(a,b),drop(c),cast(d,int64)"` |
| `stats` | Shows schema and row count. | `... \| arrowpipe stats` |
| `pipeline` | Chain multiple transformations in one command. | `... \| arrowpipe pipeline --filter "x > 5" --select "a,b" --aggregate "sum(c)"` |

### Data Introspection

| Command | Description | Example |
|---------|-------------|---------|
| `inspect` | Shows chunk information. | `... \| arrowpipe inspect` |
| `schema` | Displays the Arrow schema. | `... \| arrowpipe schema` |

### Other Utilities

| Command | Description | Example |
|---------|-------------|---------|
| `init` | Initializes a new `arrowpipe` project. | `arrowpipe init my-project` |
| `create-dummy-data` | Generates sample datasets. | `arrowpipe create-dummy-data --rows 100 --out data.arrow` |
| `benchmark` | Runs performance benchmarks. | `arrowpipe benchmark --runs 10 'from-csv'` |

## Filter Expressions

The `filter` command supports:
- **Comparison operators**: `==`, `!=`, `>`, `<`, `>=`, `<=`
- **Logical operators**: `&&` (AND), `||` (OR)
- **String literals**: Use single or double quotes: `region == 'north'`
- **Numeric literals**: Use without quotes: `age > 21`

Examples:
```sh
# Simple filter
cat data.csv | arrowpipe from-csv | arrowpipe filter --expr "age > 21"

# Multiple conditions with AND
cat data.csv | arrowpipe from-csv | arrowpipe filter --expr "age > 21 && status == 'active'"

# Multiple conditions with OR
cat data.csv | arrowpipe from-csv | arrowpipe filter --expr "status == 'active' || status == 'pending'"
```

## Column Operations

The `columns` command supports three operations:
- **rename**: `rename(oldName,newName)`
- **drop**: `drop(columnName)`
- **cast**: `cast(columnName,newType)` - Types: `int64`, `float64`, `string`, `bool`

Multiple operations can be combined with commas:

```sh
cat data.arrow | arrowpipe columns --ops "rename(region,area),drop(id),cast(age,int64)"
```

List columns:
```sh
cat data.arrow | arrowpipe columns --list
```

## Development

This project is built with Go and Cobra.

- **To run tests:** `go test ./...`
- **To build the binary:** `go build ./cmd/arrowpipe`

## Contributing

Contributions are welcome! Please open an issue or submit a pull request to help make `arrowpipe` even better.

## License

MIT