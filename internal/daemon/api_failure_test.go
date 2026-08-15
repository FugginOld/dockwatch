package daemon

import (
	"net"
	"testing"

	"github.com/fugginold/dockwatch/internal/actions/mocks"
	t2 "github.com/fugginold/dockwatch/pkg/types"
)

// An API the operator explicitly enabled and that cannot bind must end the run.
// Logging and carrying on left dockwatch scanning with no update endpoint while the
// process still exited 0, so an orchestrator saw a healthy container either way.
func TestRunFailsWhenTheAPICannotBind(t *testing.T) {
	occupied, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Skipf("cannot occupy the API port to force a bind failure: %s", err)
	}
	defer func() { _ = occupied.Close() }()

	client := mocks.CreateMockClient(&mocks.TestData{}, false, false)
	d := New(Config{Client: client})

	err = d.Run(Options{
		Filter:          func(t2.FilterableContainer) bool { return false },
		EnableUpdateAPI: true,
		APIToken:        "test-token",
	})

	if err == nil {
		t.Fatal("Run returned nil when the API could not bind; the process would exit 0 with no API")
	}
}
