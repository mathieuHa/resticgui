# Notifications and Healthchecks

ResticGUI can tell you when something goes wrong, without you opening it.

## Notifiers

Open **Notifications**, the bell in the top bar, and add a notifier. A
notifier is a [Shoutrrr](https://github.com/nicholas-fedor/shoutrrr) URL,
which covers ntfy, Gotify, Discord, Slack, Telegram, email and many more:

| Service  | Example URL |
| -------- | ----------- |
| ntfy     | `ntfy://ntfy.sh/my-backups` (add `?scheme=http` for a server without TLS) |
| Gotify   | `gotify://gotify.example.com/APP_TOKEN` |
| Discord  | `discord://TOKEN@WEBHOOK_ID` |
| Slack    | `slack://hook:TOKEN@webhook` |
| Telegram | `telegram://BOT_TOKEN@telegram?chats=@channel` |
| Email    | `smtp://user:password@smtp.example.com:587/?from=backup@example.com&to=me@example.com` |

Each notifier chooses what it receives:

- **Failures**: a backup, cleanup (forget or prune) or check that failed, or
  that was interrupted because ResticGUI stopped.
- **Missed scheduled backups**: a scheduled backup that is more than an hour
  late, for example because the machine was off or the repository was busy.
  It is sent once per missed run. Missed backups are looked for every five
  minutes.
- **Successes**: every successful backup, cleanup and check, with the
  snapshot and how much was added.

![Notifiers](../../images/ui-notifications.png){.screenshot}

**Send a test** sends a message right away, so you can check the URL.

The URL usually holds a token. ResticGUI stores it encrypted, never shows it
again (the list shows the service only), and leaves it out of error messages.
To change it, edit the notifier and enter a new URL; leaving the field empty
keeps the saved one.

## Healthchecks

[Healthchecks.io](https://healthchecks.io), or a self-hosted Healthchecks,
tells you when backups *stop happening*, which a notification on failure
cannot. In a repository's **Settings**, set its ping URL, such as
`https://hc-ping.com/your-uuid`. Each backup then sends:

- `<url>/start` when it starts,
- `<url>` when it succeeds, with the snapshot in the body,
- `<url>/fail` when it fails or is interrupted, with the error in the body.

![Healthchecks and hooks in a repository's settings](../../images/ui-hooks.png){.screenshot}

Set the check's period to match the backup schedule; Healthchecks alerts you
through its own integrations when a ping does not come.
