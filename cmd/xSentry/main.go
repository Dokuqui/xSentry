package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Dokuqui/xSentry/internal/rules"
)

const defaultRulesFile = "rules.example.toml"

const ignoreComment = "xSentry-ignore"

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

	foundSecret := false
	lineNumber := 0
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if strings.Contains(line, ignoreComment) {
			continue
		}

		for _, rule := range loadedRules {
			if rule.CompiledRegex == nil {
				continue
			}

			if rule.CompiledRegex.MatchString(line) {
				fmt.Fprintf(os.Stderr, "🚨 [xSentry] Secret found on line %d: %s\n", lineNumber, rule.Name)
				foundSecret = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(2)
	}

	if foundSecret {
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "✅ [xSentry] No secrets found.\n")
	os.Exit(0)
}
