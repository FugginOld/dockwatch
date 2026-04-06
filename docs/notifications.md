# Notifications

Notification delivery support has been removed from Dockwatch.

The following legacy notification flags are still accepted for backward compatibility, but no notifications are sent:

- `--notifications`
- `--notification-email-*`
- `--notification-slack-*`
- `--notification-msteams-*`
- `--notification-gotify-*`
- `--notification-report`
- `--notification-log-stdout`

When notifications are configured, Dockwatch logs a warning and continues running without sending messages.
