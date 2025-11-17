package rules

import (
	"fmt"
	"regexp"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Rules []Rule `toml:"rules"`
}

type Rule struct {
	Name     string   `toml:"name"`
	Regex    string   `toml:"regex"`
	Keywords []string `toml:"keywords"`

	CompiledRegex *regexp.Regexp `toml:"-"`
}

func LoadRules(filePath string) ([]Rule, error) {
	var config Config

	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, fmt.Errorf("failed to decode rules file: %w", err)
	}

	for i := range config.Rules {
		r := &config.Rules[i]

		compiled, err := regexp.Compile(r.Regex)
		if err != nil {
			fmt.Printf("Warning: Skipping rule '%s', failed to compile regex: %v\n", r.Name, err)
			continue
		}
		r.CompiledRegex = compiled
	}

	return config.Rules, nil
}
