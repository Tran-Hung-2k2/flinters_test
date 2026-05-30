package aggregator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAggregateAndRank(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.csv")
	content := strings.Join([]string{
		"campaign_id,date,impressions,clicks,spend,conversions",
		"CMP001,2025-01-01,100,10,20.00,2",
		"CMP001,2025-01-02,200,20,40.00,4",
		"CMP002,2025-01-01,100,5,15.00,0",
		"CMP003,2025-01-01,50,5,10.00,1",
	}, "\n") + "\n"
	if err := os.WriteFile(input, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	stats, malformed, err := AggregateFile(input)
	if err != nil {
		t.Fatalf("aggregate: %v", err)
	}
	if malformed != 0 {
		t.Fatalf("expected no malformed rows, got %d", malformed)
	}

	if got := stats["CMP001"]; got == nil || got.TotalImpressions != 300 || got.TotalClicks != 30 || got.TotalSpendCents != 6000 || got.TotalConversions != 6 {
		t.Fatalf("unexpected totals for CMP001: %+v", got)
	}

	ctrRows := TopByCTR(stats, 10)
	if len(ctrRows) != 3 {
		t.Fatalf("expected 3 CTR rows, got %d", len(ctrRows))
	}
	if ctrRows[0].CampaignID != "CMP001" || ctrRows[1].CampaignID != "CMP003" || ctrRows[2].CampaignID != "CMP002" {
		t.Fatalf("unexpected CTR order: %+v", ctrRows)
	}
	if ctrRows[2].CPA != nil {
		t.Fatalf("expected nil CPA for zero-conversion campaign")
	}

	cpaRows := TopByCPA(stats, 10)
	if len(cpaRows) != 2 {
		t.Fatalf("expected 2 CPA rows, got %d", len(cpaRows))
	}
	if cpaRows[0].CampaignID != "CMP001" || cpaRows[1].CampaignID != "CMP003" {
		t.Fatalf("unexpected CPA order: %+v", cpaRows)
	}
}

func TestWriteCSV(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out.csv")
	rows := []ReportRow{{
		CampaignID:       "CMP001",
		TotalImpressions: 300,
		TotalClicks:      30,
		TotalSpendCents:  6000,
		TotalConversions: 6,
		CTR:              0.1,
		CPA:              floatPtr(10),
	}, {
		CampaignID:       "CMP002",
		TotalImpressions: 100,
		TotalClicks:      1,
		TotalSpendCents:  1234,
		TotalConversions: 0,
		CTR:              0.01,
		CPA:              nil,
	}}

	if err := WriteCSV(output, rows); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	got := string(data)
	expected := strings.Join([]string{
		"campaign_id,total_impressions,total_clicks,total_spend,total_conversions,CTR,CPA",
		"CMP001,300,30,60.00,6,0.1000,10.00",
		"CMP002,100,1,12.34,0,0.0100,",
		"",
	}, "\n")
	if got != expected {
		t.Fatalf("unexpected csv output:\n%s", got)
	}
}

func TestTopByCTRLimitAndOrder(t *testing.T) {
	stats := make(map[string]*Campaign)
	for i := 1; i <= 12; i++ {
		campaignID := fmt.Sprintf("CMP%03d", i)
		stats[campaignID] = &Campaign{
			CampaignID:       campaignID,
			TotalImpressions: 100,
			TotalClicks:      int64(i),
			TotalSpendCents:  int64(i) * 100,
			TotalConversions: int64(i),
		}
	}

	rows := TopByCTR(stats, 10)
	if len(rows) != 10 {
		t.Fatalf("expected 10 CTR rows, got %d", len(rows))
	}

	for i := 0; i < 10; i++ {
		expectedID := fmt.Sprintf("CMP%03d", 12-i)
		if rows[i].CampaignID != expectedID {
			t.Fatalf("unexpected CTR order at %d: got %s, want %s", i, rows[i].CampaignID, expectedID)
		}
	}
}

func TestTopByCPAExcludesZeroAndOrder(t *testing.T) {
	stats := map[string]*Campaign{
		"CMP001": {
			CampaignID:       "CMP001",
			TotalImpressions: 1000,
			TotalClicks:      100,
			TotalSpendCents:  5000,
			TotalConversions: 10, // CPA = 5.00
		},
		"CMP002": {
			CampaignID:       "CMP002",
			TotalImpressions: 1000,
			TotalClicks:      100,
			TotalSpendCents:  7000,
			TotalConversions: 10, // CPA = 7.00
		},
		"CMP003": {
			CampaignID:       "CMP003",
			TotalImpressions: 1000,
			TotalClicks:      100,
			TotalSpendCents:  9000,
			TotalConversions: 10, // CPA = 9.00
		},
		"CMP004": {
			CampaignID:       "CMP004",
			TotalImpressions: 1000,
			TotalClicks:      100,
			TotalSpendCents:  1234,
			TotalConversions: 0, // excluded
		},
	}

	rows := TopByCPA(stats, 10)
	if len(rows) != 3 {
		t.Fatalf("expected 3 CPA rows, got %d", len(rows))
	}

	expectedOrder := []string{"CMP001", "CMP002", "CMP003"}
	for i, expectedID := range expectedOrder {
		if rows[i].CampaignID != expectedID {
			t.Fatalf("unexpected CPA order at %d: got %s, want %s", i, rows[i].CampaignID, expectedID)
		}
	}
}

func TestAggregateFileMissingInput(t *testing.T) {
	_, _, err := AggregateFile(filepath.Join(t.TempDir(), "missing.csv"))
	if err == nil {
		t.Fatalf("expected error for missing input file")
	}
}

func TestAggregateFileMalformedRowsCount(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.csv")
	content := strings.Join([]string{
		"campaign_id,date,impressions,clicks,spend,conversions",
		"CMP001,2025-01-01,100,10,20.00,2",
		"CMP002,2025-01-01,not-int,5,15.00,1",   // bad impressions
		"CMP003,2025-01-01,100,5,15.00",         // missing conversions
		"CMP004,2025-01-01,100,5,15.00,1,extra", // too many fields
		"CMP005,2025-01-01,100,5,12.345,1",      // too many decimals
		"CMP007,2025-01-01,100,,15.00,1",        // empty clicks
		"CMP008,2025-01-01,100,5,15.00,",        // empty conversions
		"CMP009,2025-01-01,100,5,15..00,1",      // invalid spend format
		"CMP010,2025-01-01,100,5,15.00,1x",      // non-numeric conversions
		"CMP006,2025-01-01,100,5,15.00,1",
		"",
	}, "\n")
	if err := os.WriteFile(input, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	stats, malformed, err := AggregateFile(input)
	if err != nil {
		t.Fatalf("aggregate: %v", err)
	}
	if malformed != 8 {
		t.Fatalf("expected 8 malformed rows, got %d", malformed)
	}
	if got := stats["CMP001"]; got == nil || got.TotalImpressions != 100 || got.TotalClicks != 10 {
		t.Fatalf("unexpected totals for CMP001: %+v", got)
	}
	if got := stats["CMP006"]; got == nil || got.TotalImpressions != 100 || got.TotalClicks != 5 {
		t.Fatalf("unexpected totals for CMP006: %+v", got)
	}
}

func TestAggregateFileHeaderMismatch(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.csv")
	content := strings.Join([]string{
		"campaign_id,date,impressions,clicks,spend", // missing conversions
		"CMP001,2025-01-01,100,10,20.00",
	}, "\n")
	if err := os.WriteFile(input, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	_, _, err := AggregateFile(input)
	if err == nil {
		t.Fatalf("expected header mismatch error")
	}
}

func TestAggregateFileLineTooLarge(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.csv")
	veryLong := strings.Repeat("A", 4*1024*1024+128)
	content := strings.Join([]string{
		"campaign_id,date,impressions,clicks,spend,conversions",
		"CMP001,2025-01-01,100,10,20.00,2",
		veryLong,
	}, "\n")
	if err := os.WriteFile(input, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	_, _, err := AggregateFile(input)
	if err == nil {
		t.Fatalf("expected error for oversized line")
	}
	if !strings.Contains(err.Error(), "input line too large for buffer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func floatPtr(v float64) *float64 {
	return &v
}
