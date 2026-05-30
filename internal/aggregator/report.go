package aggregator

import (
	"bufio"
	"container/heap"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type ReportRow struct {
	CampaignID       string
	TotalImpressions int64
	TotalClicks      int64
	TotalSpendCents  int64
	TotalConversions int64
	CTR              float64
	CPA              *float64
}

func TopByCTR(stats map[string]*Campaign, limit int) []ReportRow {
	// Select top-k rows by CTR using a min-heap to avoid full sorting.
	if limit <= 0 { // no limit means sort all rows
		rows := make([]ReportRow, 0, len(stats))
		for _, total := range stats {
			rows = append(rows, buildReportRow(total))
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].CTR != rows[j].CTR {
				return rows[i].CTR > rows[j].CTR
			}
			if rows[i].TotalImpressions != rows[j].TotalImpressions {
				return rows[i].TotalImpressions > rows[j].TotalImpressions
			}
			return rows[i].CampaignID < rows[j].CampaignID
		})
		return rows
	}
	rows := make([]ReportRow, 0, minInt(limit, len(stats)))
	h := &reportRowHeap{}
	heap.Init(h)
	for _, total := range stats {
		row := buildReportRow(total)
		if h.Len() < limit {
			heap.Push(h, row)
			continue
		}
		if betterCTR(row, h.rows[0]) {
			h.rows[0] = row
			heap.Fix(h, 0)
		}
	}

	rows = append(rows, h.rows...)
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CTR != rows[j].CTR {
			return rows[i].CTR > rows[j].CTR
		}
		if rows[i].TotalImpressions != rows[j].TotalImpressions {
			return rows[i].TotalImpressions > rows[j].TotalImpressions
		}
		return rows[i].CampaignID < rows[j].CampaignID
	})

	return rows
}

func TopByCPA(stats map[string]*Campaign, limit int) []ReportRow {
	// Select top-k rows by CPA (lowest) using a max-heap to avoid full sorting.
	if limit <= 0 {
		rows := make([]ReportRow, 0, len(stats))
		for _, total := range stats {
			if total.TotalConversions == 0 {
				continue
			}
			rows = append(rows, buildReportRow(total))
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].CPA != nil && rows[j].CPA != nil && *rows[i].CPA != *rows[j].CPA {
				return *rows[i].CPA < *rows[j].CPA
			}
			if rows[i].TotalConversions != rows[j].TotalConversions {
				return rows[i].TotalConversions > rows[j].TotalConversions
			}
			return rows[i].CampaignID < rows[j].CampaignID
		})
		return rows
	}
	rows := make([]ReportRow, 0, minInt(limit, len(stats)))
	h := &reportRowHeap{useCPA: true}
	heap.Init(h)
	for _, total := range stats {
		if total.TotalConversions == 0 {
			continue
		}
		row := buildReportRow(total)
		if h.Len() < limit {
			heap.Push(h, row)
			continue
		}
		if betterCPA(row, h.rows[0]) {
			h.rows[0] = row
			heap.Fix(h, 0)
		}
	}

	rows = append(rows, h.rows...)
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CPA != nil && rows[j].CPA != nil && *rows[i].CPA != *rows[j].CPA {
			return *rows[i].CPA < *rows[j].CPA
		}
		if rows[i].TotalConversions != rows[j].TotalConversions {
			return rows[i].TotalConversions > rows[j].TotalConversions
		}
		return rows[i].CampaignID < rows[j].CampaignID
	})

	return rows
}

func WriteCSV(path string, rows []ReportRow) error {
	// Ensure the target directory exists before writing.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Buffered writer reduces disk I/O overhead for large outputs.
	writer := bufio.NewWriterSize(file, 1<<20)
	if _, err := writer.WriteString("campaign_id,total_impressions,total_clicks,total_spend,total_conversions,CTR,CPA\n"); err != nil {
		return err
	}

	for _, row := range rows {
		if _, err := writer.WriteString(row.CampaignID); err != nil {
			return err
		}
		if err := writer.WriteByte(','); err != nil {
			return err
		}
		if _, err := writer.WriteString(strconv.FormatInt(row.TotalImpressions, 10)); err != nil {
			return err
		}
		if err := writer.WriteByte(','); err != nil {
			return err
		}
		if _, err := writer.WriteString(strconv.FormatInt(row.TotalClicks, 10)); err != nil {
			return err
		}
		if err := writer.WriteByte(','); err != nil {
			return err
		}
		if _, err := writer.WriteString(formatCents(row.TotalSpendCents)); err != nil {
			return err
		}
		if err := writer.WriteByte(','); err != nil {
			return err
		}
		if _, err := writer.WriteString(strconv.FormatInt(row.TotalConversions, 10)); err != nil {
			return err
		}
		if err := writer.WriteByte(','); err != nil {
			return err
		}
		if _, err := writer.WriteString(strconv.FormatFloat(row.CTR, 'f', 4, 64)); err != nil {
			return err
		}
		if err := writer.WriteByte(','); err != nil {
			return err
		}
		if row.CPA != nil {
			if _, err := writer.WriteString(strconv.FormatFloat(*row.CPA, 'f', 2, 64)); err != nil {
				return err
			}
		}
		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}

	return writer.Flush()
}

