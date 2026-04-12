package helpers

import (
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
