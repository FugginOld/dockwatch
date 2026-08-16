package metrics_test

import (
	"testing"
	"time"

	"github.com/fugginold/dockwatch/pkg/metrics"
	"github.com/fugginold/dockwatch/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// waitForEmpty polls m.QueueIsEmpty() until it returns true or a deadline is exceeded.
func waitForEmpty(t *testing.T, m *metrics.Metrics) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if m.QueueIsEmpty() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	// Final check — will fail the test if still not empty
	assert.True(t, m.QueueIsEmpty(), "metrics queue should be empty after draining")
}

// mockReport implements types.Report using prepared slices.
type mockReport struct {
	scanned []types.ContainerReport
	updated []types.ContainerReport
	failed  []types.ContainerReport
	skipped []types.ContainerReport
	stale   []types.ContainerReport
	fresh   []types.ContainerReport
}

func (r *mockReport) Scanned() []types.ContainerReport { return r.scanned }
func (r *mockReport) Updated() []types.ContainerReport { return r.updated }
func (r *mockReport) Failed() []types.ContainerReport  { return r.failed }
func (r *mockReport) Skipped() []types.ContainerReport { return r.skipped }
func (r *mockReport) Stale() []types.ContainerReport   { return r.stale }
func (r *mockReport) Fresh() []types.ContainerReport   { return r.fresh }
func (r *mockReport) All() []types.ContainerReport     { return nil }

func TestNewMetric_Counts(t *testing.T) {
	report := &mockReport{
		scanned: make([]types.ContainerReport, 3),
		updated: make([]types.ContainerReport, 1),
		failed:  make([]types.ContainerReport, 2),
		stale:   make([]types.ContainerReport, 0),
	}

	m := metrics.NewMetric(report)
	require.NotNil(t, m)

	assert.Equal(t, 3, m.Scanned)
	// Updated = Updated + Stale (backwards compat)
	assert.Equal(t, 1, m.Updated)
	assert.Equal(t, 2, m.Failed)
}

func TestNewMetric_UpdatedIncludesStale(t *testing.T) {
	report := &mockReport{
		updated: make([]types.ContainerReport, 2),
		stale:   make([]types.ContainerReport, 3),
	}

	m := metrics.NewMetric(report)
	assert.Equal(t, 5, m.Updated)
}

func TestNewMetric_Empty(t *testing.T) {
	report := &mockReport{}
	m := metrics.NewMetric(report)

	assert.Equal(t, 0, m.Scanned)
	assert.Equal(t, 0, m.Updated)
	assert.Equal(t, 0, m.Failed)
}

func TestMetrics_QueueIsEmpty_InitiallyEmpty(t *testing.T) {
	m := metrics.Default()
	waitForEmpty(t, m)
	assert.True(t, m.QueueIsEmpty())
}

func TestMetrics_Register_And_QueueIsEmpty(t *testing.T) {
	m := metrics.Default()
	metric := &metrics.Metric{Scanned: 5, Updated: 2, Failed: 1}
	m.Register(metric)
	waitForEmpty(t, m)
	assert.True(t, m.QueueIsEmpty())
}

func TestRegisterScan_NilMetric(t *testing.T) {
	// Registering nil should represent a skipped scan and not panic.
	assert.NotPanics(t, func() {
		metrics.RegisterScan(nil)
	})
	waitForEmpty(t, metrics.Default())
}

func TestRegisterScan_ValidMetric(t *testing.T) {
	m := &metrics.Metric{Scanned: 10, Updated: 3, Failed: 0}
	assert.NotPanics(t, func() {
		metrics.RegisterScan(m)
	})
	waitForEmpty(t, metrics.Default())
}

func TestMetrics_Default_ReturnsSameInstance(t *testing.T) {
	m1 := metrics.Default()
	m2 := metrics.Default()
	assert.Same(t, m1, m2)
}

// Default is reached from the cron goroutine and from a finishing scan's goroutine
// at the same time. Without synchronization both can pass the nil check, and the
// second promauto.NewGauge hits MustRegister -> "duplicate metrics collector
// registration attempted", which panics the daemon.
func TestDefaultIsASingletonUnderConcurrency(t *testing.T) {
	const goroutines = 50

	// Start from an unbuilt singleton and a fresh registry: another test in this
	// package has already called Default(), so without this every goroutine below
	// just gets the existing instance and the race is never exercised.
	metrics.ResetForTest()

	start := make(chan struct{})
	got := make(chan *metrics.Metrics, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					got <- nil
					t.Errorf("Default panicked: %v", r)
				}
			}()
			<-start
			got <- metrics.Default()
		}()
	}
	close(start)

	first := <-got
	for i := 1; i < goroutines; i++ {
		if m := <-got; m != first {
			t.Fatalf("Default returned a second instance; the metrics singleton was built twice")
		}
	}
}
