package update_test

import (
	"net/http/httptest"
	"testing"

	updateAPI "github.com/fugginold/dockwatch/pkg/api/update"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type fakeScanner struct {
	images []string
	wait   bool
	called bool
}

func (f *fakeScanner) Scan(images []string, wait bool) {
	f.called = true
	f.images = images
	f.wait = wait
}

func TestUpdate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Update API Suite")
}

var _ = Describe("the update API handler", func() {
	It("should parse comma-separated image query values and wait", func() {
		scanner := &fakeScanner{}
		handler := updateAPI.New(scanner)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/v1/update?image=nginx,redis&image=alpine", nil)

		handler.Handle(rec, req)

		Expect(scanner.images).To(Equal([]string{"nginx", "redis", "alpine"}))
		Expect(scanner.wait).To(BeTrue())
	})

	It("should scan with nil images and not wait when image query is absent", func() {
		scanner := &fakeScanner{}
		handler := updateAPI.New(scanner)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/v1/update", nil)

		handler.Handle(rec, req)

		Expect(scanner.called).To(BeTrue())
		Expect(scanner.images).To(BeNil())
		Expect(scanner.wait).To(BeFalse())
	})
})
