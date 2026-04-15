package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type scheduleConfig struct {
	Schedule string `json:"schedule"`
}

func loadScheduleConfig(path string) (string, bool, error) {
	if strings.TrimSpace(path) == "" {
		return "", false, nil
	}

	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	var config scheduleConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return "", false, err
	}

	schedule := strings.TrimSpace(config.Schedule)
	return schedule, schedule != "", nil
}

func saveScheduleConfig(path string, schedule string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(scheduleConfig{Schedule: schedule}, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')

	return os.WriteFile(path, content, 0o644)
}
