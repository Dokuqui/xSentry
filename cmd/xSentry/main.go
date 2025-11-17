package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func main() {
	awsKeyRegex := regexp.MustCompile(`AKIA[A-Z0-9]{16}`)

	foundSecret := false
	lineNumber := 0

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if awsKeyRegex.MatchString(line) {
			fmt.Fprintf(os.Stderr, "🚨 [xSentry] Secret found on line %d: AWS Access Key\n", lineNumber)
			foundSecret = true
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(2)
	}

	if foundSecret {
		os.Exit(1)
	}

	fmt.Println("✅ [xSentry] No secrets found.")
	os.Exit(0)
}
