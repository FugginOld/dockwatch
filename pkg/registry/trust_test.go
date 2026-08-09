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

		// The value an operator reaches for first is either the domain they type in
		// image names or the key docker login writes into config.json. Neither is the
		// normalized address, so comparing raw strings would reject both and produce a
		// confusing anonymous-pull failure on a private Docker Hub repo.
		When("REPO_HOST is written in another accepted form", func() {
			It("should still match the image registry", func() {
				_ = os.Setenv("REPO_USER", "containrrr-user")
				_ = os.Setenv("REPO_PASS", "containrrr-pass")
				defer func() { _ = os.Unsetenv("REPO_HOST") }()

				for _, repoHost := range []string{
					"docker.io",
					"index.docker.io",
					"https://index.docker.io/v1/",
				} {
					_ = os.Setenv("REPO_HOST", repoHost)
					config, err := EncodedEnvAuth("nginx:latest")
					Expect(err).NotTo(HaveOccurred(), "REPO_HOST %q should match Docker Hub", repoHost)
					Expect(config).NotTo(BeEmpty())
				}

				_ = os.Setenv("REPO_HOST", "Harbor.Example.com")
				config, err := EncodedEnvAuth("harbor.example.com/team/tool:latest")
				Expect(err).NotTo(HaveOccurred(), "host comparison should be case-insensitive")
				Expect(config).NotTo(BeEmpty())
			})
		})
	})

	// EncodedAuth is the only caller in production, and the scoping fix depends on it
	// treating the mismatch error as "fall through to the docker config" rather than
	// returning the environment credentials anyway.
	Describe("EncodedAuth falling back when the scope does not match", func() {
		It("should not return the environment credentials", func() {
			_ = os.Setenv("REPO_USER", "containrrr-user")
			_ = os.Setenv("REPO_PASS", "containrrr-pass")
			_ = os.Setenv("REPO_HOST", "harbor.example.com")
			_ = os.Unsetenv("DOCKER_CONFIG")
			defer func() { _ = os.Unsetenv("REPO_HOST") }()

			envCreds, err := EncodedEnvAuth("harbor.example.com/team/tool:latest")
			Expect(err).NotTo(HaveOccurred())

			auth, _ := EncodedAuth("evil.example.com/tool:latest")
			Expect(auth).NotTo(Equal(envCreds), "credentials must not reach an out-of-scope registry")
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

var _ = Describe("EnvCredentialsAreUnscoped", func() {
	AfterEach(func() {
		_ = os.Unsetenv("REPO_USER")
		_ = os.Unsetenv("REPO_PASS")
		_ = os.Unsetenv("REPO_HOST")
	})

	It("should be true when credentials are set without a scope", func() {
		_ = os.Setenv("REPO_USER", "u")
		_ = os.Setenv("REPO_PASS", "p")
		Expect(EnvCredentialsAreUnscoped()).To(BeTrue())
	})
	It("should be false once REPO_HOST scopes them", func() {
		_ = os.Setenv("REPO_USER", "u")
		_ = os.Setenv("REPO_PASS", "p")
		_ = os.Setenv("REPO_HOST", "harbor.example.com")
		Expect(EnvCredentialsAreUnscoped()).To(BeFalse())
	})
	It("should be false when no environment credentials are set", func() {
		Expect(EnvCredentialsAreUnscoped()).To(BeFalse())
	})
})
