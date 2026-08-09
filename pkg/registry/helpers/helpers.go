package helpers

import (
	"strings"

	"github.com/distribution/reference"
)

// domains for Docker Hub, the default registry
const (
	DefaultRegistryDomain       = "docker.io"
	DefaultRegistryHost         = "index.docker.io"
	LegacyDefaultRegistryDomain = "index.docker.io"
)

// GetRegistryAddress parses an image name
// and returns the address of the specified registry
func GetRegistryAddress(imageRef string) (string, error) {
	normalizedRef, err := reference.ParseNormalizedNamed(imageRef)
	if err != nil {
		return "", err
	}

	address := reference.Domain(normalizedRef)

	if address == DefaultRegistryDomain {
		address = DefaultRegistryHost
	}
	return address, nil
}

// NormalizeRegistryHost maps the forms a registry host is commonly written in onto
// the address GetRegistryAddress returns, so the two can be compared.
//
// It accepts a bare domain, the docker.io alias for Docker Hub, and the
// "https://index.docker.io/v1/" key that docker login writes into config.json.
// A port is significant and is preserved: a different port is a different registry.
func NormalizeRegistryHost(host string) string {
	host = strings.TrimSpace(host)

	if _, after, found := strings.Cut(host, "://"); found {
		host = after
	}
	host, _, _ = strings.Cut(host, "/")

	if strings.EqualFold(host, DefaultRegistryDomain) || strings.EqualFold(host, LegacyDefaultRegistryDomain) {
		return DefaultRegistryHost
	}
	return host
}
