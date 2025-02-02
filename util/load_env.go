package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Load environment variables from a .env file.
func LoadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		fmt.Println("ERROR OPENING .env FILE:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create a scanner to read the file line by line.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignore empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split "KEY=VALUE" pairs.
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present.
		value = strings.Trim(value, `"'`)

		// Set environment variable manually.
		os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("ERROR READING .env FILE:", err)
		os.Exit(1)
	}
}
