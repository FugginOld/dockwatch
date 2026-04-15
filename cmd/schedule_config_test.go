package cmd

import (
	"path/filepath"
	"testing"
)

func TestScheduleConfigRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dockwatch.json")
	want := "@every 6h"

	if err := saveScheduleConfig(path, want); err != nil {
		t.Fatalf("save schedule config: %v", err)
	}

	got, found, err := loadScheduleConfig(path)
	if err != nil {
		t.Fatalf("load schedule config: %v", err)
	}
	if !found {
		t.Fatal("expected schedule config to be found")
	}
	if got != want {
		t.Fatalf("schedule = %q, want %q", got, want)
	}
}

func TestLoadScheduleConfigMissingFile(t *testing.T) {
	got, found, err := loadScheduleConfig(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("load missing schedule config: %v", err)
	}
	if found {
		t.Fatal("expected missing schedule config not to be found")
	}
	if got != "" {
		t.Fatalf("schedule = %q, want empty", got)
	}
}
