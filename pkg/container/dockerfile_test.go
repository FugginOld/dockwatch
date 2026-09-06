package container

import (
	"os"
	"regexp"
	"testing"
)

var (
	fromDirective  = regexp.MustCompile(`(?im)^FROM\s`)
	labelDirective = regexp.MustCompile(`(?im)^LABEL\s+(.*)$`)
	lineContinued  = regexp.MustCompile(`\\r?\n\s*`)
)

// The runtime image must carry the label IsDockwatch looks for. Without it
// dockwatch does not recognise its own container: it stops itself like any other
// stale container and never starts the replacement, and
// CheckForMultipleDockwatchInstances stops reaping the renamed old instance. A
// LABEL line was dropped once already, in e9cc67b.
func TestRuntimeImageCarriesDockwatchLabel(t *testing.T) {
	raw, err := os.ReadFile("../../Dockerfile")
	if err != nil {
		t.Fatalf("reading Dockerfile: %v", err)
	}

	// Fold continuations first so a LABEL split over several lines reads as one.
	dockerfile := lineContinued.ReplaceAllString(string(raw), " ")

	// Only the last stage becomes the published image, so a LABEL left in the
	// builder is the same defect with the line still present.
	stages := fromDirective.Split(dockerfile, -1)
	if len(stages) < 3 {
		t.Fatalf("expected a multi-stage Dockerfile, found %d FROM directives; "+
			"the runtime-stage split below is no longer meaningful", len(stages)-1)
	}
	runtime := stages[len(stages)-1]

	// Accept any spelling docker accepts: quoted or bare key and value, and the
	// label sharing a LABEL directive with other keys.
	pair := regexp.MustCompile(`"?` + regexp.QuoteMeta(dockwatchLabel) + `"?\s*=\s*"?true"?(\s|$)`)
	for _, directive := range labelDirective.FindAllStringSubmatch(runtime, -1) {
		if pair.MatchString(directive[1]) {
			return
		}
	}
	t.Errorf("runtime stage does not set %s=true; dockwatch will stop itself on self-update", dockwatchLabel)
}
