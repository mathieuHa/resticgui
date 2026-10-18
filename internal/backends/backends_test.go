package backends

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		name              string
		backend, location string
		env, secrets      map[string]string
		errs              []string
	}{
		{name: "local", backend: "local", location: "/srv/restic/photos"},
		{name: "local relative", backend: "local", location: "restic/photos", errs: []string{"must be an absolute path"}},
		{name: "local with a backend prefix", backend: "local", location: "s3:host/bucket", errs: []string{"must be an absolute path", "is a s3 location"}},
		{name: "missing location", backend: "sftp", location: " ", errs: []string{"location is required"}},
		{
			name: "s3", backend: "s3", location: "s3:https://s3.example.com/bucket",
			env:     map[string]string{"AWS_DEFAULT_REGION": "eu-west-3", "AWS_ACCESS_KEY_ID": "id"},
			secrets: map[string]string{"AWS_SECRET_ACCESS_KEY": "secret"},
		},
		{
			name: "s3 key id stored as a secret", backend: "s3", location: "s3:host/bucket",
			secrets: map[string]string{"AWS_ACCESS_KEY_ID": "id", "AWS_SECRET_ACCESS_KEY": "secret"},
		},
		{
			name: "s3 missing credentials", backend: "s3", location: "s3:host/bucket",
			errs: []string{"AWS_ACCESS_KEY_ID is required", "AWS_SECRET_ACCESS_KEY is required"},
		},
		{
			name: "secret in env is refused", backend: "s3", location: "s3:host/bucket",
			env:     map[string]string{"AWS_ACCESS_KEY_ID": "id", "AWS_SECRET_ACCESS_KEY": "secret"},
			secrets: map[string]string{},
			errs:    []string{"env variable AWS_SECRET_ACCESS_KEY is not one of"},
		},
		{name: "wrong prefix", backend: "b2", location: "s3:host/bucket", errs: []string{`a b2 location starts with "b2:"`, "B2_ACCOUNT_ID is required"}},
		{name: "prefix only", backend: "rest", location: "rest:", errs: []string{`a rest location starts with "rest:"`}},
		{
			name: "rest with credentials", backend: "rest", location: "rest:https://host:8000/",
			env: map[string]string{"RESTIC_REST_USERNAME": "me"}, secrets: map[string]string{"RESTIC_REST_PASSWORD": "pw"},
		},
		{
			name: "unrelated variable", backend: "sftp", location: "sftp:me@host:/srv",
			env:  map[string]string{"LD_PRELOAD": "/tmp/evil.so"},
			errs: []string{"env variable LD_PRELOAD is not used by this backend"},
		},
		{name: "unknown backend", backend: "ftp", location: "ftp://host", errs: []string{`unknown backend "ftp"`}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.backend, tc.location, tc.env, tc.secrets)
			if len(tc.errs) == 0 {
				assert.NoError(t, err)
				return
			}
			if assert.Error(t, err) {
				for _, want := range tc.errs {
					assert.Contains(t, err.Error(), want)
				}
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	// the names must match the CHECK constraint of repositories.backend
	assert.Equal(t, []string{"azure", "b2", "gs", "local", "rclone", "rest", "s3", "sftp"}, Names())
	for name, b := range Registry {
		assert.Equal(t, name, b.Name)
		env, secrets := map[string]string{}, map[string]string{}
		for _, k := range b.Required {
			if contains(b.Secrets, k) {
				secrets[k] = "x"
			} else {
				env[k] = "x"
			}
		}
		assert.NoError(t, Validate(name, b.Example, env, secrets), "the example of %s validates", name)
	}
}
