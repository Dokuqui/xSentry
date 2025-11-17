package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Dokuqui/xSentry/internal/git"
	"github.com/Dokuqui/xSentry/internal/ignore"
	"github.com/Dokuqui/xSentry/internal/rules"
	"github.com/Dokuqui/xSentry/internal/scanner"
)

const defaultRulesFile = "rules.example.toml"
const defaultIgnoreFile = ".xSentry-ignore"

func main() {
	rulesPath := flag.String("rules", defaultRulesFile, "Path to the rules file")
	ignorePath := flag.String("ignore", defaultIgnoreFile, "Path to the ignore file")
	repoPath := flag.String("path", "", "Path to a Git repository to scan")
	scanHistory := flag.Bool("scan-history", false, "Scan all commits in history")
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

	var foundSecret bool
	if *repoPath != "" {
		fmt.Fprintf(os.Stderr, "✅ [xSentry] Running in Git-aware mode on path: %s\n", *repoPath)

		repo, err := git.OpenRepository(*repoPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 [xSentry] %v\n", err)
			os.Exit(2)
		}

		if *scanHistory {
			fmt.Fprintf(os.Stderr, "✅ [xSentry] Starting full history scan...\n\n")

			patchChannel, err := git.GetCommitPatches(repo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error getting commit history: %v\n", err)
				os.Exit(2)
			}

			for patchString := range patchChannel {
				foundInPatch, scanErr := scanner.ScanPatch(patchString, loadedRules, ign)
				if scanErr != nil {
					fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error during patch scan: %v\n", scanErr)
				}
				if foundInPatch {
					foundSecret = true
				}
			}
		} else {
			patchString, err := git.GetHeadPatch(repo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error getting HEAD patch: %v\n", err)
				os.Exit(2)
			}
			foundSecret, err = scanner.ScanPatch(patchString, loadedRules, ign)
			if err != nil {
				fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error during patch scan: %v\n", err)
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "✅ [xSentry] Running in stdin mode...\n")

		lines, readErr := io.ReadAll(os.Stdin)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error reading from stdin: %v\n", readErr)
			os.Exit(2)
		}

		if len(lines) == 0 {
			err = nil
			foundSecret = false
		} else {
			patchString := scanner.BuildFakePatch(string(lines))

			foundSecret, err = scanner.ScanPatch(patchString, loadedRules, ign)
		}
	}

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
