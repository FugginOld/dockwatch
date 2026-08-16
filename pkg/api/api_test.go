package api

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	token = "123123123"
)

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

var _ = Describe("API", func() {
	api := New(token)

	Describe("RequireToken middleware", func() {
		It("should return 401 Unauthorized when token is not provided", func() {
			handlerFunc := api.RequireToken(testHandler)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/hello", nil)

			handlerFunc(rec, req)

			Expect(rec.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should return 401 Unauthorized when token is invalid", func() {
			handlerFunc := api.RequireToken(testHandler)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/hello", nil)
			req.Header.Set("Authorization", "Bearer 123")

			handlerFunc(rec, req)

			Expect(rec.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should return 200 OK when token is valid", func() {
			handlerFunc := api.RequireToken(testHandler)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/hello", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			handlerFunc(rec, req)

			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("Start", func() {
		It("should return nil when no handlers are registered", func() {
			api := New(token)
			Expect(api.Start(false)).To(Succeed())
		})

		It("should return an error when token is missing", func() {
			api := &API{Token: "", hasHandlers: true}
			Expect(api.Start(false)).To(MatchError(tokenMissingMsg))
		})

		// Serving in a goroutine used to swallow the bind error, so a port already in
		// use was logged and forgotten: dockwatch carried on scanning with no API and
		// the process still exited 0, which reads as healthy to an orchestrator while
		// the update endpoint does not exist.
		It("should return the bind error when the address is already in use", func() {
			occupied, err := net.Listen("tcp", ":0")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = occupied.Close() }()

			api := New(token)
			api.RegisterFunc("/test", testHandler)
			api.Addr = occupied.Addr().String()

			Expect(api.Start(false)).NotTo(Succeed(),
				"a bind failure must reach the caller, not just the log")
		})
	})

	// The update handler runs the whole scan inline, so a write deadline is a cap on
	// how long an update may take: Go arms it before calling the handler, so a slow
	// update can never flush its response and the caller sees a reset connection for
	// an update that actually succeeded.
	Describe("server timeouts", func() {
		It("should bound the request read but never the response write", func() {
			server := New(token).newServer()

			Expect(server.WriteTimeout).To(BeZero(),
				"a write deadline would truncate a long-running update")
			Expect(server.ReadHeaderTimeout).NotTo(BeZero(),
				"a header deadline is what closes the slow-loris hole")
			Expect(server.IdleTimeout).NotTo(BeZero())
		})
	})
})

func testHandler(w http.ResponseWriter, req *http.Request) {
	_, _ = io.WriteString(w, "Hello!")
}
