package notifications

import (
	"os"
	"strings"

	ty "github.com/fugginold/dockwatch/pkg/types"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
)

// LocalLog is used for logs that should never trigger notifications.
var LocalLog = log.WithField("notify", "no")

// NewNotifier creates and returns a new Notifier, using global configuration.
func NewNotifier(c *cobra.Command) ty.Notifier {
	f := c.Flags()
	configuredTypes, _ := f.GetStringSlice("notifications")
	if len(configuredTypes) > 0 {
		log.Warn("Notification delivery is disabled")
	}

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

// GetTemplateData populates the static notification data from flags and environment
func GetTemplateData(c *cobra.Command) StaticData {
	f := c.PersistentFlags()

	hostname, _ := f.GetString("notifications-hostname")
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	title := ""
	if skip, _ := f.GetBool("notification-skip-title"); !skip {
		tag, _ := f.GetString("notification-title-tag")
		if tag == "" {
			// For legacy email support
			tag, _ = f.GetString("notification-email-subjecttag")
		}
		title = GetTitle(hostname, tag)
	}

	return StaticData{
		Host:  hostname,
		Title: title,
	}
}
