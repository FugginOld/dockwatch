package auth_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fugginold/dockwatch/internal/actions/mocks"
	"github.com/fugginold/dockwatch/pkg/registry/auth"

	ref "github.com/distribution/reference"
	wtTypes "github.com/fugginold/dockwatch/pkg/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Registry Auth Suite")
}
func SkipIfCredentialsEmpty(credentials *wtTypes.RegistryCredentials, fn func()) func() {
	if credentials.Username == "" {
		return func() {
			Skip("Username missing. Skipping integration test")
		}
	} else if credentials.Password == "" {
		return func() {
			Skip("Password missing. Skipping integration test")
		}
	} else {
		return fn
	}
}

var GHCRCredentials = &wtTypes.RegistryCredentials{
	Username: os.Getenv("CI_INTEGRATION_TEST_REGISTRY_GH_USERNAME"),
	Password: os.Getenv("CI_INTEGRATION_TEST_REGISTRY_GH_PASSWORD"),
}

var _ = Describe("the auth module", func() {
	mockId := "mock-id"
	mockName := "mock-container"
	mockImage := "ghcr.io/k6io/operator:latest"
	mockCreated := time.Now()
	mockDigest := "ghcr.io/k6io/operator@sha256:d68e1e532088964195ad3a0a71526bc2f11a78de0def85629beb75e2265f0547"

	mockContainer := mocks.CreateMockContainerWithDigest(
		mockId,
		mockName,
		mockImage,
		mockCreated,
		mockDigest)

	Describe("GetToken", func() {
		It("should parse the token from the response",
			SkipIfCredentialsEmpty(GHCRCredentials, func() {
				creds := fmt.Sprintf("%s:%s", GHCRCredentials.Username, GHCRCredentials.Password)
				token, err := auth.GetToken(context.Background(), mockContainer, creds)
				Expect(err).NotTo(HaveOccurred())
				Expect(token).NotTo(Equal(""))
			}),
		)
	})

	Describe("GetAuthURL", func() {
		It("should create a valid auth url object based on the challenge header supplied", func() {
			challenge := `bearer realm="https://ghcr.io/token",service="ghcr.io",scope="repository:user/image:pull"`
			imageRef, err := ref.ParseNormalizedNamed("fugginold/dockwatch")
			Expect(err).NotTo(HaveOccurred())
			expected := &url.URL{
				Host:     "ghcr.io",
				Scheme:   "https",
				Path:     "/token",
				RawQuery: "scope=repository%3Afugginold%2Fdockwatch%3Apull&service=ghcr.io",
			}

			URL, err := auth.GetAuthURL(challenge, imageRef)
			Expect(err).NotTo(HaveOccurred())
			Expect(URL).To(Equal(expected))
		})

		When("given an invalid challenge header", func() {
			It("should return an error", func() {
				challenge := `bearer realm="https://ghcr.io/token"`
				imageRef, err := ref.ParseNormalizedNamed("fugginold/dockwatch")
				Expect(err).NotTo(HaveOccurred())
				URL, err := auth.GetAuthURL(challenge, imageRef)
				Expect(err).To(HaveOccurred())
				Expect(URL).To(BeNil())
			})
		})

		When("deriving the auth scope from an image name", func() {
			It("should prepend official dockerhub images with \"library/\"", func() {
				Expect(getScopeFromImageAuthURL("registry")).To(Equal("library/registry"))
				Expect(getScopeFromImageAuthURL("docker.io/registry")).To(Equal("library/registry"))
				Expect(getScopeFromImageAuthURL("index.docker.io/registry")).To(Equal("library/registry"))
			})
			It("should not include vanity hosts\"", func() {
				Expect(getScopeFromImageAuthURL("docker.io/fugginold/dockwatch")).To(Equal("fugginold/dockwatch"))
				Expect(getScopeFromImageAuthURL("index.docker.io/fugginold/dockwatch")).To(Equal("fugginold/dockwatch"))
			})
			It("should not destroy three segment image names\"", func() {
				Expect(getScopeFromImageAuthURL("piksel/fugginold/dockwatch")).To(Equal("piksel/fugginold/dockwatch"))
				Expect(getScopeFromImageAuthURL("ghcr.io/piksel/fugginold/dockwatch")).To(Equal("piksel/fugginold/dockwatch"))
			})
			It("should not prepend library/ to image names if they're not on dockerhub", func() {
				Expect(getScopeFromImageAuthURL("ghcr.io/dockwatch")).To(Equal("dockwatch"))
				Expect(getScopeFromImageAuthURL("ghcr.io/fugginold/dockwatch")).To(Equal("fugginold/dockwatch"))
			})
		})
		It("should not crash when an empty field is received", func() {
			input := `bearer realm="https://ghcr.io/token",service="ghcr.io",scope="repository:user/image:pull",`
			imageRef, err := ref.ParseNormalizedNamed("fugginold/dockwatch")
			Expect(err).NotTo(HaveOccurred())
			res, err := auth.GetAuthURL(input, imageRef)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})
		It("should not crash when a field without a value is received", func() {
			input := `bearer realm="https://ghcr.io/token",service="ghcr.io",scope="repository:user/image:pull",valuelesskey`
			imageRef, err := ref.ParseNormalizedNamed("fugginold/dockwatch")
			Expect(err).NotTo(HaveOccurred())
			res, err := auth.GetAuthURL(input, imageRef)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})
		// The registry chooses the realm, so it also chooses where our credentials go
		// and over what transport. A plaintext realm turns the Basic header into
		// cleartext credentials for anyone on the path.
		It("should reject a realm that is not https", func() {
			imageRef, err := ref.ParseNormalizedNamed("fugginold/dockwatch")
			Expect(err).NotTo(HaveOccurred())

			for _, realm := range []string{
				`http://ghcr.io/token`, // plaintext
				`ftp://ghcr.io/token`,  // not http at all
				`/token`,               // scheme-relative, no host
				`https:///token`,       // https but no host
			} {
				input := `bearer realm="` + realm + `",service="ghcr.io"`
				res, err := auth.GetAuthURL(input, imageRef)
				Expect(err).To(HaveOccurred(), "realm %q should be rejected", realm)
				Expect(res).To(BeNil())
			}
		})
		// A registry controls the realm value, so an unparseable one must be an
		// error rather than a nil dereference: there is no recover() anywhere, so
		// a panic here takes down the daemon and every other container's updates.
		It("should return an error when the realm is not a parseable URL", func() {
			imageRef, err := ref.ParseNormalizedNamed("fugginold/dockwatch")
			Expect(err).NotTo(HaveOccurred())

			for _, realm := range []string{
				`https://ghcr.io:notaport/token`, // invalid port
				`https://ghcr.io/%zz`,            // bad percent-escape
				`://ghcr.io/token`,               // missing scheme
			} {
				input := `bearer realm="` + realm + `",service="ghcr.io"`
				res, err := auth.GetAuthURL(input, imageRef)
				Expect(err).To(HaveOccurred(), "realm %q should be rejected", realm)
				Expect(res).To(BeNil())
			}
		})
	})

	Describe("GetChallengeURL", func() {
		It("should create a valid challenge url object based on the image ref supplied", func() {
			expected := url.URL{Host: "ghcr.io", Scheme: "https", Path: "/v2/"}
			imageRef, _ := ref.ParseNormalizedNamed("ghcr.io/fugginold/dockwatch:latest")
			Expect(auth.GetChallengeURL(imageRef)).To(Equal(expected))
		})
		It("should assume Docker Hub for image refs with no explicit registry", func() {
			expected := url.URL{Host: "index.docker.io", Scheme: "https", Path: "/v2/"}
			imageRef, _ := ref.ParseNormalizedNamed("fugginold/dockwatch:latest")
			Expect(auth.GetChallengeURL(imageRef)).To(Equal(expected))
		})
		It("should use index.docker.io if the image ref specifies docker.io", func() {
			expected := url.URL{Host: "index.docker.io", Scheme: "https", Path: "/v2/"}
			imageRef, _ := ref.ParseNormalizedNamed("docker.io/fugginold/dockwatch:latest")
			Expect(auth.GetChallengeURL(imageRef)).To(Equal(expected))
		})
	})
})

