package aggregator

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

func AggregateFileParallel(path string, workers int) (map[string]*Campaign, int64, error) {
	// Parallel parser/aggregator with per-worker maps and a merge step.
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 4*1024*1024)
	rowCh := make(chan rowData, 8192)
	mergeCh := make(chan map[string]*Campaign, workers)
	var malformed atomic.Int64
	var readErr atomic.Value

	// Reader goroutine scans bytes into rows and pushes them to workers.
	go func() {
		defer close(rowCh)
		var lineBuf []byte
		lineNumber := 0
		for {
			b, err := reader.ReadByte()
			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(lineBuf) > 0 {
						if err := parseLineToRow(lineBuf, &lineNumber, rowCh, &malformed); err != nil {
							readErr.Store(err)
						}
					}
					return
				}
				readErr.Store(fmt.Errorf("read byte: %w", err))
				return
			}

			if b == '\n' {
				if err := parseLineToRow(lineBuf, &lineNumber, rowCh, &malformed); err != nil {
					readErr.Store(err)
					return
				}
				lineBuf = lineBuf[:0]
				continue
			}
			if b != '\r' {
				lineBuf = append(lineBuf, b)
			}
		}
	}()

	// Worker pool aggregates rows into local maps to avoid contention.
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := make(map[string]*Campaign)
			for row := range rowCh {
				total := local[row.CampaignID]
				if total == nil {
					total = &Campaign{CampaignID: row.CampaignID}
					local[row.CampaignID] = total
				}
				total.TotalImpressions += row.Impressions
				total.TotalClicks += row.Clicks
				total.TotalSpendCents += row.SpendCents
				total.TotalConversions += row.Conversions
			}
			mergeCh <- local
		}()
	}

	// Close the merge channel when all workers finish.
	go func() {
		wg.Wait()
		close(mergeCh)
	}()

	// Merge worker maps into a final result.
	stats := make(map[string]*Campaign)
	for local := range mergeCh {
		for campaignID, totals := range local {
			current := stats[campaignID]
			if current == nil {
				stats[campaignID] = totals
				continue
			}
			current.TotalImpressions += totals.TotalImpressions
			current.TotalClicks += totals.TotalClicks
			current.TotalSpendCents += totals.TotalSpendCents
			current.TotalConversions += totals.TotalConversions
		}
	}

	// Surface any reader errors after merging.
	if err := readErr.Load(); err != nil {
		return nil, malformed.Load(), err.(error)
	}

	return stats, malformed.Load(), nil
}
