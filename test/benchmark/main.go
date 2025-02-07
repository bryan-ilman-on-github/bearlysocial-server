package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
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

// func runPythonScript(csvFileName, outputImage string) error {
// 	cmd := exec.Command("py", "./test/bench/plot_response_times.py", csvFileName, outputImage) // Use "py" for Windows.
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	return cmd.Run()
// }

// ensureLogDir ensures the "log" directory exists. If not, it creates it.
func ensureLogDir() error {
	if _, err := os.Stat("log"); os.IsNotExist(err) {
		return os.Mkdir("log", 0755) // Create the directory with read/write/execute permissions.
	}
	return nil
}

func promptInteger(prompt string, defaultValue int) int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (default: %d): ", prompt, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid input detected; default value will be used instead.")
		return defaultValue
	}

	return value
}

func promptDuration(prompt string, defaultValue time.Duration) time.Duration {
    reader := bufio.NewReader(os.Stdin)
    fmt.Printf("%s (default: %v): ", prompt, defaultValue)
    input, _ := reader.ReadString('\n')
    input = strings.TrimSpace(input)

    if input == "" {
        return defaultValue
    }

    value, err := strconv.Atoi(input)
    if err != nil {
        fmt.Println("Invalid input detected; default value will be used instead.")
        return defaultValue
    }

    return time.Duration(value) * time.Millisecond
}

func promptString(prompt, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (default: %s): ", prompt, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

func main() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Define default values.
	const defaultTotalRequests = 1024
	const defaultConcurrency = 128
	const defaultSleepDuration = 256
	defaultURL := "http://localhost:80/benchmark"

	// Define flags for automatic defaults.
	autoYes := flag.Bool("yes", false, "use default values (long)")
	autoYesShort := flag.Bool("y", false, "use default values (short)")
	flag.Parse()

	// Use defaults if 'yes' or 'y' is passed.
	useDefaults := *autoYes || *autoYesShort

	var totalRequests, concurrency int
	var sleepDuration time.Duration
	var url string

	if useDefaults {
		totalRequests = defaultTotalRequests
		concurrency = defaultConcurrency
		sleepDuration = defaultSleepDuration
		url = defaultURL
	} else {
		totalRequests = promptInteger("Total Requests", defaultTotalRequests)
		concurrency = promptInteger("Concurrency", defaultConcurrency)
		sleepDuration = promptDuration("Sleep (ms)", defaultSleepDuration)
		url = promptString("URL", defaultURL)
	}

	var wg sync.WaitGroup
	stat := &stats{}

	start := time.Now()

	// Start benchmarking worker goroutines.
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go worker(&wg, url, stat)
		if i%concurrency == 0 {
			time.Sleep(sleepDuration * time.Millisecond) // Adjust delay for ramping if needed.
		}
	}

	wg.Wait()
	duration := time.Since(start)

	// Ensure the "log" directory exists.
	if err := ensureLogDir(); err != nil {
		fmt.Println("FAILED TO CREATE LOG DIRECTORY:", err)
		return
	}

	// Save response times to a CSV file.
	// csvFileName := fmt.Sprintf("log/response_time_table-%s.csv", timestamp)
	// err := saveResponseTimes(stat.responseTimes, csvFileName)
	// if err != nil {
	// 	fmt.Println("FAILED TO SAVE RESPONSE TIMES:", err)
	// 	return
	// }
	// fmt.Println("RESPONSE TIMES SAVED TO:", csvFileName)

	// Log benchmark results.
	logFileName := fmt.Sprintf("log/benchmark_results-%s.log", timestamp)
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
`, url, totalRequests, concurrency, sleepDuration,
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

	// // Run Python script to generate the graph.
	// outputImage := fmt.Sprintf("log/response_time_graph-%s.png", timestamp)
	// err = runPythonScript(csvFileName, outputImage)
	// if err != nil {
	// 	fmt.Println("FAILED TO RUN PYTHON SCRIPT:", err)
	// 	return
	// }

	// fmt.Println("RESPONSE TIME DISTRIBUTION GRAPH SAVED TO:", outputImage)
}