var scopeImageRegexp = MatchRegexp("^repository:[a-z0-9]+(/[a-z0-9]+)*:pull$")

func getScopeFromImageAuthURL(imageName string) string {
	normalizedRef, _ := ref.ParseNormalizedNamed(imageName)
	challenge := `bearer realm="https://dummy.host/token",service="dummy.host",scope="repository:user/image:pull"`
	URL, _ := auth.GetAuthURL(challenge, normalizedRef)

	scope := URL.Query().Get("scope")
	Expect(scopeImageRegexp.Match(scope)).To(BeTrue())
	return strings.Replace(scope[11:], ":pull", "", 1)
}

// A 401 from the token endpoint unmarshals cleanly into an empty TokenResponse, so
// without a status check the function returned the literal "Bearer " as a success.
// The empty token was not caught downstream either, so the real auth rejection was
// swallowed and resurfaced as a confusing 401 on the digest request.
func TestBearerFromResponse(t *testing.T) {
	newRes := func(status int, body string) *http.Response {
		rec := httptest.NewRecorder()
		rec.Code = status
		rec.Body = bytes.NewBufferString(body)
		return rec.Result()
	}

	// Token-shaped body on purpose: an empty body would be caught by the
	// empty-token guard below, leaving the status check itself unexercised.
	if _, err := auth.BearerFromResponseForTest(newRes(401, `{"token":"abc"}`)); err == nil {
		t.Error("a 401 from the token endpoint must be an error, not an empty bearer")
	}

	if _, err := auth.BearerFromResponseForTest(newRes(200, `{}`)); err == nil {
		t.Error("a 200 carrying no token must be an error, not \"Bearer \"")
	}

	got, err := auth.BearerFromResponseForTest(newRes(200, `{"token":"abc"}`))
	if err != nil || got != "Bearer abc" {
		t.Errorf("token: got %q, %v", got, err)
	}

	// Azure ACR and some GitLab configurations return only access_token. Reading
	// just "token" left the bearer empty and sent every poll down the full-pull path.
	got, err = auth.BearerFromResponseForTest(newRes(200, `{"access_token":"xyz"}`))
	if err != nil || got != "Bearer xyz" {
		t.Errorf("access_token: got %q, %v", got, err)
	}
}

