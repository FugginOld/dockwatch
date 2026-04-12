package update_test

import (
	"net/http/httptest"
	"testing"

	updateAPI "github.com/fugginold/dockwatch/pkg/api/update"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUpdate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Update API Suite")
}

var _ = Describe("the update API handler", func() {
	It("should parse comma-separated image query values", func() {
		var got []string
		handler := updateAPI.New(func(images []string) {
			got = append([]string{}, images...)
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/v1/update?image=nginx,redis&image=alpine", nil)

		handler.Handle(rec, req)

		Expect(got).To(Equal([]string{"nginx", "redis", "alpine"}))
	})

	It("should call update with nil images when image query is absent", func() {
		called := false
		handler := updateAPI.New(func(images []string) {
			called = true
			Expect(images).To(BeNil())
		}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/v1/update", nil)

		handler.Handle(rec, req)

		Expect(called).To(BeTrue())
	})
})
