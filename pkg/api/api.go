package api

import (
	"errors"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const tokenMissingMsg = "api token is empty or has not been set. exiting"

// API is the http server responsible for serving the HTTP API endpoints
type API struct {
	Token       string
	mux         *http.ServeMux
	hasHandlers bool
}

// New is a factory function creating a new API instance
func New(token string) *API {
	return &API{
		Token:       token,
		mux:         http.NewServeMux(),
		hasHandlers: false,
	}
}

// RequireToken is wrapper around http.HandlerFunc that checks token validity
func (api *API) RequireToken(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		want := fmt.Sprintf("Bearer %s", api.Token)
		if auth != want {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Debug("Valid token found.")
		fn(w, r)
	}
}

// RegisterFunc is a wrapper around http.HandleFunc that also sets the flag used to determine whether to launch the API
func (api *API) RegisterFunc(path string, fn http.HandlerFunc) {
	api.hasHandlers = true
	api.mux.HandleFunc(path, api.RequireToken(fn))
}

// RegisterHandler is a wrapper around http.Handler that also sets the flag used to determine whether to launch the API
func (api *API) RegisterHandler(path string, handler http.Handler) {
	api.hasHandlers = true
	api.mux.Handle(path, api.RequireToken(handler.ServeHTTP))
}

// Start the API and serve over HTTP. Requires an API Token to be set.
func (api *API) Start(block bool) error {

	if !api.hasHandlers {
		log.Debug("Dockwatch HTTP API skipped.")
		return nil
	}

	if api.Token == "" {
		return errors.New(tokenMissingMsg)
	}

	if block {
		return api.runHTTPServer()
	} else {
		go func() {
			if err := api.runHTTPServer(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.WithError(err).Error("failed to start HTTP API")
			}
		}()
	}
	return nil
}

func (api *API) runHTTPServer() error {
	return http.ListenAndServe(":8080", api.mux)
}
