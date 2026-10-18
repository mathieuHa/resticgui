// Package notify tells people about backups: Shoutrrr notifications (ntfy,
// Gotify, Discord, Slack, Telegram, email...) on a failure, a success or a
// scheduled backup that did not happen, and Healthchecks.io pings.
package notify

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

// Sender delivers one message to a Shoutrrr URL.
type Sender func(rawURL, title, message string) error

// Shoutrrr sends through the service the URL names.
func Shoutrrr(rawURL, title, message string) error {
	var r router.ServiceRouter
	service, err := r.Locate(rawURL)
	if err != nil {
		return Redact(err, rawURL)
	}
	if err := service.Send(message, &types.Params{"title": title}); err != nil {
		return Redact(err, rawURL)
	}
	return nil
}

// ServiceName returns the scheme of a Shoutrrr URL, such as "ntfy", after
// checking that Shoutrrr knows it and that the URL is complete enough to
// build the service. Its errors never repeat the URL.
func ServiceName(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" {
		return "", errors.New("not a URL such as ntfy://ntfy.sh/topic")
	}
	var r router.ServiceRouter
	if _, err := r.Locate(rawURL); err != nil {
		return "", Redact(fmt.Errorf("not a notification URL Shoutrrr understands: %w", err), rawURL)
	}
	return strings.SplitN(u.Scheme, "+", 2)[0], nil
}

// Redact removes from err's message the URL and the parts of it that can
// carry a token: user, password, path segments and query values.
func Redact(err error, rawURL string) error {
	if err == nil {
		return nil
	}
	msg := strings.ReplaceAll(err.Error(), rawURL, "…")
	if u, perr := url.Parse(rawURL); perr == nil {
		var secrets []string
		if u.User != nil {
			secrets = append(secrets, u.User.Username())
			if p, ok := u.User.Password(); ok {
				secrets = append(secrets, p)
			}
		}
		for _, seg := range strings.Split(u.Path, "/") {
			secrets = append(secrets, seg)
		}
		for _, vs := range u.Query() {
			secrets = append(secrets, vs...)
		}
		for _, s := range secrets {
			// short values (a port, "yes") are not secrets and would mangle the message
			if len(s) >= 6 {
				msg = strings.ReplaceAll(msg, s, "…")
			}
		}
	}
	return errors.New(msg)
}

// Host is what is shown of a URL that cannot be returned: its host.
func Host(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
