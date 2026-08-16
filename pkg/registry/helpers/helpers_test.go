package helpers

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helper Suite")
}

var _ = Describe("the helpers", func() {
	Describe("GetRegistryAddress", func() {
		It("should return error if passed empty string", func() {
			_, err := GetRegistryAddress("")
			Expect(err).To(HaveOccurred())
		})
		It("should return index.docker.io for image refs with no explicit registry", func() {
			Expect(GetRegistryAddress("dockwatch")).To(Equal("index.docker.io"))
			Expect(GetRegistryAddress("fugginold/dockwatch")).To(Equal("index.docker.io"))
		})
		It("should return index.docker.io for image refs with docker.io domain", func() {
			Expect(GetRegistryAddress("docker.io/dockwatch")).To(Equal("index.docker.io"))
			Expect(GetRegistryAddress("docker.io/fugginold/dockwatch")).To(Equal("index.docker.io"))
		})
		It("should return the host if passed an image name containing a local host", func() {
			Expect(GetRegistryAddress("henk:80/dockwatch")).To(Equal("henk:80"))
			Expect(GetRegistryAddress("localhost/dockwatch")).To(Equal("localhost"))
		})
		It("should return the server address if passed a fully qualified image name", func() {
			Expect(GetRegistryAddress("github.com/containrrr/config")).To(Equal("github.com"))
		})
	})

	Describe("NormalizeRegistryHost", func() {
		It("should map every spelling of Docker Hub onto the address GetRegistryAddress returns", func() {
			for _, host := range []string{
				"docker.io",
				"index.docker.io",
				"https://index.docker.io/v1/",
				// The host Hub actually serves the v2 API from, which operators copy
				// out of the registry docs.
				"registry-1.docker.io",
				"DOCKER.IO",
			} {
				Expect(NormalizeRegistryHost(host)).To(Equal(DefaultRegistryHost), host)
			}
		})
		It("should keep a port significant, since a different port is a different registry", func() {
			Expect(NormalizeRegistryHost("harbor.example.com:5000")).To(Equal("harbor.example.com:5000"))
			Expect(NormalizeRegistryHost("harbor.example.com")).NotTo(Equal("harbor.example.com:5000"))
		})
		It("should not let a crafted value widen the scope onto Docker Hub", func() {
			Expect(NormalizeRegistryHost("evil.com/../index.docker.io")).To(Equal("evil.com"))
			Expect(NormalizeRegistryHost("index.docker.io.evil.com")).To(Equal("index.docker.io.evil.com"))
			Expect(NormalizeRegistryHost("https://")).To(BeEmpty())
		})
	})
})
