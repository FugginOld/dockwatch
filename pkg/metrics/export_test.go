package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// ResetForTest rebuilds the singleton against a fresh collector registry, so a test
// can exercise the construction path itself rather than a singleton some earlier
// test in the package already built.
func ResetForTest() {
	registerer = prometheus.NewRegistry()
	metrics = nil
	defaultOnce = sync.Once{}
}
