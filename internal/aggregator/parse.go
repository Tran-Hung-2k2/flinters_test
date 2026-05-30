package aggregator

import (
	"bytes"
	"fmt"
	"strconv"
	"sync/atomic"
)

var expectedHeader = []string{"campaign_id", "date", "impressions", "clicks", "spend", "conversions"}

func parseLineToRow(line []byte, lineNumber *int, rowCh chan<- rowData, malformed *atomic.Int64) error {
	// Track line numbers to validate the CSV header and count rows.
	*lineNumber = *lineNumber + 1
	if len(line) == 0 {
		return nil
	}
	if *lineNumber == 1 {
		// Validate header once on the first line.
		if !headerMatches(line) {
			return fmt.Errorf("unexpected csv header: %s", string(line))
		}
		return nil
	}

	// Parse the row; malformed rows are counted and skipped.
	campaignID, impressions, clicks, spendCents, conversions, ok := parseRowFast(line)
	if !ok {
		malformed.Add(1)
		return nil
	}

	rowCh <- rowData{
		CampaignID:  campaignID,
		Impressions: impressions,
		Clicks:      clicks,
		SpendCents:  spendCents,
		Conversions: conversions,
	}
	return nil
}

func parseRowFast(line []byte) (string, int64, int64, int64, int64, bool) {
	// Fast field slicing using comma offsets to avoid allocations.
	first := bytes.IndexByte(line, ',')
	if first < 0 {
		return "", 0, 0, 0, 0, false
	}
	second := bytes.IndexByte(line[first+1:], ',')
	if second < 0 {
		return "", 0, 0, 0, 0, false
	}
	second += first + 1
	third := bytes.IndexByte(line[second+1:], ',')
	if third < 0 {
		return "", 0, 0, 0, 0, false
	}
	third += second + 1
	fourth := bytes.IndexByte(line[third+1:], ',')
	if fourth < 0 {
		return "", 0, 0, 0, 0, false
	}
	fourth += third + 1
	fifth := bytes.IndexByte(line[fourth+1:], ',')
	if fifth < 0 {
		return "", 0, 0, 0, 0, false
	}
	fifth += fourth + 1

	// Convert only the campaign_id to string; numeric fields parse from bytes.
	campaignID := string(line[:first])
	impressions, ok := parseInt64(line[second+1 : third])
	if !ok {
		return "", 0, 0, 0, 0, false
	}
	clicks, ok := parseInt64(line[third+1 : fourth])
	if !ok {
		return "", 0, 0, 0, 0, false
	}
	spendCents, ok := parseMoneyToCents(line[fourth+1 : fifth])
	if !ok {
		return "", 0, 0, 0, 0, false
	}
	conversions, ok := parseInt64(line[fifth+1:])
	if !ok {
		return "", 0, 0, 0, 0, false
	}

	return campaignID, impressions, clicks, spendCents, conversions, true
}

func headerMatches(line []byte) bool {
	// Ensure the header has expected columns in the correct order.
	parts := bytes.SplitN(line, []byte{','}, 6)
	if len(parts) != len(expectedHeader) {
		return false
	}
	for i, part := range parts {
		if string(part) != expectedHeader[i] {
			return false
		}
	}
	return true
}

func trimNewline(line []byte) []byte {
	// Trim trailing newline/carriage return without extra allocations.
	if len(line) == 0 {
		return line
	}
	if line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}
	return line
}

func parseInt64(b []byte) (int64, bool) {
	// Manual int parser avoids strconv overhead.
	if len(b) == 0 {
		return 0, false
	}
	var neg bool
	var i int
	if b[0] == '-' {
		neg = true
		i = 1
		if len(b) == 1 {
			return 0, false
		}
	}
	var n int64
	for ; i < len(b); i++ {
		c := b[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int64(c-'0')
	}
	if neg {
		n = -n
	}
	return n, true
}

func parseMoneyToCents(b []byte) (int64, bool) {
	// Parse money with up to 2 decimal places into cents.
	if len(b) == 0 {
		return 0, false // Reject empty strings as invalid money values.
	}
	var neg bool
	var i int
	if b[0] == '-' {
		neg = true
		i = 1 // Skip the negative sign for parsing.
		if len(b) == 1 {
			return 0, false // Reject lone '-' as invalid money value.
		}
	}

	var whole int64    // Whole-dollar part before the decimal point.
	var frac int64     // Fractional part after the decimal point, up to 2 digits.
	var fracDigits int // Count of digits in the fractional part to enforce max 2 decimals.
	var sawDigit bool  // Track if we've seen any digits to reject non-numeric strings.
	var afterDot bool  // Track if we've encountered the decimal point to switch parsing mode.

	// Scan bytes once:
	// - digits before '.' build the whole-dollar part
	// - up to 2 digits after '.' build the cents part
	// - reject multiple '.' or more than 2 fractional digits
	for ; i < len(b); i++ {
		c := b[i]
		switch {
		case c == '.':
			if afterDot {
				return 0, false // Reject multiple decimal points as invalid money value.
			}
			afterDot = true
		case c >= '0' && c <= '9':
			sawDigit = true
			if !afterDot {
				whole = whole*10 + int64(c-'0') // Build whole-dollar part.
			} else {
				if fracDigits >= 2 {
					return 0, false // Reject more than 2 fractional digits as invalid money value.
				}
				frac = frac*10 + int64(c-'0') // Build fractional part.
				fracDigits++
			}
		default:
			return 0, false // Reject non-numeric characters as invalid money values.
		}
	}

	if !sawDigit {
		return 0, false // Reject strings with no digits as invalid money values.
	}
	switch fracDigits {
	case 0:
		frac = 0 // No decimal point means zero cents.
	case 1:
		frac *= 10 // If only 1 fractional digit, treat it as tenths of a dollar (e.g., "12.3" -> 1230 cents).
	}

	total := whole*100 + frac // Combine whole dollars and cents into total cents.
	if neg {
		total = -total // Apply negative sign if needed.
	}
	return total, true
}

func parseMoneyToCentsLegacy(b []byte) (int64, bool) {
	// Legacy float parsing; kept for comparison/benchmarks.
	if len(b) == 0 {
		return 0, false
	}
	value, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return 0, false
	}
	return int64(value * 100), true
}

func splitRowLegacy(line []byte) [][]byte {
	// Split into 6 columns for the legacy parser.
	return bytes.SplitN(line, []byte{','}, 6)
}
