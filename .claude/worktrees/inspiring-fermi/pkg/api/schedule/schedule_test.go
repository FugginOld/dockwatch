package schedule

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandleGetSchedule(t *testing.T) {
	h := New(
		func() string { return "@every 1h" },
		func() time.Time { return time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC) },
		func(string) (time.Time, error) { return time.Time{}, nil },
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/schedule", nil)

	h.Handle(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got == "" || !strings.Contains(got, "@every 1h") {
		t.Fatalf("expected schedule in response body, got %q", got)
	}
}

func TestHandleUpdateSchedule(t *testing.T) {
	updated := ""
	h := New(
		func() string { return "@every 1h" },
		func() time.Time { return time.Time{} },
		func(spec string) (time.Time, error) {
			updated = spec
			return time.Date(2026, 4, 6, 13, 0, 0, 0, time.UTC), nil
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/schedule?schedule=@every%2030m", nil)

	h.Handle(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if updated != "@every 30m" {
		t.Fatalf("expected updated schedule @every 30m, got %q", updated)
	}
}

func TestHandleUpdateScheduleMissingParam(t *testing.T) {
	h := New(
		func() string { return "@every 1h" },
		func() time.Time { return time.Time{} },
		func(string) (time.Time, error) { return time.Time{}, nil },
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/schedule", nil)

	h.Handle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandleUpdateScheduleInvalidSpec(t *testing.T) {
	h := New(
		func() string { return "@every 1h" },
		func() time.Time { return time.Time{} },
		func(string) (time.Time, error) { return time.Time{}, errors.New("invalid") },
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/schedule?cron=invalid", nil)

	h.Handle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
