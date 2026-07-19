package daemon

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"positive timeout", Config{Timeout: 10 * time.Second}, false},
		{"zero timeout", Config{}, false},
		{"negative timeout", Config{Timeout: -1}, true},
		{"rolling restart with monitor only", Config{RollingRestart: true, MonitorOnly: true}, true},
		{"rolling restart alone", Config{RollingRestart: true}, false},
		{"monitor only alone", Config{MonitorOnly: true}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}
