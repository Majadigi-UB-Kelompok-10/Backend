package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	n := flag.Int("n", 10000, "total requests (simulasi jumlah user unik)")
	c := flag.Int("c", 200, "concurrency (goroutine paralel)")
	url := flag.String("url", "http://localhost:8888/health", "target URL")
	flag.Parse()

	fmt.Printf("Target  : %s\n", *url)
	fmt.Printf("Users   : %d (tiap user = IP unik)\n", *n)
	fmt.Printf("Parallel: %d goroutine sekaligus\n\n", *c)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: *c + 10,
			MaxConnsPerHost:     *c + 10,
		},
	}

	type result struct {
		status   int
		duration time.Duration
		err      bool
	}

	results := make([]result, *n)
	sem := make(chan struct{}, *c)
	var wg sync.WaitGroup
	var done atomic.Int64

	start := time.Now()

	for i := 0; i < *n; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			ip := fmt.Sprintf("10.%d.%d.%d", (idx/65536)%256, (idx/256)%256, idx%256)

			req, _ := http.NewRequest("GET", *url, nil)
			req.Header.Set("X-Forwarded-For", ip)

			t0 := time.Now()
			resp, err := client.Do(req)
			dur := time.Since(t0)

			r := result{duration: dur}
			if err != nil {
				r.err = true
			} else {
				r.status = resp.StatusCode
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			results[idx] = r

			cnt := done.Add(1)
			if cnt%1000 == 0 {
				fmt.Printf("  ... %d/%d selesai (%.1fs)\n", cnt, int64(*n), time.Since(start).Seconds())
			}
		}(i)
	}

	wg.Wait()
	total := time.Since(start)

	statusCount := map[int]int{}
	var errCount, successCount int
	var durations []float64

	for _, r := range results {
		if r.err {
			errCount++
			continue
		}
		statusCount[r.status]++
		durations = append(durations, r.duration.Seconds()*1000) // ms
		if r.status >= 200 && r.status < 300 {
			successCount++
		}
	}

	sort.Float64s(durations)

	pct := func(p float64) float64 {
		if len(durations) == 0 {
			return 0
		}
		idx := int(float64(len(durations)) * p / 100)
		if idx >= len(durations) {
			idx = len(durations) - 1
		}
		return durations[idx]
	}

	var total_ms float64
	for _, d := range durations {
		total_ms += d
	}
	avg := 0.0
	if len(durations) > 0 {
		avg = total_ms / float64(len(durations))
	}

	rps := float64(*n) / total.Seconds()

	fmt.Println("\n========================================")
	fmt.Printf("HASIL LOAD TEST — %d user unik\n", *n)
	fmt.Println("========================================")
	fmt.Printf("Total waktu      : %.3f detik\n", total.Seconds())
	fmt.Printf("Throughput       : %.0f req/s\n", rps)
	fmt.Println()
	fmt.Println("Status:")
	for code, cnt := range statusCount {
		persen := float64(cnt) / float64(*n) * 100
		fmt.Printf("  %d : %d (%.1f%%)\n", code, cnt, persen)
	}
	if errCount > 0 {
		fmt.Printf("  error: %d (%.1f%%)\n", errCount, float64(errCount)/float64(*n)*100)
	}
	fmt.Println()
	fmt.Println("Latency (ms):")
	fmt.Printf("  min   : %.2f ms\n", pct(0))
	fmt.Printf("  p50   : %.2f ms\n", pct(50))
	fmt.Printf("  p75   : %.2f ms\n", pct(75))
	fmt.Printf("  p90   : %.2f ms\n", pct(90))
	fmt.Printf("  p95   : %.2f ms\n", pct(95))
	fmt.Printf("  p99   : %.2f ms\n", pct(99))
	fmt.Printf("  max   : %.2f ms\n", pct(100))
	fmt.Printf("  avg   : %.2f ms\n", avg)
	fmt.Println()
	fmt.Printf("Sukses: %d/%d (%.1f%%)\n", successCount, *n, float64(successCount)/float64(*n)*100)

	if float64(successCount)/float64(*n) >= 0.99 {
		fmt.Println("\n✓ LULUS: server mampu menangani semua user tanpa gangguan")
	} else {
		fmt.Printf("\n✗ PERHATIAN: %.1f%% request gagal\n", float64(*n-successCount)/float64(*n)*100)
	}
}
