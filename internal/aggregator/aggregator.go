package aggregator

type Campaign struct {
	// Aggregate totals per campaign_id.
	CampaignID       string
	TotalImpressions int64
	TotalClicks      int64
	TotalSpendCents  int64
	TotalConversions int64
}

type rowData struct {
	// Parsed row values passed to workers in the parallel pipeline.
	CampaignID  string
	Impressions int64
	Clicks      int64
	SpendCents  int64
	Conversions int64
}

func AggregateFile(path string) (map[string]*Campaign, int64, error) {
	// Default to the fast single-threaded parser.
	return aggregateFast(path)
}
