package aggregator

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

func AggregateFileLegacy(path string) (map[string]*Campaign, int64, error) {
	// Baseline parser using byte-splitting and float parsing.
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	// Large buffer reduces syscalls for big input files.
	reader := bufio.NewReaderSize(file, 4*1024*1024)
	stats := make(map[string]*Campaign)
	var malformedRows int64
	lineNumber := 0

	// Read each line, validate, then aggregate.
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) == 0 && errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return nil, malformedRows, fmt.Errorf("read line %d: %w", lineNumber+1, readErr)
		}

		line = trimNewline(line)
		lineNumber++
		if len(line) == 0 {
			continue
		}
		if lineNumber == 1 {
			if !headerMatches(line) {
				return nil, malformedRows, fmt.Errorf("unexpected csv header: %s", string(line))
			}
			continue
		}

		// Split and parse fields in a straightforward way.
		parts := splitRowLegacy(line)
		if len(parts) != 6 {
			malformedRows++
			continue
		}

		campaignID := string(parts[0])
		impressions, ok := parseInt64(parts[2])
		if !ok {
			malformedRows++
			continue
		}
		clicks, ok := parseInt64(parts[3])
		if !ok {
			malformedRows++
			continue
		}
		spendCents, ok := parseMoneyToCentsLegacy(parts[4])
		if !ok {
			malformedRows++
			continue
		}
		conversions, ok := parseInt64(parts[5])
		if !ok {
			malformedRows++
			continue
		}

		total := stats[campaignID]
		if total == nil {
			total = &Campaign{CampaignID: campaignID}
			stats[campaignID] = total
		}
		total.TotalImpressions += impressions
		total.TotalClicks += clicks
		total.TotalSpendCents += spendCents
		total.TotalConversions += conversions
	}

	return stats, malformedRows, nil
}
