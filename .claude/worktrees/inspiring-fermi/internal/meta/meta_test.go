package meta_test

import (
	"strings"
	"testing"

	"github.com/fugginold/dockwatch/internal/meta"
	"github.com/stretchr/testify/assert"
)

func TestVersion_DefaultValue(t *testing.T) {
	assert.Equal(t, "v0.2.0", meta.Version)
}

func TestUserAgent_ContainsVersion(t *testing.T) {
	assert.True(t, strings.Contains(meta.UserAgent, meta.Version),
		"UserAgent %q should contain Version %q", meta.UserAgent, meta.Version)
}

func TestUserAgent_HasDockwatchPrefix(t *testing.T) {
	assert.True(t, strings.HasPrefix(meta.UserAgent, "Dockwatch/"),
		"UserAgent %q should start with 'Dockwatch/'", meta.UserAgent)
}

func TestUserAgent_Format(t *testing.T) {
	expected := "Dockwatch/" + meta.Version
	assert.Equal(t, expected, meta.UserAgent)
}
