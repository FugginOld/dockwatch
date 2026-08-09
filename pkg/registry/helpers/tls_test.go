package helpers

import (
	"net/http"
	"net/url"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ShouldSkipRegistryTLSVerify", func() {
	const envKey = "DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY"

	var savedValue string
	var wasSet bool

	BeforeEach(func() {
		savedValue, wasSet = os.LookupEnv(envKey)
		_ = os.Unsetenv(envKey)
	})

	AfterEach(func() {
		if wasSet {
			_ = os.Setenv(envKey, savedValue)
		} else {
			_ = os.Unsetenv(envKey)
		}
	})

	It("returns false when the env var is not set", func() {
		Expect(ShouldSkipRegistryTLSVerify()).To(BeFalse())
	})

	It("returns true when the env var is set to \"true\"", func() {
		_ = os.Setenv(envKey, "true")
		Expect(ShouldSkipRegistryTLSVerify()).To(BeTrue())
	})

	It("returns true when the env var is set to \"1\"", func() {
		_ = os.Setenv(envKey, "1")
		Expect(ShouldSkipRegistryTLSVerify()).To(BeTrue())
	})

	It("returns false when the env var is set to \"false\"", func() {
		_ = os.Setenv(envKey, "false")
		Expect(ShouldSkipRegistryTLSVerify()).To(BeFalse())
	})

	It("returns false when the env var is set to \"0\"", func() {
		_ = os.Setenv(envKey, "0")
		Expect(ShouldSkipRegistryTLSVerify()).To(BeFalse())
	})

	It("returns false and does not panic when the env var has an invalid value", func() {
		_ = os.Setenv(envKey, "notabool")
		Expect(ShouldSkipRegistryTLSVerify()).To(BeFalse())
	})
})

// Go's stdlib drops the Authorization header across a redirect to a different host,
// but it compares hosts only -- not schemes. Without a policy of our own, a registry
// that redirects https -> http on the same host receives the bearer token in
// cleartext.
var _ = Describe("the registry HTTP client", func() {
	It("should refuse a redirect that downgrades https to http", func() {
		client := NewHTTPClient()
		Expect(client.CheckRedirect).NotTo(BeNil(), "the client needs a redirect policy")

		from, err := url.Parse("https://registry.example.com/v2/token")
		Expect(err).NotTo(HaveOccurred())
		to, err := url.Parse("http://registry.example.com/v2/token")
		Expect(err).NotTo(HaveOccurred())

		via := []*http.Request{{URL: from}}
		Expect(client.CheckRedirect(&http.Request{URL: to}, via)).To(HaveOccurred())
	})
	It("should refuse a downgrade that happens later in a redirect chain", func() {
		client := NewHTTPClient()

		first, err := url.Parse("https://registry.example.com/v2/token")
		Expect(err).NotTo(HaveOccurred())
		second, err := url.Parse("https://auth.example.com/token")
		Expect(err).NotTo(HaveOccurred())
		third, err := url.Parse("http://auth.example.com/token")
		Expect(err).NotTo(HaveOccurred())

		via := []*http.Request{{URL: first}, {URL: second}}
		Expect(client.CheckRedirect(&http.Request{URL: third}, via)).To(HaveOccurred())
	})
	It("should allow a redirect that stays on https", func() {
		client := NewHTTPClient()

		from, err := url.Parse("https://registry.example.com/v2/token")
		Expect(err).NotTo(HaveOccurred())
		to, err := url.Parse("https://auth.example.com/token")
		Expect(err).NotTo(HaveOccurred())

		via := []*http.Request{{URL: from}}
		Expect(client.CheckRedirect(&http.Request{URL: to}, via)).To(Succeed())
	})
})
