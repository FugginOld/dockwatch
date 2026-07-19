package daemon

import "testing"

func TestDecideMode(t *testing.T) {
	cases := []struct {
		name         string
		opts         Options
		wantPeriodic bool
		wantBlock    bool
	}{
		{"default", Options{}, true, false},
		{"update api only blocks", Options{EnableUpdateAPI: true}, false, true},
		{"update api with periodic polls", Options{EnableUpdateAPI: true, UnblockHTTPAPI: true}, true, false},
		{"update api while interactive does not block", Options{EnableUpdateAPI: true, Interactive: true}, false, false},
		{"metrics api alone stays periodic", Options{EnableMetricsAPI: true}, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := decideMode(tc.opts)
			if m.periodicEnabled != tc.wantPeriodic || m.apiShouldBlock != tc.wantBlock {
				t.Fatalf("decideMode() = {periodic:%v block:%v}, want {periodic:%v block:%v}",
					m.periodicEnabled, m.apiShouldBlock, tc.wantPeriodic, tc.wantBlock)
			}
		})
	}
}
