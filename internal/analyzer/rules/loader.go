package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func LoadRules(configPath string) (*RuleSet, error) {
	var path string
	var err error

	if configPath != "" {
		path = configPath
	} else {
		path, err = findDefaultConfig()
		if err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules config: %w", err)
	}

	var ruleSet RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return nil, fmt.Errorf("failed to parse rules YAML: %w", err)
	}

	ruleSet.LoadedAt = time.Now()
	ruleSet.ConfigFilePath = path

	for i := range ruleSet.Rules {
		if ruleSet.Rules[i].Enabled == false {
			continue
		}
		if ruleSet.Rules[i].Enabled == true || ruleSet.Rules[i].Enabled == false {
			continue
		}
		ruleSet.Rules[i].Enabled = true
	}

	return &ruleSet, nil
}

func findDefaultConfig() (string, error) {
	paths := []string{
		"configs/heuristic-rules.yaml",
		"./configs/heuristic-rules.yaml",
		"../configs/heuristic-rules.yaml",
		"../../configs/heuristic-rules.yaml",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs, nil
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		homePath := filepath.Join(home, ".kube-triage", "heuristic-rules.yaml")
		if _, err := os.Stat(homePath); err == nil {
			return homePath, nil
		}
	}

	return "", fmt.Errorf("heuristic-rules.yaml not found in standard locations")
}

func (rs *RuleSet) GetEnabledRules() []Rule {
	var enabled []Rule
	for _, rule := range rs.Rules {
		if rule.Enabled {
			enabled = append(enabled, rule)
		}
	}
	return enabled
}

func (rs *RuleSet) GetRuleByID(id string) *Rule {
	for _, rule := range rs.Rules {
		if rule.ID == id {
			return &rule
		}
	}
	return nil
}

func (rs *RuleSet) GetRulesByCategory(category string) []Rule {
	var rules []Rule
	for _, rule := range rs.Rules {
		if rule.Category == category && rule.Enabled {
			rules = append(rules, rule)
		}
	}
	return rules
}

func (rs *RuleSet) Validate() error {
	seen := make(map[string]bool)
	for _, rule := range rs.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule missing ID")
		}
		if seen[rule.ID] {
			return fmt.Errorf("duplicate rule ID: %s", rule.ID)
		}
		seen[rule.ID] = true

		if rule.Name == "" {
			return fmt.Errorf("rule %s missing name", rule.ID)
		}
		if rule.Trigger.ContainerState == nil && rule.Trigger.Event == nil &&
			rule.Trigger.Resource == nil && rule.Trigger.NodeCondition == nil {
			return fmt.Errorf("rule %s has no trigger", rule.ID)
		}
	}
	return nil
}