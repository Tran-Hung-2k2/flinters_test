package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"flinters_test/internal/aggregator"
)

func main() {
	// Parse CLI flags for input and output locations.
	inputPath := flag.String("input", "ad_data.csv", "Path to the input CSV file")
	outputDir := flag.String("output", "results", "Directory for generated CSV files")
	ctrFile := flag.String("ctr-file", "top10_ctr.csv", "Output CSV filename for top CTR")
	cpaFile := flag.String("cpa-file", "top10_cpa.csv", "Output CSV filename for top CPA")
	topN := flag.Int("top", 10, "Number of top campaigns to export")
	flag.Parse()

	// Aggregate campaign stats from the input file.
	stats, malformedRows, err := aggregator.AggregateFile(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to aggregate input: %v\n", err)
		os.Exit(1)
	}

	// Report any rows that failed validation.
	if malformedRows > 0 {
		fmt.Fprintf(os.Stderr, "skipped %d malformed rows\n", malformedRows)
	}

	// Select top campaigns by CTR and CPA.
	ctrRows := aggregator.TopByCTR(stats, *topN)
	cpaRows := aggregator.TopByCPA(stats, *topN)

	// Persist ranked outputs.
	if err := aggregator.WriteCSV(filepath.Join(*outputDir, *ctrFile), ctrRows); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", *ctrFile, err)
		os.Exit(1)
	}
	if err := aggregator.WriteCSV(filepath.Join(*outputDir, *cpaFile), cpaRows); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", *cpaFile, err)
		os.Exit(1)
	}

	// Print a short processing summary.
	fmt.Printf("processed %d campaigns\n", len(stats))
}
