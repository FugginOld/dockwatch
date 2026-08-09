package registry

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Registry credential helpers", func() {
	Describe("EncodedAuth", func() {
		It("should return repo credentials from env when set", func() {
			var err error
			expected := "eyJ1c2VybmFtZSI6ImNvbnRhaW5ycnItdXNlciIsInBhc3N3b3JkIjoiY29udGFpbnJyci1wYXNzIn0="

			err = os.Setenv("REPO_USER", "containrrr-user")
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("REPO_PASS", "containrrr-pass")
			Expect(err).NotTo(HaveOccurred())

			config, err := EncodedEnvAuth("ghcr.io/fugginold/dockwatch")
			Expect(config).To(Equal(expected))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("EncodedEnvAuth", func() {
		It("should return an error if repo envs are unset", func() {
			_ = os.Unsetenv("REPO_USER")
			_ = os.Unsetenv("REPO_PASS")
			_ = os.Unsetenv("REPO_HOST")

			_, err := EncodedEnvAuth("ghcr.io/fugginold/dockwatch")
			Expect(err).To(HaveOccurred())
		})

		// Without a scope these credentials are offered to whatever registry a watched
		// image happens to live on, so one hostile or typosquatted image is enough to
		// collect them.
		When("REPO_HOST names a different registry than the image", func() {
			It("should not offer the credentials", func() {
				_ = os.Setenv("REPO_USER", "containrrr-user")
				_ = os.Setenv("REPO_PASS", "containrrr-pass")
				_ = os.Setenv("REPO_HOST", "harbor.example.com")
				defer func() { _ = os.Unsetenv("REPO_HOST") }()

				_, err := EncodedEnvAuth("evil.example.com/tool:latest")
				Expect(err).To(HaveOccurred())
			})
		})

		When("REPO_HOST matches the image registry", func() {
			It("should return the credentials", func() {
				_ = os.Setenv("REPO_USER", "containrrr-user")
				_ = os.Setenv("REPO_PASS", "containrrr-pass")
				_ = os.Setenv("REPO_HOST", "harbor.example.com")
				defer func() { _ = os.Unsetenv("REPO_HOST") }()

				config, err := EncodedEnvAuth("harbor.example.com/team/tool:latest")
				Expect(err).NotTo(HaveOccurred())
				Expect(config).NotTo(BeEmpty())
			})
		})
	})

	Describe("EncodedConfigAuth", func() {
		It("should return an error if file is not present", func() {
			var err error

			err = os.Setenv("DOCKER_CONFIG", "/dev/null/should-fail")
			Expect(err).NotTo(HaveOccurred())

			_, err = EncodedConfigAuth("")
			Expect(err).To(HaveOccurred())
		})
	})
})
