package digest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/fugginold/dockwatch/internal/meta"
	"github.com/fugginold/dockwatch/pkg/registry/auth"
	"github.com/fugginold/dockwatch/pkg/registry/helpers"
	"github.com/fugginold/dockwatch/pkg/registry/manifest"
	"github.com/fugginold/dockwatch/pkg/types"
	"github.com/sirupsen/logrus"
)

// ContentDigestHeader is the key for the key-value pair containing the digest header
const ContentDigestHeader = "Docker-Content-Digest"

// CompareDigest ...
func CompareDigest(ctx context.Context, container types.Container, registryAuth string) (bool, error) {
	if !container.HasImageInfo() {
		return false, errors.New("container image info missing")
	}

	var digest string

	registryAuth = TransformAuth(registryAuth)
	token, err := auth.GetToken(ctx, container, registryAuth)
	if err != nil {
		return false, err
	}

	digestURL, err := manifest.BuildManifestURL(container)
	if err != nil {
		return false, err
	}

	if digest, err = GetDigest(ctx, digestURL, token); err != nil {
		return false, err
	}

	logrus.WithField("remote", digest).Debug("Found a remote digest to compare with")

	for _, dig := range container.ImageInfo().RepoDigests {
		parts := strings.SplitN(dig, "@", 2)
		if len(parts) != 2 || parts[1] == "" {
			logrus.WithField("repoDigest", dig).Warn("Skipping invalid local digest entry")
			continue
		}

		localDigest := parts[1]
		fields := logrus.Fields{"local": localDigest, "remote": digest}
		logrus.WithFields(fields).Debug("Comparing")

		if localDigest == digest {
			logrus.Debug("Found a match")
			return true, nil
		}
	}

	return false, nil
}

// decodeAuth accepts either base64 alphabet.
//
// EncodeAuth writes with URLEncoding while this decoded with StdEncoding. The two
// differ only in the last two characters, so the mismatch stayed invisible until a
// credential happened to encode a byte as one of them -- and then the decode failed,
// TransformAuth returned its input untouched, and the registry was handed a
// base64-of-JSON blob where it expected user:pass. Accepting both means neither a
// future producer nor a docker config written by another tool can reopen it.
func decodeAuth(s string) ([]byte, error) {
	if b, err := base64.URLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.StdEncoding.DecodeString(s)
}

// TransformAuth from a base64 encoded json object to base64 encoded string
func TransformAuth(registryAuth string) string {
	b, err := decodeAuth(registryAuth)
	if err != nil {
		return registryAuth
	}

	credentials := &types.RegistryCredentials{}
	if err := json.Unmarshal(b, credentials); err != nil {
		return registryAuth
	}

	if credentials.Username != "" && credentials.Password != "" {
		ba := []byte(fmt.Sprintf("%s:%s", credentials.Username, credentials.Password))
		registryAuth = base64.StdEncoding.EncodeToString(ba)
	}

	return registryAuth
}

// GetDigest from registry using a HEAD request to prevent rate limiting
func GetDigest(ctx context.Context, url string, token string) (string, error) {
	client := helpers.NewHTTPClient()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", meta.UserAgent)

	// An empty token is legitimate for a registry that serves reads anonymously;
	// GetToken only returns one when the registry issued no challenge at all.
	if token != "" {
		// CREDENTIAL: deliberately not logged. Uncommenting the line below writes a
		// live secret to the log; do it only against a throwaway credential.
		// logrus.WithField("token", token).Trace("Setting request token")
		req.Header.Add("Authorization", token)
	}

	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v1+json")
	req.Header.Add("Accept", "application/vnd.oci.image.index.v1+json")

	logrus.WithField("url", url).Debug("Doing a HEAD request to fetch a digest")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		wwwAuthHeader := res.Header.Get("www-authenticate")
		if wwwAuthHeader == "" {
			wwwAuthHeader = "not present"
		}
		return "", fmt.Errorf("registry responded to head request with %q, auth: %q", res.Status, wwwAuthHeader)
	}

	// Artifactory and nginx-fronted registries sometimes strip this header on HEAD.
	// Returning "" as a valid digest made the comparison never match, so every poll
	// pulled the whole image -- the exact rate-limit consumption the HEAD exists to
	// avoid -- and logged nothing, because that is the normal "digests differ" path.
	remoteDigest := res.Header.Get(ContentDigestHeader)
	if remoteDigest == "" {
		return "", fmt.Errorf("registry did not return a %s header", ContentDigestHeader)
	}
	return remoteDigest, nil
}
