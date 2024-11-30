package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, World!"))
}

func main() {
	mode := flag.String("mode", "server", "Mode to run: 'server' or 'bench'.")
	flag.Parse()

	if *mode == "bench" {
		RunBenchmarks()
		os.Exit(0) // Exit after running benchmarks.
	}

	http.HandleFunc("/", helloHandler)

	// Use a high-performance server configuration.
	server := &http.Server{
		Addr: ":8080", // Listen on port 8080.
	}

	fmt.Println("Starting server on :8080.")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server.")
	}
}
