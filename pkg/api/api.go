package api

import (
	"crypto/hmac"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const tokenMissingMsg = "api token is empty or has not been set. exiting"

const defaultAddr = ":8080"

// The zero-value http.Server applies no deadlines at all, so a client that opens a
// connection and never finishes its request holds a goroutine and a file descriptor
// until the process dies. Since this endpoint triggers container updates, starving
// it of those is enough to stop dockwatch updating anything.
//
// There is deliberately no WriteTimeout: the update handler runs the whole scan
// inline -- pull, stop, recreate, plus any wait behind an in-flight scheduled scan --
// so any write deadline is a cap on how long an update may take. Go sets that
// deadline before calling the handler, so exceeding it means the response can never
// be flushed and the caller sees a reset connection for an update that succeeded.
const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

// API is the http server responsible for serving the HTTP API endpoints
type API struct {
	Token string
	// Addr is the address to listen on. Defaults to defaultAddr.
	Addr        string
	mux         *http.ServeMux
	hasHandlers bool
}

// New is a factory function creating a new API instance
func New(token string) *API {
	return &API{
		Token:       token,
		Addr:        defaultAddr,
		mux:         http.NewServeMux(),
		hasHandlers: false,
	}
}

// RequireToken is wrapper around http.HandlerFunc that checks token validity
func (api *API) RequireToken(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		want := fmt.Sprintf("Bearer %s", api.Token)
		if !hmac.Equal([]byte(auth), []byte(want)) {
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

	addr := api.Addr
	if addr == "" {
		addr = defaultAddr
	}

	// Bind here rather than inside the goroutine: a bind failure -- the port already
	// being in use is the common one -- would otherwise never reach the caller, who
	// would carry on as though the API were serving.
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Warn("HTTP API is running without TLS. Ensure this endpoint is behind a TLS-terminating proxy in production.")

	if block {
		return api.serve(listener)
	}

	go func() {
		if err := api.serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Error("HTTP API stopped serving")
		}
	}()
	return nil
}

func (api *API) newServer() *http.Server {
	return &http.Server{
		Handler:           api.mux,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func (api *API) serve(listener net.Listener) error {
	return api.newServer().Serve(listener)
}
