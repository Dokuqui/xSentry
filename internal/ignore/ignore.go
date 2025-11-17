package ignore

import (
	"bufio"
	"os"
	"strings"
)

type Ignorer struct {
	ignoredRules map[string]bool
}

func NewIgnorer(filePath string) (*Ignorer, error) {
	ign := &Ignorer{
		ignoredRules: make(map[string]bool),
	}

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ign, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if !strings.Contains(line, "/") {
			ign.ignoredRules[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ign, nil
}

func (i *Ignorer) IsRuleIgnored(ruleName string) bool {
	return i.ignoredRules[ruleName]
}
