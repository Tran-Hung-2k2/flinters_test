package aggregator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func BenchmarkAggregateFileFast(b *testing.B) {
	benchmarkAggregateFile(b, AggregateFile)
}

func BenchmarkAggregateFileLegacy(b *testing.B) {
	benchmarkAggregateFile(b, AggregateFileLegacy)
}

func BenchmarkAggregateFileParallel(b *testing.B) {
	benchmarkAggregateFile(b, func(path string) (map[string]*Campaign, int64, error) {
		return AggregateFileParallel(path, 0)
	})
}

func benchmarkAggregateFile(b *testing.B, fn func(string) (map[string]*Campaign, int64, error)) {
	path := benchmarkInputPath(b)
	b.ReportAllocs()
	b.ResetTimer()

	var maxPeak atomic.Uint64
	var maxCPU atomic.Uint64
	for i := 0; i < b.N; i++ {
		peak, cpuNs, stats, malformed, err := runWithPeak(fn, path)
		if err != nil {
			b.Fatalf("aggregate: %v", err)
		}
		if malformed != 0 {
			b.Fatalf("unexpected malformed rows: %d", malformed)
		}
		updatePeak(&maxPeak, peak)
		updatePeak(&maxCPU, cpuNs)
		runtime.KeepAlive(stats)
	}

	b.ReportMetric(float64(maxPeak.Load())/1024.0/1024.0, "peak_heap_MB")
	b.ReportMetric(float64(maxCPU.Load())/1e6, "cpu_ms")
}

func benchmarkInputPath(b *testing.B) string {
	b.Helper()
	if value := os.Getenv("AGG_BENCH_INPUT"); value != "" {
		return value
	}
	if os.Getenv("AGG_BENCH_FULL") == "1" {
		path := filepath.Clean(filepath.Join("..", "..", "ad_data.csv"))
		if _, err := os.Stat(path); err != nil {
			b.Fatalf("missing full dataset at %s", path)
		}
		return path
	}
	return createBenchmarkCSV(b, 100000, 2000)
}

func createBenchmarkCSV(b *testing.B, rows int, campaigns int) string {
	b.Helper()

	dir := b.TempDir()
	path := filepath.Join(dir, "benchmark.csv")
	file, err := os.Create(path)
	if err != nil {
		b.Fatalf("create benchmark csv: %v", err)
	}
	writer := bufio.NewWriterSize(file, 1<<20)
	if _, err := writer.WriteString("campaign_id,date,impressions,clicks,spend,conversions\n"); err != nil {
		b.Fatalf("write header: %v", err)
	}

	for i := 0; i < rows; i++ {
		campaignIndex := i % campaigns
		impressions := 1000 + (i % 9000)
		clicks := 10 + (i % 500)
		spend := 25 + (i % 250)
		conversions := 1 + (i % 40)
		if _, err := fmt.Fprintf(writer, "CMP%05d,2025-01-%02d,%d,%d,%d.%02d,%d\n", campaignIndex, (i%28)+1, impressions, clicks, spend, i%100, conversions); err != nil {
			b.Fatalf("write row: %v", err)
		}
	}

	if err := writer.Flush(); err != nil {
		b.Fatalf("flush benchmark csv: %v", err)
	}
	if err := file.Close(); err != nil {
		b.Fatalf("close benchmark csv: %v", err)
	}

	return path
}

func runWithPeak(fn func(string) (map[string]*Campaign, int64, error), path string) (uint64, uint64, map[string]*Campaign, int64, error) {
	done := make(chan struct{})
	var peak atomic.Uint64
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(2 * time.Millisecond)
		defer ticker.Stop()
		var mem runtime.MemStats
		for {
			select {
			case <-ticker.C:
				runtime.ReadMemStats(&mem)
				updatePeak(&peak, mem.Alloc)
			case <-done:
				return
			}
		}
	}()
	startCPU := readCPUTimeNs()
	stats, malformed, err := fn(path)
	cpuNs := readCPUTimeNs() - startCPU
	close(done)
	wg.Wait()
	return peak.Load(), cpuNs, stats, malformed, err
}

func updatePeak(target *atomic.Uint64, value uint64) {
	for {
		current := target.Load()
		if value <= current {
			return
		}
		if target.CompareAndSwap(current, value) {
			return
		}
	}
}

func readCPUTimeNs() uint64 {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0
	}
	user := uint64(ru.Utime.Sec)*1e9 + uint64(ru.Utime.Usec)*1e3
	sys := uint64(ru.Stime.Sec)*1e9 + uint64(ru.Stime.Usec)*1e3
	return user + sys
}
