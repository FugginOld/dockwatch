package schedule

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Handler is an API handler used for reading/updating the active periodic schedule.
type Handler struct {
	Path      string
	getSpecFn func() string
	getNextFn func() time.Time
	setSpecFn func(string) (time.Time, error)
}

type response struct {
	Schedule string `json:"schedule"`
	NextRun  string `json:"nextRun,omitempty"`
}

// New is a factory function creating a new Handler instance.
func New(
	getSpecFn func() string,
	getNextFn func() time.Time,
	setSpecFn func(string) (time.Time, error),
) *Handler {
	return &Handler{
		Path:      "/v1/schedule",
		getSpecFn: getSpecFn,
		getNextFn: getNextFn,
		setSpecFn: setSpecFn,
	}
}

// Handle supports:
// - GET: Returns current schedule and next run time.
// - POST/PUT: Updates the active schedule using query parameter schedule or cron.
func (handle *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		writeResponse(w, responseFromFns(handle.getSpecFn, handle.getNextFn), http.StatusOK)
		return
	case http.MethodPost, http.MethodPut:
		spec := strings.TrimSpace(r.URL.Query().Get("schedule"))
		if spec == "" {
			spec = strings.TrimSpace(r.URL.Query().Get("cron"))
		}
		if spec == "" {
			http.Error(w, `{"error":"missing schedule query parameter"}`, http.StatusBadRequest)
			return
		}

		nextRun, err := handle.setSpecFn(spec)
		if err != nil {
			log.WithError(err).Warn("Rejected invalid schedule update")
			http.Error(w, `{"error":"invalid schedule"}`, http.StatusBadRequest)
			return
		}

		writeResponse(w, response{Schedule: spec, NextRun: nextRun.Format(time.RFC3339)}, http.StatusOK)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func responseFromFns(getSpecFn func() string, getNextFn func() time.Time) response {
	nextRun := getNextFn()
	out := response{Schedule: getSpecFn()}
	if !nextRun.IsZero() {
		out.NextRun = nextRun.Format(time.RFC3339)
	}
	return out
}

func writeResponse(w http.ResponseWriter, payload response, statusCode int) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
