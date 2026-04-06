package notifications

import (
	"os"
	"strings"

	ty "github.com/fugginold/dockwatch/pkg/types"
	log "github.com/sirupsen/logrus"
)

// LocalLog is used for logs that should never trigger notifications.
var LocalLog = log.WithField("notify", "no")

// NewNotifier creates and returns a new Notifier, using global configuration.
func NewNotifier() ty.Notifier {
	return &noopNotifier{}
}

type noopNotifier struct{}

func (n *noopNotifier) StartNotification() {}

func (n *noopNotifier) SendNotification(ty.Report) {}

func (n *noopNotifier) AddLogHook() {}

func (n *noopNotifier) GetNames() []string { return []string{} }

func (n *noopNotifier) GetURLs() []string { return []string{} }

func (n *noopNotifier) Close() {}

// GetTitle formats the title based on the passed hostname and tag
func GetTitle(hostname string, tag string) string {
	tb := strings.Builder{}

	if tag != "" {
		tb.WriteRune('[')
		tb.WriteString(tag)
		tb.WriteRune(']')
		tb.WriteRune(' ')
	}

	tb.WriteString("Dockwatch updates")

	if hostname != "" {
		tb.WriteString(" on ")
		tb.WriteString(hostname)
	}

	return tb.String()
}

// GetTemplateData populates static data used by preview templates.
func GetTemplateData() StaticData {
	hostname, _ := os.Hostname()
	title := GetTitle(hostname, "")

	return StaticData{
		Host:  hostname,
		Title: title,
	}
}
