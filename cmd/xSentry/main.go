package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Dokuqui/xSentry/internal/git"
	"github.com/Dokuqui/xSentry/internal/ignore"
	"github.com/Dokuqui/xSentry/internal/reporter"
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
	installHook := flag.Bool("install-hook", false, "Install the xSentry pre-commit hook")
	scanStaged := flag.Bool("scan-staged", false, "Run in pre-commit hook mode (scans staged files)")
	reportURL := flag.String("report-url", "", "URL to POST JSON findings to")
	flag.Parse()

	if *installHook {
		err := installPreCommitHook()
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 [xSentry] Failed to install hook: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ [xSentry] Pre-commit hook installed successfully.")
		os.Exit(0)
	}

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

	var allFindings []scanner.Finding
	var scanErr error

	if *scanStaged {
		fmt.Fprintf(os.Stderr, "✅ [xSentry] Running in pre-commit hook mode...\n")
		patchString, err := git.GetStagedPatch()
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error getting staged files: %v\n", err)
			os.Exit(2)
		}
		if patchString != "" {
			findings, err := scanner.ScanPatch(patchString, loadedRules, ign)
			if err != nil {
				scanErr = err
			}
			allFindings = append(allFindings, findings...)
		}

	} else if *repoPath != "" {
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
				findings, err := scanner.ScanPatch(patchString, loadedRules, ign)
				if err != nil {
					fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error during patch scan: %v\n", err)
				}
				allFindings = append(allFindings, findings...)
			}
		} else {
			patchString, err := git.GetHeadPatch(repo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error getting HEAD patch: %v\n", err)
				os.Exit(2)
			}
			findings, err := scanner.ScanPatch(patchString, loadedRules, ign)
			if err != nil {
				scanErr = err
			}
			allFindings = append(allFindings, findings...)
		}
	} else {
		fmt.Fprintf(os.Stderr, "✅ [xSentry] Running in stdin mode...\n")
		lines, readErr := io.ReadAll(os.Stdin)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error reading from stdin: %v\n", readErr)
			os.Exit(2)
		}

		if len(lines) > 0 {
			patchString := scanner.BuildFakePatch(string(lines))
			findings, err := scanner.ScanPatch(patchString, loadedRules, ign)
			if err != nil {
				scanErr = err
			}
			allFindings = append(allFindings, findings...)
		}
	}

	if scanErr != nil {
		fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error during scan: %v\n", scanErr)
		os.Exit(2)
	}

	if err := reporter.ReportFindings(allFindings, *reportURL); err != nil {
		fmt.Fprintf(os.Stderr, "🚨 [xSentry] Error sending report: %v\n", err)
	}

	if len(allFindings) > 0 {
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "✅ [xSentry] No secrets found.\n")
	os.Exit(0)
}
