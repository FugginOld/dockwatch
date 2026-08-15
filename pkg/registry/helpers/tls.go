package helpers

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	registryClient     *http.Client
	registryClientOnce sync.Once
)

// ShouldSkipRegistryTLSVerify returns true if the DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY
// environment variable is set to a truthy value.
func ShouldSkipRegistryTLSVerify() bool {
	raw := os.Getenv("DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY")
	if raw == "" {
		return false
	}

	skip, err := strconv.ParseBool(raw)
	if err != nil {
		logrus.WithError(err).WithField("DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY", raw).Warn("Invalid boolean value for registry TLS skip verify setting")
		return false
	}

	return skip
}

// NewHTTPClient returns a shared *http.Client with TLS verification configured
// according to ShouldSkipRegistryTLSVerify. The client is initialized once on
// first call and reused on all subsequent calls, so the TLS setting is read
// only at first use and the security warning is emitted at most once.
func NewHTTPClient() *http.Client {
	registryClientOnce.Do(func() {
		skipTLSVerify := ShouldSkipRegistryTLSVerify()

		if skipTLSVerify {
			logrus.Warn("Registry TLS certificate verification is disabled. This should only be used for testing or trusted private networks.")
		}

		tr := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			// The dial timeout above stops at connect. A registry that completes the
			// handshake and then never sends a response line would otherwise park the
			// update goroutine forever.
			ResponseHeaderTimeout: 30 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: skipTLSVerify}, //nolint:gosec // controlled by explicit env var
		}

		registryClient = &http.Client{
			Transport: tr,
			// Whole-request ceiling, covering a slow body as well as slow headers.
			// Generous enough for a manifest HEAD or a token fetch, which is all this
			// client is used for -- image layers go through the docker daemon.
			Timeout: 60 * time.Second,
			// net/http strips the Authorization header when a redirect crosses to a
			// different host, but it compares hosts only, not schemes. A registry that
			// redirects https -> http on the same host would therefore be handed the
			// bearer token in cleartext.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 0 && via[0].URL.Scheme == "https" && req.URL.Scheme != "https" {
					return fmt.Errorf("refusing redirect from https to %s: registry credentials would be sent in cleartext", req.URL.Scheme)
				}
				if len(via) >= 10 {
					return errors.New("stopped after 10 redirects")
				}
				return nil
			},
		}
	})

	return registryClient
}
