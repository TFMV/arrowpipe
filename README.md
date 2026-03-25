# ArrowPipe 🏹

**A powerful command-line toolkit for high-performance data processing and analysis, built on Apache Arrow.**

---

[![Go version](https://img.shields.io/github/go-mod/go-version/TFMV/arrowpipe)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/TFMV/arrowpipe)](https://goreportcard.com/report/github.com/TFMV/arrowpipe)
[![Build Status](https://github.com/TFMV/arrowpipe/actions/workflows/go.yml/badge.svg)](https://github.com/TFMV/arrowpipe/actions)
[![License](https://img.shields.io/github/license/TFMV/arrowpipe)](./LICENSE)

`arrowpipe` is a CLI for data artisans. It brings the power of in-memory, columnar data processing to your terminal, allowing you to build complex, high-performance data pipelines with simple, chainable commands. It uses Apache Arrow as its core engine for speed and efficiency.

Think of it as `sed`, `awk`, and `jq` for structured, tabular data, but supercharged.

## Core Concepts

- **Unix Philosophy**: `arrowpipe` reads from `stdin` and writes to `stdout`. This allows you to pipe commands together to create sophisticated data workflows right in your shell.
- **Apache Arrow**: All data flowing between `arrowpipe` commands is in the Arrow IPC format. This eliminates the overhead of parsing and serialization at each step, making your pipelines incredibly fast.
- **Rich Command Set**: Go beyond simple conversion. Filter, select, aggregate, inspect, and transform your data with a comprehensive set of commands.

## 🚀 Installation

Ensure you have Go installed (version 1.21 or newer). Then, install `arrowpipe` with:

```sh
go install github.com/TFMV/arrowpipe/cmd/arrowpipe@latest
```

Verify the installation:
```sh
arrowpipe --version
```

## チュートリアル： Quick Start

Let's see `arrowpipe` in action. Imagine you have a CSV file of sales data, `sales.csv`:

**`sales.csv`**
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
  arrowpipe aggregate "sum(sales)" | \
  arrowpipe to-json
```

**Output:**
```json
{"sales_sum":"150.75"}
```

This pipeline:
1. Converts the CSV to Arrow format (`from-csv`).
2. Filters the data to keep only the relevant rows (`filter`).
3. Calculates the sum of the `sales` column for the filtered data (`aggregate`).
4. Formats the final result as JSON (`to-json`).

## ⚙️ Command Reference

`arrowpipe` provides a rich set of commands for a variety of data tasks.

### Data Conversion

| Command | Description | Example |
|---|---|---|
| `from-csv` | Converts CSV data to Arrow. | `cat data.csv \| arrowpipe from-csv` |
| `to-csv` | Converts Arrow data to CSV. | `cat data.arrow \| arrowpipe to-csv` |
| `from-json` | Converts JSON to Arrow. | `cat data.json \| arrowpipe from-json` |
| `to-json` | Converts Arrow to JSON. | `cat data.arrow \| arrowpipe to-json` |

### Data Transformation & Analysis

| Command | Description | Example |
|---|---|---|
| `filter` | Filters rows based on an expression. | `... \| arrowpipe filter "score > 80"` |
| `select` | Selects a subset of columns. | `... \| arrowpipe select "id,name,score"` |
| `aggregate` | Performs aggregations (sum, mean, etc.). | `... \| arrowpipe aggregate "avg(score),max(age)"` |
| `columns`| Rename, cast, or drop columns. | `... \| arrowpipe columns "rename(a,b),drop(c)"`|
| `stats`| Computes summary statistics. | `... \| arrowpipe stats` |

### Data Introspection

| Command | Description | Example |
|---|---|---|
| `inspect` | Shows a human-readable preview of the data. | `... \| arrowpipe inspect` |
| `schema` | Displays the Arrow schema of the data stream. | `... \| arrowpipe schema` |


### Other Utilities
| Command | Description | Example |
|---|---|---|
| `init`| Initializes a new `arrowpipe` project (e.g., config file). | `arrowpipe init my-project` |
|`create-dummy-data`| Generates sample datasets for testing.| `arrowpipe create-dummy-data --rows 100 > dummy.arrow`|
|`benchmark`| Runs performance benchmarks.|`arrowpipe benchmark --runs 10 'from-csv'`|

## 🛠️ Development

This project is built with Go and Cobra.

- **To run tests:** `go test ./...`
- **To build the binary:** `go build ./cmd/arrowpipe`

## 🤝 Contributing

Contributions are welcome! Please open an issue or submit a pull request to help make `arrowpipe` even better.
