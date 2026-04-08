package helpers

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
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

// NewHTTPClient returns an *http.Client with TLS verification configured
// according to ShouldSkipRegistryTLSVerify.
func NewHTTPClient() *http.Client {
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
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: skipTLSVerify}, //nolint:gosec // controlled by explicit env var
	}

	return &http.Client{Transport: tr}
}
