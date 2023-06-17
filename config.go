package main

import (
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"os"
)

type script struct {
	ID         uuid.UUID `yaml:"id"`
	Path       string    `yaml:"path"`
	Token      string    `yaml:"token,omitempty"`
	Concurrent bool      `yaml:"concurrent"`
	Shell      string    `yaml:"shell"`
}

type configuration struct {
	DefaultToken string   `yaml:"default_token"`
	Scripts      []script `yaml:"scripts"`
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

	return c, nil
}
