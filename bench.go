package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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
	responseTimes []int64 // To store response times in milliseconds.
	mu            sync.Mutex
}

func worker(wg *sync.WaitGroup, url string, stat *stats) {
	defer wg.Done()

	start := time.Now()
	resp, err := http.Get(url)
	duration := time.Since(start)
	durationMs := duration.Milliseconds()

	stat.mu.Lock()
	defer stat.mu.Unlock()

	stat.responseTimes = append(stat.responseTimes, durationMs)
	stat.totalRequests++
	stat.totalDuration += duration

	if duration < stat.minDuration || stat.minDuration == 0 {
		stat.minDuration = duration
	}
	if duration > stat.maxDuration {
		stat.maxDuration = duration
	}

	if err != nil {
		stat.failCount++
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		// body, _ := ioutil.ReadAll(resp.Body) // Read the response body.
		// fmt.Printf("400 RESPONSE BODY: %s\n", string(body))
		stat.failCount++
		return
	}

	if resp.StatusCode != http.StatusOK {
		stat.failCount++
		return
	}

	stat.successCount++
}

func saveResponseTimes(responseTimes []int64, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header.
	writer.Write([]string{"RequestNumber", "ResponseTime(ms)"})

	// Write data.
	for i, time := range responseTimes {
		writer.Write([]string{fmt.Sprintf("%d", i+1), fmt.Sprintf("%d", time)})
	}
	return nil
}

func runPythonScript(csvFileName, outputImage string) error {
	cmd := exec.Command("py", "plot_response_times.py", csvFileName, outputImage) // Use "py" for Windows.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunBenchmarks() {
	const totalRequests = 1024
	const concurrency = 128
	const sleep = 256
	url := "http://localhost:8080/data"

	var wg sync.WaitGroup
	stat := &stats{}

	start := time.Now()

	// Start benchmarking worker goroutines.
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go worker(&wg, url, stat)
		if i%concurrency == 0 {
			time.Sleep(sleep * time.Millisecond) // Adjust delay for ramping if needed.
		}
	}

	wg.Wait()
	duration := time.Since(start)

	// Save response times to a CSV file.
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	csvFileName := fmt.Sprintf("response_times-%s.csv", timestamp)
	err := saveResponseTimes(stat.responseTimes, csvFileName)
	if err != nil {
		fmt.Println("FAILED TO SAVE RESPONSE TIMES:", err)
		return
	}
	fmt.Println("RESPONSE TIMES SAVED TO:", csvFileName)

	// Log benchmark results.
	logFileName := fmt.Sprintf("bench-%s.log", timestamp)
	file, err := os.Create(logFileName)
	if err != nil {
		fmt.Println("FAILED TO CREATE LOG FILE:", err)
		return
	}
	defer file.Close()

	logContent := fmt.Sprintf(`
TARGET: %s
TOTAL REQUESTS: %d
CONCURRENCY: %d
SLEEP: %d ms

BENCHMARK RESULTS:
DATE: %s
TOTAL REQUESTS: %d
SUCCESS COUNT: %d
FAIL COUNT: %d

TOTAL DURATION: %v
AVERAGE RESPONSE TIME: %v
MIN RESPONSE TIME: %v
MAX RESPONSE TIME: %v

THROUGHPUT: %.2f REQUESTS/SEC
`, url, totalRequests, concurrency, sleep,
time.Now().Format("Monday, 02 January 2006 15:04:05"),
stat.totalRequests, stat.successCount, stat.failCount,
duration, stat.totalDuration/time.Duration(stat.totalRequests),
stat.minDuration, stat.maxDuration, float64(stat.totalRequests)/duration.Seconds(),
)

	_, err = file.WriteString(logContent)
	if err != nil {
		fmt.Println("FAILED TO WRITE TO LOG FILE:", err)
		return
	}

	fmt.Println("BENCHMARK RESULTS SAVED TO:", logFileName)

	// Run Python script to generate the graph.
	outputImage := fmt.Sprintf("response_time_distribution-%s.png", timestamp)
	err = runPythonScript(csvFileName, outputImage)
	if err != nil {
		fmt.Println("FAILED TO RUN PYTHON SCRIPT:", err)
		return
	}

	fmt.Println("RESPONSE TIME DISTRIBUTION GRAPH SAVED TO:", outputImage)
}
