package scanner

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Dokuqui/xSentry/internal/ignore"
	"github.com/Dokuqui/xSentry/internal/rules"
)

const ignoreComment = "xSentry-ignore"

var hunkHeaderRegex = regexp.MustCompile(`^@@ \-\d+,\d+ \+(\d+),(\d+) @@`)

type ScanResult struct {
	FoundSecret bool
	Line        int
	Details     string
}

func ScanPatch(patchString string, loadedRules []rules.Rule, ign *ignore.Ignorer) (bool, error) {
	scanner := bufio.NewScanner(strings.NewReader(patchString))

	var currentFile string
	var currentLineNumber int
	var foundSecret bool

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, ignoreComment) {
			continue
		}

		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			currentLineNumber = 0
			continue
		}

		if matches := hunkHeaderRegex.FindStringSubmatch(line); len(matches) > 0 {
			startLine, err := strconv.Atoi(matches[1])
			if err == nil {
				currentLineNumber = startLine
			}
			continue
		}

		if strings.HasPrefix(line, "+") && currentFile != "" {
			scanLine := strings.TrimPrefix(line, "+")

			for _, rule := range loadedRules {
				if ign.IsRuleIgnored(rule.Name) {
					continue
				}

				matches := rule.CompiledRegex.FindAllStringSubmatch(scanLine, -1)
				if len(matches) == 0 {
					continue
				}

				if rule.Entropy > 0 {
					for _, match := range matches {
						candidate := match[0]
						if len(match) > 1 {
							candidate = match[1]
						}

						entropy := calculateShannonEntropy(candidate)
						if entropy > rule.Entropy {
							fmt.Fprintf(os.Stderr, "🚨 [xSentry] Secret found:\n")
							fmt.Fprintf(os.Stderr, "    File:   %s\n", currentFile)
							fmt.Fprintf(os.Stderr, "    Line:   %d\n", currentLineNumber)
							fmt.Fprintf(os.Stderr, "    Rule:   %s (Entropy: %.2f)\n\n", rule.Name, entropy)
							foundSecret = true
						}
					}
				} else {
					fmt.Fprintf(os.Stderr, "🚨 [xSentry] Secret found:\n")
					fmt.Fprintf(os.Stderr, "    File:   %s\n", currentFile)
					fmt.Fprintf(os.Stderr, "    Line:   %d\n", currentLineNumber)
					fmt.Fprintf(os.Stderr, "    Rule:   %s\n\n", rule.Name)
					foundSecret = true
				}
			}
			currentLineNumber++
		} else if strings.HasPrefix(line, " ") {
			currentLineNumber++
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading patch string: %w", err)
	}

	return foundSecret, nil
}

func BuildFakePatch(input string) string {
	var patch strings.Builder

	patch.WriteString("+++ b/stdin\n")

	scanner := bufio.NewScanner(strings.NewReader(input))
	lineCount := 0
	var lines []string

	for scanner.Scan() {
		lineCount++
		lines = append(lines, "+"+scanner.Text())
	}

	patch.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", lineCount, lineCount))

	patch.WriteString(strings.Join(lines, "\n"))

	return patch.String()
}
