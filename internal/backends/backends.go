// Package backends describes the restic backends resticgui supports: how a
// repository location is written for each, and which environment variables
// configure it. Adding a backend is one entry in Registry.
package backends

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
)

// Backend is one restic backend.
type Backend struct {
	Name string `json:"name"`
	// Prefix starts every location of this backend ("s3:"); empty for local paths.
	Prefix string `json:"prefix"`
	// Example is a location shown to users.
	Example string `json:"example"`
	// Env lists the non-secret variables the backend reads, Secrets the
	// secret ones (stored encrypted, never returned by the API). A variable of
	// Env may also be stored as a secret, as the legacy import does for
	// access key IDs.
	Env     []string `json:"env"`
	Secrets []string `json:"secrets"`
	// Required lists the variables, from Env or Secrets, that must be set.
	Required []string `json:"required"`
}

// Registry holds the supported backends by name. The names match the
// repositories.backend column.
var Registry = map[string]Backend{
	"local": {
		Name: "local", Example: "/srv/restic/photos",
	},
	"s3": {
		Name: "s3", Prefix: "s3:", Example: "s3:https://s3.eu-west-3.amazonaws.com/bucket/photos",
		Env:      []string{"AWS_DEFAULT_REGION", "AWS_ACCESS_KEY_ID", "AWS_PROFILE"},
		Secrets:  []string{"AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN"},
		Required: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
	},
	"b2": {
		Name: "b2", Prefix: "b2:", Example: "b2:bucket:photos",
		Env:      []string{"B2_ACCOUNT_ID"},
		Secrets:  []string{"B2_ACCOUNT_KEY"},
		Required: []string{"B2_ACCOUNT_ID", "B2_ACCOUNT_KEY"},
	},
	"rest": {
		Name: "rest", Prefix: "rest:", Example: "rest:https://backup.example.com:8000/photos/",
		Env:     []string{"RESTIC_REST_USERNAME"},
		Secrets: []string{"RESTIC_REST_PASSWORD"},
	},
	"sftp": {
		// authentication is the SSH configuration of the user running resticgui
		Name: "sftp", Prefix: "sftp:", Example: "sftp:user@host:/srv/restic/photos",
	},
	"rclone": {
		// the remote is configured in rclone's own configuration
		Name: "rclone", Prefix: "rclone:", Example: "rclone:remote:photos",
		Env: []string{"RCLONE_CONFIG"},
	},
	"azure": {
		Name: "azure", Prefix: "azure:", Example: "azure:container:/photos",
		Env:      []string{"AZURE_ACCOUNT_NAME", "AZURE_ENDPOINT_SUFFIX"},
		Secrets:  []string{"AZURE_ACCOUNT_KEY", "AZURE_ACCOUNT_SAS"},
		Required: []string{"AZURE_ACCOUNT_NAME"},
	},
	"gs": {
		Name: "gs", Prefix: "gs:", Example: "gs:bucket:/photos",
		Env:      []string{"GOOGLE_PROJECT_ID", "GOOGLE_APPLICATION_CREDENTIALS"},
		Secrets:  []string{"GOOGLE_ACCESS_TOKEN"},
		Required: []string{"GOOGLE_PROJECT_ID"},
	},
}

// Names returns the backend names, sorted.
func Names() []string {
	names := make([]string, 0, len(Registry))
	for n := range Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

var ErrUnknown = errors.New("unknown backend")

// Validate checks a repository's backend, location and variables, and
// returns every problem found, joined.
func Validate(backend, location string, env, secrets map[string]string) error {
	b, ok := Registry[backend]
	if !ok {
		return fmt.Errorf("%w %q (supported: %s)", ErrUnknown, backend, strings.Join(Names(), ", "))
	}
	var errs []error
	switch {
	case strings.TrimSpace(location) == "":
		errs = append(errs, errors.New("location is required"))
	case b.Prefix == "":
		if !filepath.IsAbs(location) {
			errs = append(errs, fmt.Errorf("a local location must be an absolute path, such as %s", b.Example))
		}
		for _, other := range Registry {
			if other.Prefix != "" && strings.HasPrefix(location, other.Prefix) {
				errs = append(errs, fmt.Errorf("location %q is a %s location, not a local path", location, other.Name))
			}
		}
	case !strings.HasPrefix(location, b.Prefix) || len(location) == len(b.Prefix):
		errs = append(errs, fmt.Errorf("a %s location starts with %q, such as %s", b.Name, b.Prefix, b.Example))
	}
	if i := strings.Index(location, "://"); i >= 0 {
		if u, err := url.Parse(location[strings.LastIndex(location[:i], ":")+1:]); err == nil && u.User != nil {
			if _, hasPassword := u.User.Password(); hasPassword {
				errs = append(errs, errors.New("the location carries a password, which would be shown in the interface: give it as a credential instead (RESTIC_REST_PASSWORD for a REST server)"))
			}
		}
	}
	errs = append(errs, checkKeys("env", env, b.Env)...)
	errs = append(errs, checkKeys("secret", secrets, append(append([]string{}, b.Env...), b.Secrets...))...)
	for _, key := range b.Required {
		if env[key] == "" && secrets[key] == "" {
			errs = append(errs, fmt.Errorf("%s is required for %s", key, b.Name))
		}
	}
	return errors.Join(errs...)
}

func checkKeys(what string, values map[string]string, allowed []string) []error {
	var errs []error
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !contains(allowed, k) {
			if len(allowed) == 0 {
				errs = append(errs, fmt.Errorf("%s variable %s is not used by this backend", what, k))
			} else {
				errs = append(errs, fmt.Errorf("%s variable %s is not one of %s", what, k, strings.Join(allowed, ", ")))
			}
		}
	}
	return errs
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
