# FV-SEC001 - Software Engineer Challenge — Ad Performance Aggregator

## Solution Overview

This repository contains a Go CLI that aggregates totals per `campaign_id` and produces two result files:

- `top10_ctr.csv`
- `top10_cpa.csv`

### Methods Implemented

1. **Fast (default)**
    - Single-threaded, byte-level parser that streams the file line-by-line.
    - Parses `spend` directly into cents to avoid float overhead.

2. **Legacy**
    - Uses `bytes.SplitN` and `strconv.ParseFloat` for simplicity.
    - Higher allocation pressure and slower runtime.

3. **Parallel**
    - Reader goroutine scans bytes and parses rows.
    - Worker pool aggregates into local maps, then merges totals.
    - Lower contention, but higher coordination overhead.

### Selected Approach

The CLI uses **Fast** by default because it delivers the best overall time and memory footprint on the 1GB dataset while keeping the implementation stable and easy to reason about.

### Benchmark Environment

- OS: Linux
- Architecture: amd64
- CPU: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
- RAM: 8GB
- Dataset: full `ad_data.csv` (~1GB)

### Benchmark Command

```bash
AGG_BENCH_FULL=1 go test ./internal/aggregator \
  -run '^$' \
  -bench 'BenchmarkAggregateFile(Fast|Legacy|Parallel)$' \
  -benchmem \
  -benchtime=10x
```

### Benchmark Results

| Metric      | Fast    | Legacy  | Parallel |
| ----------- | ------- | ------- | -------- |
| Time        | 3.14s   | 7.39s   | 12.34s   |
| CPU time    | 3359 ms | 8149 ms | 25026 ms |
| Peak heap   | 8.66 MB | 9.45 MB | 9.44 MB  |
| Allocated   | 219 MB  | 5.37 GB | 219 MB   |
| Alloc count | 26.8M   | 80.5M   | 26.8M    |

Relative comparison (Fast = 1.0x):

| Metric          |  Fast | Legacy | Parallel |
| --------------- | ----: | -----: | -------: |
| Runtime         | 1.00x |  2.35x |    3.93x |
| CPU time        | 1.00x |  2.43x |    7.45x |
| Allocated bytes | 1.00x | 24.54x |    1.00x |
| Alloc count     | 1.00x |  3.00x |    1.00x |

### Implementation Details

- Input is read via a large buffered reader to reduce syscalls.
- Each row is parsed without allocating intermediate slices.
- Aggregation uses a map keyed by `campaign_id` and sums totals in integer space.
- Top-k selection uses heaps to avoid sorting all campaigns when only 10 are needed.
- Malformed rows are skipped and counted.

## Setup

- Go 1.21+
- Standard library only (no external dependencies)

## How to Run

Run tests to verify correctness:

```bash
go test ./...
```

Run the CLI on the full dataset:

```bash
go run . --input ad_data.csv --output results
```

Run full benchmarks with all implementations:

```bash
AGG_BENCH_FULL=1 go test ./internal/aggregator -run '^$' -bench 'BenchmarkAggregateFile(Fast|Legacy|Parallel)$' -benchmem -benchtime=2x
```

## Output Format

Each output file contains these columns:

- `campaign_id`
- `total_impressions`
- `total_clicks`
- `total_spend`
- `total_conversions`
- `CTR`
- `CPA`

`CTR` is written with 4 decimal places. `CPA` is written with 2 decimal places and is left blank for campaigns with zero conversions in the CTR ranking file. The CPA ranking file excludes campaigns with zero conversions.
