package scanner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Dokuqui/xSentry/internal/rules"
)

const ignoreComment = "xSentry-ignore"

type ScanResult struct {
	FoundSecret bool
	Line        int
	Details     string
}

func RunScanner(input io.Reader, loadedRules []rules.Rule) (bool, error) {
	var findings []ScanResult
	lineNumber := 0
	scanner := bufio.NewScanner(input)

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if strings.Contains(line, ignoreComment) {
			continue
		}

		for _, rule := range loadedRules {
			matches := rule.CompiledRegex.FindAllStringSubmatch(line, -1)
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
						findings = append(findings, ScanResult{
							FoundSecret: true,
							Line:        lineNumber,
							Details:     fmt.Sprintf("%s (Entropy: %.2f)", rule.Name, entropy),
						})
					}
				}
			} else {
				findings = append(findings, ScanResult{
					FoundSecret: true,
					Line:        lineNumber,
					Details:     rule.Name,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading input: %w", err)
	}

	if len(findings) > 0 {
		for _, finding := range findings {
			fmt.Fprintf(os.Stderr, "🚨 [xSentry] Secret found on line %d: %s\n", finding.Line, finding.Details)
		}
		return true, nil
	}

	return false, nil
}
