package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const FileName = ".term.yml"

type TermConfig struct {
	Project   string              `yaml:"project"`
	History   HistoryConfig       `yaml:"history,omitempty"`
	Shortcuts map[string]Shortcut `yaml:"shortcuts"`
}

type HistoryConfig struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

func (h HistoryConfig) IsEnabled() bool {
	return h.Enabled == nil || *h.Enabled
}

type Shortcut struct {
	Description string   `yaml:"description"`
	Parallel    bool     `yaml:"parallel,omitempty"`
	Confirm     bool     `yaml:"confirm,omitempty"`
	Danger      string   `yaml:"danger,omitempty"`
	Args        []string `yaml:"args,omitempty"`
	Steps       []Step   `yaml:"steps"`
}

type Step struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

func (s *Step) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		s.Command = value.Value
		return nil
	case yaml.MappingNode:
		type raw Step
		return value.Decode((*raw)(s))
	default:
		return fmt.Errorf("step must be a string or object")
	}
}

func (s Step) MarshalYAML() (any, error) {
	if s.Name == "" {
		return s.Command, nil
	}
	type raw Step
	return raw(s), nil
}

func Load(dir string) (*TermConfig, error) {
	path := filepath.Join(dir, FileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg TermConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Shortcuts == nil {
		cfg.Shortcuts = map[string]Shortcut{}
	}
	return &cfg, nil
}

func Write(dir string, cfg TermConfig) error {
	if cfg.Project == "" {
		return errors.New("project is required")
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, FileName), data, 0644)
}

func FindNearest(start string) (string, *TermConfig, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", nil, err
	}
	for {
		cfg, err := Load(dir)
		if err == nil {
			return dir, cfg, nil
		}
		if !os.IsNotExist(err) {
			return "", nil, err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil, os.ErrNotExist
		}
		dir = parent
	}
}

func ProjectNameFromDir(dir string) string {
	base := filepath.Base(dir)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "project"
	}
	return base
}