// Only the scheme and the parameter keys are case-insensitive. Lowercasing the whole
// challenge corrupted case-sensitive Artifactory realms and services, and the token
// request 404'd -- visible only as a full pull every cycle.
func TestGetAuthURLPreservesValueCase(t *testing.T) {
	imageRef, err := ref.ParseNormalizedNamed("fugginold/dockwatch")
	if err != nil {
		t.Fatal(err)
	}

	challenge := `Bearer realm="https://art.example.com/v2/token",service="docker-Local"`
	URL, err := auth.GetAuthURL(challenge, imageRef)
	if err != nil {
		t.Fatal(err)
	}

	if got := URL.Query().Get("service"); got != "docker-Local" {
		t.Errorf("service case was corrupted: got %q, want %q", got, "docker-Local")
	}
	if !strings.Contains(URL.Path, "/v2/token") {
		t.Errorf("realm path was corrupted: got %q", URL.Path)
	}
}

// A registry that serves /v2/ without a challenge allows anonymous reads. Treating
// that as an unsupported challenge type failed the digest check outright, so every
// poll fell back to pulling the whole image.
func TestTokenForChallenge(t *testing.T) {
	imageRef, err := ref.ParseNormalizedNamed("myreg.example.com/team/app")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	token, err := auth.TokenForChallengeForTest(ctx, "", 200, imageRef, "")
	if err != nil {
		t.Errorf("a registry that issues no challenge should be usable anonymously: %v", err)
	}
	if token != "" {
		t.Errorf("anonymous access should carry no token, got %q", token)
	}

	// A missing header on a non-2xx is not consent to go anonymous.
	if _, err := auth.TokenForChallengeForTest(ctx, "", 500, imageRef, ""); err == nil {
		t.Error("a 500 with no challenge must not be treated as anonymous access")
	}

	if _, err := auth.TokenForChallengeForTest(ctx, "Basic realm=\"x\"", 401, imageRef, ""); err == nil {
		t.Error("a basic challenge with no credentials must error")
	}
	if got, err := auth.TokenForChallengeForTest(ctx, "Basic realm=\"x\"", 401, imageRef, "creds"); err != nil || got != "Basic creds" {
		t.Errorf("basic challenge: got %q, %v", got, err)
	}
}
