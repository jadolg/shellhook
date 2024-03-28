package main

import (
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"os"
)

type script struct {
	ID          uuid.UUID     `yaml:"id"`
	Path        string        `yaml:"path,omitempty"`
	Inline      string        `yaml:"inline,omitempty"`
	Token       string        `yaml:"token,omitempty"`
	Concurrent  bool          `yaml:"concurrent"`
	Shell       string        `yaml:"shell"`
	User        string        `yaml:"user"`
	Environment []environment `yaml:"environment"`
}

type environment struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

func (s script) isValid() bool {
	return (s.Path != "" && s.Inline == "") || (s.Path == "" && s.Inline != "")
}

type configuration struct {
	DefaultToken string        `yaml:"default_token"`
	Scripts      []script      `yaml:"scripts"`
	Environment  []environment `yaml:"environment"`
}

func getConfig(configFile string) (configuration, error) {
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return configuration{}, err
	}
	c := configuration{}
	err = yaml.Unmarshal(yamlFile, &c)

	if err != nil {
		return configuration{}, err
	}

	for _, s := range c.Scripts {
		if !s.isValid() {
			return configuration{}, fmt.Errorf("invalid script: %v", s)
		}
	}

	return c, nil
}
