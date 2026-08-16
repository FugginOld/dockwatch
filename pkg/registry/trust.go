package registry

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"

	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/docker/cli/cli/config/types"
	"github.com/fugginold/dockwatch/pkg/registry/helpers"
	log "github.com/sirupsen/logrus"
)

// EncodedAuth returns an encoded auth config for the given registry
// loaded from environment variables or docker config
// as available in that order
func EncodedAuth(ref string) (string, error) {
	auth, err := EncodedEnvAuth(ref)
	if err != nil {
		auth, err = EncodedConfigAuth(ref)
	}
	return auth, err
}

// EnvCredentialsAreUnscoped reports whether REPO_USER/REPO_PASS are set without a
// REPO_HOST to scope them, which means they are offered to whatever registry each
// watched image lives on.
//
// Checked at startup rather than lazily on first use: with REPO_HOST optional, the
// resulting warning is the only mitigation for the default configuration, so it
// needs to be in the first screen of logs rather than appearing mid-scan.
func EnvCredentialsAreUnscoped() bool {
	return os.Getenv("REPO_USER") != "" &&
		os.Getenv("REPO_PASS") != "" &&
		os.Getenv("REPO_HOST") == ""
}

// EncodedEnvAuth returns an encoded auth config for the given image reference
// loaded from environment variables.
//
// REPO_HOST, when set, restricts those credentials to a single registry. Without
// it they are offered to whatever registry the image happens to live on, so a
// single watched image on a hostile or typosquatted registry is enough to collect
// them -- which is why docker's own credential lookup is per-registry.
//
// Returns an error if the variables are unset, or if they are scoped to a
// different registry than the image, so the caller falls back to the docker config.
func EncodedEnvAuth(imageRef string) (string, error) {
	username := os.Getenv("REPO_USER")
	password := os.Getenv("REPO_PASS")
	if username == "" || password == "" {
		return "", errors.New("registry auth environment variables (REPO_USER, REPO_PASS) not set")
	}

	if repoHost := os.Getenv("REPO_HOST"); repoHost != "" {
		server, err := helpers.GetRegistryAddress(imageRef)
		if err != nil {
			return "", err
		}
		// Compare normalized: an operator reaches for "docker.io" or the
		// "https://index.docker.io/v1/" key from config.json long before the
		// normalized address, and rejecting those reads as a broken credential.
		if !strings.EqualFold(helpers.NormalizeRegistryHost(server), helpers.NormalizeRegistryHost(repoHost)) {
			log.WithFields(log.Fields{"registry": server, "repo_host": repoHost}).
				Debug("Environment credentials are scoped to another registry, not using them")
			return "", errors.New("environment credentials are scoped to a different registry")
		}
	}

	auth := types.AuthConfig{
		Username: username,
		Password: password,
	}

	log.Debugf("Loaded auth credentials for registry user %s from environment", auth.Username)
	// CREDENTIAL: deliberately not logged. Uncommenting the line below writes a
	// live secret to the log; do it only against a throwaway credential.
	// log.Tracef("Using auth password %s", auth.Password)

	return EncodeAuth(auth)
}

// EncodedConfigAuth returns an encoded auth config for the given registry
// loaded from the docker config
// Returns an empty string if credentials cannot be found for the referenced server
// The docker config must be mounted on the container
func EncodedConfigAuth(imageRef string) (string, error) {
	server, err := helpers.GetRegistryAddress(imageRef)
	if err != nil {
		log.Errorf("Could not get registry from image ref %s", imageRef)
		return "", err
	}

	configDir := os.Getenv("DOCKER_CONFIG")
	if configDir == "" {
		configDir = "/"
	}
	configFile, err := cliconfig.Load(configDir)
	if err != nil {
		log.Errorf("Unable to find default config file: %s", err)
		return "", err
	}
	credStore := CredentialsStore(*configFile)
	auth, err := credStore.Get(server) // returns (types.AuthConfig{}) if server not in credStore
	if err != nil {
		// Discarded, a credential helper that is missing, broken or refusing to
		// unlock was indistinguishable from "this registry has no credentials".
		log.WithError(err).WithField("registry", server).
			Warn("Credential store returned an error; continuing without its credentials")
	}

	if auth == (types.AuthConfig{}) {
		log.WithField("config_file", configFile.Filename).Debugf("No credentials for %s found", server)
		return "", nil
	}
	log.Debugf("Loaded auth credentials for user %s, on registry %s, from file %s", auth.Username, server, configFile.Filename)
	// CREDENTIAL: deliberately not logged. Uncommenting the line below writes a
	// live secret to the log; do it only against a throwaway credential.
	// log.Tracef("Using auth password %s", auth.Password)
	return EncodeAuth(auth)
}

// CredentialsStore returns a new credentials store based
// on the settings provided in the configuration file.
func CredentialsStore(configFile configfile.ConfigFile) credentials.Store {
	if configFile.CredentialsStore != "" {
		return credentials.NewNativeStore(&configFile, configFile.CredentialsStore)
	}
	return credentials.NewFileStore(&configFile)
}

// EncodeAuth Base64 encode an AuthConfig struct for transmission over HTTP
func EncodeAuth(authConfig types.AuthConfig) (string, error) {
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}
