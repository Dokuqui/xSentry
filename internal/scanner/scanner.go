package scanner

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Dokuqui/xSentry/internal/ignore"
	"github.com/Dokuqui/xSentry/internal/rules"
)

const ignoreComment = "xSentry-ignore"

var hunkHeaderRegex = regexp.MustCompile(`^@@ \-\d+,\d+ \+(\d+),(\d+) @@`)

type Finding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Details string `json:"details"`
}

func ScanPatch(patchString string, loadedRules []rules.Rule, ign *ignore.Ignorer) ([]Finding, error) {
	scanner := bufio.NewScanner(strings.NewReader(patchString))

	var currentFile string
	var currentLineNumber int
	var findings []Finding

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
							findings = append(findings, Finding{
								File:    currentFile,
								Line:    currentLineNumber,
								Details: fmt.Sprintf("%s (Entropy: %.2f)", rule.Name, entropy),
							})
						}
					}
				} else {
					findings = append(findings, Finding{
						File:    currentFile,
						Line:    currentLineNumber,
						Details: rule.Name,
					})
				}
			}
			currentLineNumber++
		} else if strings.HasPrefix(line, " ") {
			currentLineNumber++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading patch string: %w", err)
	}

	return findings, nil
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
