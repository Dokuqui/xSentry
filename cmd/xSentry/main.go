package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Dokuqui/xSentry/internal/ignore"
	"github.com/Dokuqui/xSentry/internal/rules"
	"github.com/Dokuqui/xSentry/internal/scanner"
)

const defaultRulesFile = "rules.example.toml"
const defaultIgnoreFile = ".xSentry-ignore"

func main() {
	rulesPath := flag.String("rules", defaultRulesFile, "Path to the rules file")
	ignorePath := flag.String("ignore", defaultIgnoreFile, "Path to the ignore file")
	flag.Parse()

	loadedRules, err := rules.LoadRules(*rulesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error loading rules file '%s': %v\n", *rulesPath, err)
		os.Exit(2)
	}

	if len(loadedRules) == 0 {
		fmt.Fprintf(os.Stderr, "🚨 [xSentry] No rules loaded. Exiting.\n")
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "✅ [xSentry] Successfully loaded %d rules.\n", len(loadedRules))

	ign, err := ignore.NewIgnorer(*ignorePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error loading ignore file '%s': %v\n", *ignorePath, err)
		os.Exit(2)
	}

	foundSecret, err := scanner.RunScanner(os.Stdin, loadedRules, ign)
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
