package types

import (
	"os"

	"go.xrstf.de/kubesort/pkg/sort"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	FlattenLists              bool               `yaml:"flattenLists"`
	ObjectRules               []sort.SortingRule `yaml:"objectRules"`
	DisableDefaultObjectRules bool               `yaml:"disableDefaultObjectRules"`
}

func (c *Configuration) Validate() error {
	for _, rule := range c.ObjectRules {
		if err := rule.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func LoadConfig(filename string) (*Configuration, error) {
	cfg := &Configuration{}

	if filename != "" {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}

		if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
			return nil, err
		}

		if err := cfg.Validate(); err != nil {
			return nil, err
		}
	}

	// prepend the default rules
	if !cfg.DisableDefaultObjectRules {
		cfg.ObjectRules = append(defaultObjectRules, cfg.ObjectRules...)
	}

	return cfg, nil
}
