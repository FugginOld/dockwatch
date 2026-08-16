package daemon

import (
	"testing"

	"github.com/fugginold/dockwatch/pkg/filters"
)

// Set now stops the old scheduler before starting the replacement rather than
// after. That ordering is strictly more correct, but the window it closes is
// microseconds wide and I could not build a test that distinguishes the two
// orderings -- so this pins the risk the reorder itself introduces instead:
// moving the Stop() above the validation would destroy a working schedule on a
// typo. (Verified: it fails if Stop moves above the AddFunc error check.)
func TestSetKeepsTheOldScheduleWhenTheNewOneIsInvalid(t *testing.T) {
	c := &Controller{runner: NewRunner(Config{}), filter: filters.NoFilter}

	if _, err := c.Set("@every 1h"); err != nil {
		t.Fatal(err)
	}
	before := c.Current()

	if _, err := c.Set("this is not a cron spec"); err == nil {
		t.Fatal("expected an invalid spec to be rejected")
	}

	if c.Current() != before {
		t.Errorf("a rejected schedule replaced the running one: %q -> %q", before, c.Current())
	}
	if c.scheduler == nil {
		t.Error("a rejected schedule left the controller with no scheduler at all")
	}
}
