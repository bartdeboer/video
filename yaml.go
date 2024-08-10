package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadYaml function to handle the YAML loading logic
func LoadYaml(initial interface{}) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %v", err)
	}

	// File paths to check
	paths := []string{
		filepath.Join(homeDir, ".video.yaml"),
		filepath.Join(".", ".video.yaml"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil { // File exists
			if err := parseYAMLFile(path, initial); err != nil {
				return fmt.Errorf("error parsing YAML file (%s): %v", path, err)
			}
			break // Stop after the first successful load
		}
	}

	return nil
}

// Helper function to parse YAML file
func parseYAMLFile(path string, initial interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, initial)
	if err != nil {
		return fmt.Errorf("error unmarshalling YAML: %v", err)
	}

	return nil
}
