package main

import (
	"fmt"
	"os"

	"github.com/Dokuqui/xSentry/internal/rules"
	"github.com/Dokuqui/xSentry/internal/scanner"
)

const defaultRulesFile = "rules.example.toml"

func main() {
	loadedRules, err := rules.LoadRules(defaultRulesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error loading rules file '%s': %v\n", defaultRulesFile, err)
		os.Exit(2)
	}

	if len(loadedRules) == 0 {
		fmt.Fprintf(os.Stderr, "🚨 [xSentry] No rules loaded. Exiting.\n")
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "✅ [xSentry] Successfully loaded %d rules.\n", len(loadedRules))

	foundSecret, err := scanner.RunScanner(os.Stdin, loadedRules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error during scan: %v\n", err)
		os.Exit(2)
	}

	if foundSecret {
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "✅ [xSentry] No secrets found.\n")
	os.Exit(0)
}
