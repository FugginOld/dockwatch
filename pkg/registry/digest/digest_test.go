package digest_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fugginold/dockwatch/internal/actions/mocks"
	"github.com/fugginold/dockwatch/internal/meta"
	"github.com/fugginold/dockwatch/pkg/registry/digest"
	wtTypes "github.com/fugginold/dockwatch/pkg/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func TestDigest(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(GinkgoT(), "Digest Suite")
}

var (
	DockerHubCredentials = &wtTypes.RegistryCredentials{
		Username: os.Getenv("CI_INTEGRATION_TEST_REGISTRY_DH_USERNAME"),
		Password: os.Getenv("CI_INTEGRATION_TEST_REGISTRY_DH_PASSWORD"),
	}
	GHCRCredentials = &wtTypes.RegistryCredentials{
		Username: os.Getenv("CI_INTEGRATION_TEST_REGISTRY_GH_USERNAME"),
		Password: os.Getenv("CI_INTEGRATION_TEST_REGISTRY_GH_PASSWORD"),
	}
)

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

var _ = Describe("Digests", func() {
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

	mockContainerNoImage := mocks.CreateMockContainerWithImageInfoP(mockId, mockName, mockImage, mockCreated, nil)

	When("a digest comparison is done", func() {
		It("should return true if digests match",
			SkipIfCredentialsEmpty(GHCRCredentials, func() {
				creds := fmt.Sprintf("%s:%s", GHCRCredentials.Username, GHCRCredentials.Password)
				matches, err := digest.CompareDigest(context.Background(), mockContainer, creds)
				Expect(err).NotTo(HaveOccurred())
				Expect(matches).To(Equal(true))
			}),
		)

		It("should return false if digests differ", func() {

		})
		It("should return an error if the registry isn't available", func() {

		})
		It("should return an error when container contains no image info", func() {
			matches, err := digest.CompareDigest(context.Background(), mockContainerNoImage, `user:pass`)
			Expect(err).To(HaveOccurred())
			Expect(matches).To(Equal(false))
		})
	})
	When("using different registries", func() {
		It("should work with DockerHub",
			SkipIfCredentialsEmpty(DockerHubCredentials, func() {
				fmt.Println(DockerHubCredentials != nil) // to avoid crying linters
			}),
		)
		It("should work with GitHub Container Registry",
			SkipIfCredentialsEmpty(GHCRCredentials, func() {
				fmt.Println(GHCRCredentials != nil) // to avoid crying linters
			}),
		)
	})
	When("sending a HEAD request", func() {
		var server *ghttp.Server
		BeforeEach(func() {
			server = ghttp.NewServer()
		})
		AfterEach(func() {
			server.Close()
		})
		It("should use a custom user-agent", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeader(http.Header{
						"User-Agent": []string{meta.UserAgent},
					}),
					ghttp.RespondWith(http.StatusOK, "", http.Header{
						digest.ContentDigestHeader: []string{
							mockDigest,
						},
					}),
				),
			)
			dig, err := digest.GetDigest(context.Background(), server.URL(), "token")
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
			Expect(err).NotTo(HaveOccurred())
			Expect(dig).To(Equal(mockDigest))
		})

		// Artifactory and nginx-fronted registries strip this header on HEAD.
		// Returning "" as a valid digest made the comparison never match, so every
		// poll pulled the whole image -- the rate-limit consumption the HEAD exists
		// to avoid -- and logged nothing, because that is the normal "differ" path.
		It("should return an error when the registry omits the digest header", func() {
			server.AppendHandlers(
				ghttp.RespondWith(http.StatusOK, "", http.Header{}),
			)
			dig, err := digest.GetDigest(context.Background(), server.URL(), "token")
			Expect(err).To(HaveOccurred())
			Expect(dig).To(BeEmpty())
		})

		// An empty token now means "the registry issued no challenge", which is how
		// a registry serving anonymous reads is handled -- previously that path
		// errored out and every poll fell back to a full pull. A registry that does
		// want auth still answers 401, so a genuinely missing token fails loudly.
		It("should send no authorization header when there is no token", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					func(_ http.ResponseWriter, req *http.Request) {
						Expect(req.Header.Get("Authorization")).To(BeEmpty())
					},
					ghttp.RespondWith(http.StatusOK, "", http.Header{
						digest.ContentDigestHeader: []string{mockDigest},
					}),
				),
			)
			dig, err := digest.GetDigest(context.Background(), server.URL(), "")
			Expect(err).NotTo(HaveOccurred())
			Expect(dig).To(Equal(mockDigest))
		})
	})

	When("transforming registry auth", func() {
		It("should leave malformed base64 auth unchanged", func() {
			input := "!!!not-base64!!!"
			Expect(digest.TransformAuth(input)).To(Equal(input))
		})

		It("should transform encoded json credentials into basic auth", func() {
			payload, err := json.Marshal(wtTypes.RegistryCredentials{Username: "user", Password: "pass"})
			Expect(err).NotTo(HaveOccurred())

			input := base64.StdEncoding.EncodeToString(payload)
			expected := base64.StdEncoding.EncodeToString([]byte("user:pass"))

			Expect(digest.TransformAuth(input)).To(Equal(expected))
		})
	})

	When("reading registry tls skip verify settings", func() {
		It("should tolerate invalid DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY values", func() {
			old := os.Getenv("DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY")
			defer func() { _ = os.Setenv("DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY", old) }()

			_ = os.Setenv("DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY", "not-a-bool")
			_, err := digest.GetDigest(context.Background(), "http://127.0.0.1:1", "token")
			Expect(err).To(HaveOccurred())
		})
	})
})
