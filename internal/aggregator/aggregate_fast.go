package aggregator

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

func aggregateFast(path string) (map[string]*Campaign, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 4*1024*1024)
	stats := make(map[string]*Campaign)
	var malformedRows int64
	lineNumber := 0

	// Stream lines to avoid loading the whole file into memory.
	for {
		line, readErr := reader.ReadSlice('\n')
		if len(line) == 0 && errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			if errors.Is(readErr, bufio.ErrBufferFull) {
				return nil, malformedRows, fmt.Errorf("input line too large for buffer")
			}
			return nil, malformedRows, fmt.Errorf("read line %d: %w", lineNumber+1, readErr)
		}

		// Normalize line endings before parsing.
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

		// Parse and aggregate the row into totals.
		campaignID, impressions, clicks, spendCents, conversions, ok := parseRowFast(line)
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
