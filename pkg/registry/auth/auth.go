package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/fugginold/dockwatch/pkg/registry/helpers"
	"github.com/fugginold/dockwatch/pkg/types"
	ref "github.com/distribution/reference"
	"github.com/sirupsen/logrus"
)

// ChallengeHeader is the HTTP Header containing challenge instructions
const ChallengeHeader = "WWW-Authenticate"

// GetToken fetches a token for the registry hosting the provided image
func GetToken(ctx context.Context, container types.Container, registryAuth string) (string, error) {
	normalizedRef, err := ref.ParseNormalizedNamed(container.ImageName())
	if err != nil {
		return "", err
	}

	URL := GetChallengeURL(normalizedRef)
	logrus.WithField("URL", URL.String()).Debug("Built challenge URL")

	var req *http.Request
	if req, err = GetChallengeRequest(ctx, URL); err != nil {
		return "", err
	}

	client := helpers.NewHTTPClient()
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return "", err
	}
	defer res.Body.Close()
	v := res.Header.Get(ChallengeHeader)

	logrus.WithFields(logrus.Fields{
		"status": res.Status,
		"header": v,
	}).Debug("Got response to challenge request")

	return tokenForChallenge(ctx, v, res.StatusCode, normalizedRef, registryAuth)
}

// tokenForChallenge decides what Authorization header, if any, the challenge calls
// for. Split out from GetToken because GetToken can only talk to a real registry
// host, which leaves this decision untestable.
func tokenForChallenge(ctx context.Context, header string, status int, imageRef ref.Named, registryAuth string) (string, error) {
	// Match the scheme case-insensitively but hand the challenge on with its
	// original case: the realm and service values inside it are case-sensitive.
	challenge := strings.ToLower(header)
	if strings.HasPrefix(challenge, "basic") {
		if registryAuth == "" {
			return "", fmt.Errorf("no credentials available")
		}

		return fmt.Sprintf("Basic %s", registryAuth), nil
	}
	if strings.HasPrefix(challenge, "bearer") {
		return GetBearerHeader(ctx, header, imageRef, registryAuth)
	}

	// No challenge at all means the registry served /v2/ without asking for auth,
	// so reads are anonymous. Treating that as an unsupported challenge type sent
	// every poll down the full-pull path instead.
	if strings.TrimSpace(header) == "" && status >= 200 && status <= 299 {
		logrus.Debug("Registry issued no challenge; proceeding without a token")
		return "", nil
	}

	return "", errors.New("unsupported challenge type from registry")
}

// GetChallengeRequest creates a request for getting challenge instructions
func GetChallengeRequest(ctx context.Context, URL url.URL) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", URL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Dockwatch (Docker)")
	return req, nil
}

// GetBearerHeader tries to fetch a bearer token from the registry based on the challenge instructions
func GetBearerHeader(ctx context.Context, challenge string, imageRef ref.Named, registryAuth string) (string, error) {
	client := helpers.NewHTTPClient()
	authURL, err := GetAuthURL(challenge, imageRef)

	if err != nil {
		return "", err
	}

	var r *http.Request
	if r, err = http.NewRequestWithContext(ctx, "GET", authURL.String(), nil); err != nil {
		return "", err
	}

	if registryAuth != "" {
		logrus.Debug("Credentials found.")
		// CREDENTIAL: Uncomment to log registry credentials
		// logrus.Tracef("Credentials: %v", registryAuth)
		r.Header.Add("Authorization", fmt.Sprintf("Basic %s", registryAuth))
	} else {
		logrus.Debug("No credentials found.")
	}

	var authResponse *http.Response
	if authResponse, err = client.Do(r); err != nil {
		return "", err
	}

	return bearerFromResponse(authResponse)
}

// bearerFromResponse turns a token endpoint response into an Authorization header.
//
// Every check here was missing: the body was never closed, so the descriptor leaked
// once per container per poll interval and accumulated for the life of the process;
// the read error was discarded; and the status was never looked at, so a 401
// unmarshalled cleanly into an empty TokenResponse and the function returned the
// literal "Bearer " as a success. The real rejection was swallowed and resurfaced
// later as a confusing 401 on the digest request.
func bearerFromResponse(res *http.Response) (string, error) {
	defer res.Body.Close()

	// Bounded: this read now happens before the status check, so a hostile token
	// endpoint reaches it on every response. 1 MiB is orders of magnitude more than
	// any real token document.
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("registry token endpoint responded with %q", res.Status)
	}

	tokenResponse := &types.TokenResponse{}
	if err := json.Unmarshal(body, tokenResponse); err != nil {
		return "", err
	}

	token := tokenResponse.Bearer()
	if token == "" {
		return "", errors.New("registry token endpoint returned no token")
	}

	return fmt.Sprintf("Bearer %s", token), nil
}

// GetAuthURL from the instructions in the challenge
func GetAuthURL(challenge string, imageRef ref.Named) (*url.URL, error) {
	// Only the scheme name and the parameter keys are case-insensitive; the values
	// are not. Lowercasing the whole header corrupted case-sensitive Artifactory
	// realms and services -- docker-Local became docker-local and the token request
	// 404'd, which showed up only as a full pull on every poll.
	raw := challenge
	if len(raw) >= len("bearer") && strings.EqualFold(raw[:len("bearer")], "bearer") {
		raw = raw[len("bearer"):]
	}

	pairs := strings.Split(raw, ",")
	values := make(map[string]string, len(pairs))

	for _, pair := range pairs {
		trimmed := strings.Trim(pair, " ")
		if key, val, ok := strings.Cut(trimmed, "="); ok {
			values[strings.ToLower(strings.TrimSpace(key))] = strings.Trim(val, ` "`)
		}
	}
	logrus.WithFields(logrus.Fields{
		"realm":   values["realm"],
		"service": values["service"],
	}).Debug("Checking challenge header content")
	if values["realm"] == "" || values["service"] == "" {

		return nil, fmt.Errorf("challenge header did not include all values needed to construct an auth url")
	}

	authURL, err := url.Parse(values["realm"])
	if err != nil {
		return nil, fmt.Errorf("challenge header realm %q is not a valid URL: %w", values["realm"], err)
	}

	// The registry picks the realm, so it picks where the credentials in the next
	// request are sent and over what transport. Requiring https with a host means a
	// registry cannot downgrade us to sending a Basic header in cleartext.
	if authURL.Scheme != "https" || authURL.Host == "" {
		return nil, fmt.Errorf("challenge header realm %q must be an https URL with a host", values["realm"])
	}

	q := authURL.Query()
	q.Add("service", values["service"])

	scopeImage := ref.Path(imageRef)

	scope := fmt.Sprintf("repository:%s:pull", scopeImage)
	logrus.WithFields(logrus.Fields{"scope": scope, "image": imageRef.Name()}).Debug("Setting scope for auth token")
	q.Add("scope", scope)

	authURL.RawQuery = q.Encode()
	return authURL, nil
}

// GetChallengeURL returns the URL to check auth requirements
// for access to a given image
func GetChallengeURL(imageRef ref.Named) url.URL {
	host, _ := helpers.GetRegistryAddress(imageRef.Name())

	URL := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/v2/",
	}
	return URL
}
