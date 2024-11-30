package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type stats struct {
	totalRequests int
	successCount  int
	failCount     int
	totalDuration time.Duration
	minDuration   time.Duration
	maxDuration   time.Duration
	mu            sync.Mutex
}

func worker(wg *sync.WaitGroup, url string, stat *stats) {
	defer wg.Done()

	start := time.Now()
	resp, err := http.Get(url)
	duration := time.Since(start)

	stat.mu.Lock()
	defer stat.mu.Unlock()

	stat.totalRequests++
	stat.totalDuration += duration

	if duration < stat.minDuration || stat.minDuration == 0 {
		stat.minDuration = duration
	}
	if duration > stat.maxDuration {
		stat.maxDuration = duration
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		stat.failCount++
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Failed with status:", resp.StatusCode)
		}
		return
	}

	stat.successCount++
	resp.Body.Close()
}

func RunBenchmarks() {
	const totalRequests = 8192
	const concurrency = 128
	url := "http://localhost:8080"

	var wg sync.WaitGroup
	stat := &stats{}

	start := time.Now()

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go worker(&wg, url, stat)
		if i%concurrency == 0 {
			time.Sleep(128 * time.Millisecond) // Adjust delay for ramping if needed.
		}
	}

	wg.Wait()
	duration := time.Since(start)

	stat.mu.Lock()
	defer stat.mu.Unlock()

	fmt.Printf("Load Test Results:\n")
	fmt.Printf("Total Requests: %d\n", stat.totalRequests)
	fmt.Printf("Success Count: %d\n", stat.successCount)
	fmt.Printf("Fail Count: %d\n", stat.failCount)
	fmt.Printf("Total Duration: %v\n", duration)
	fmt.Printf("Average Response Time: %v\n", stat.totalDuration/time.Duration(stat.totalRequests))
	fmt.Printf("Min Response Time: %v\n", stat.minDuration)
	fmt.Printf("Max Response Time: %v\n", stat.maxDuration)
	fmt.Printf("Throughput: %.2f requests/sec\n", float64(stat.totalRequests)/duration.Seconds())
}