func buildReportRow(total *Campaign) ReportRow {
	// Precompute CTR/CPA values for the report output.
	row := ReportRow{
		CampaignID:       total.CampaignID,
		TotalImpressions: total.TotalImpressions,
		TotalClicks:      total.TotalClicks,
		TotalSpendCents:  total.TotalSpendCents,
		TotalConversions: total.TotalConversions,
	}
	if total.TotalImpressions > 0 {
		row.CTR = float64(total.TotalClicks) / float64(total.TotalImpressions)
	}
	if total.TotalConversions > 0 {
		cpa := float64(total.TotalSpendCents) / 100.0 / float64(total.TotalConversions)
		row.CPA = &cpa
	}
	return row
}

func formatCents(cents int64) string {
	// Render cents as a fixed 2-decimal string.
	if cents < 0 {
		return "-" + formatCents(-cents)
	}
	whole := cents / 100
	frac := cents % 100
	if frac < 10 {
		return strconv.FormatInt(whole, 10) + ".0" + strconv.FormatInt(frac, 10)
	}
	return strconv.FormatInt(whole, 10) + "." + strconv.FormatInt(frac, 10)
}

// reportRowHeap implements a heap.Interface for ReportRow, supporting both CTR and CPA ordering based on the useCPA flag.
type reportRowHeap struct {
	rows   []ReportRow
	useCPA bool
}

func (h reportRowHeap) Len() int {
	return len(h.rows)
}

func (h reportRowHeap) Less(i, j int) bool {
	if h.useCPA {
		// Max-heap on CPA; keep worst CPA at root.
		if h.rows[i].CPA != nil && h.rows[j].CPA != nil && *h.rows[i].CPA != *h.rows[j].CPA {
			return *h.rows[i].CPA > *h.rows[j].CPA
		}
		if h.rows[i].TotalConversions != h.rows[j].TotalConversions {
			return h.rows[i].TotalConversions < h.rows[j].TotalConversions
		}
		return h.rows[i].CampaignID > h.rows[j].CampaignID
	}

	// Min-heap on CTR; keep worst CTR at root.
	if h.rows[i].CTR != h.rows[j].CTR {
		return h.rows[i].CTR < h.rows[j].CTR
	}
	if h.rows[i].TotalImpressions != h.rows[j].TotalImpressions {
		return h.rows[i].TotalImpressions < h.rows[j].TotalImpressions
	}
	return h.rows[i].CampaignID > h.rows[j].CampaignID
}

func (h reportRowHeap) Swap(i, j int) {
	h.rows[i], h.rows[j] = h.rows[j], h.rows[i]
}

func (h *reportRowHeap) Push(x any) {
	h.rows = append(h.rows, x.(ReportRow))
}

func (h *reportRowHeap) Pop() any {
	last := len(h.rows) - 1
	item := h.rows[last]
	h.rows = h.rows[:last]
	return item
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func betterCTR(a, b ReportRow) bool {
	if a.CTR != b.CTR {
		return a.CTR > b.CTR
	}
	if a.TotalImpressions != b.TotalImpressions {
		return a.TotalImpressions > b.TotalImpressions
	}
	return a.CampaignID < b.CampaignID
}

func betterCPA(a, b ReportRow) bool {
	if a.CPA != nil && b.CPA != nil && *a.CPA != *b.CPA {
		return *a.CPA < *b.CPA
	}
	if a.TotalConversions != b.TotalConversions {
		return a.TotalConversions > b.TotalConversions
	}
	return a.CampaignID < b.CampaignID
}
