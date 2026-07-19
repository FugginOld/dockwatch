package update

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Scanner runs guarded container update scans. wait blocks for the
// single-flight guard; when false the run is skipped if one is already
// in progress. The guard itself lives in the Scanner, not this handler.
type Scanner interface {
	Scan(images []string, wait bool)
}

// New is a factory function creating a new Handler instance.
func New(scanner Scanner) *Handler {
	return &Handler{
		scanner: scanner,
		Path:    "/v1/update",
	}
}

// Handler is an API handler used for triggering container update scans.
type Handler struct {
	scanner Scanner
	Path    string
}

// Handle parses the requested images and delegates the guarded scan to the
// Scanner. A targeted image update waits for the guard; a full scan is
// skipped when one is already running.
func (handle *Handler) Handle(_ http.ResponseWriter, r *http.Request) {
	log.Info("Updates triggered by HTTP API request.")

	var images []string
	if imageQueries, found := r.URL.Query()["image"]; found {
		for _, image := range imageQueries {
			images = append(images, strings.Split(image, ",")...)
		}
	}

	handle.scanner.Scan(images, len(images) > 0)
}
