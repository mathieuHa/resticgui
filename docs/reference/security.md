# Security

ResticGUI can read every file it backs up and write anywhere it restores,
with the rights of its user, often root. Treat access to it like a root
shell on the machine.

## A login

Set a password as soon as ResticGUI is reachable by anyone but you:

```sh
resticgui hash-password      # in Docker: docker exec -it resticgui /opt/resticgui hash-password
```

```yaml
auth:
  username: admin
  password_hash: "$2a$10$..."
```

Without `auth.password_hash`, ResticGUI logs a warning at start and the
interface and API are open.

## HTTPS

ResticGUI serves plain HTTP. Keep it on `127.0.0.1` and reach it through an
SSH tunnel, or put it behind a reverse proxy that adds HTTPS. Live progress
is a server-sent event stream: turn buffering off for `/api/v2/events`. With
Caddy:

```text
backup.example.com {
	reverse_proxy 127.0.0.1:3001
}
```

With nginx:

```nginx
location / {
    proxy_pass http://127.0.0.1:3001;
    proxy_buffering off;
    proxy_read_timeout 1h;
}
```

## What is stored, and how

- Repository passwords, storage credentials, notifier URLs, Healthchecks
  URLs and webhook URLs are encrypted (AES-256-GCM) with the key in
  `store.key_file` or `$RESTICGUI_SECRET_KEY`. Back the key up apart from the
  database.
- None of them is ever sent back by the API or written to the logs.
- restic gets the credentials of one repository at a time, as environment
  variables, never on its command line.

## Limit what it can touch

- `paths.roots` limits the folders that can be backed up, browsed and
  restored into.
- In Docker or Kubernetes, mount only the folders to back up, read-only
  when you do not restore in place.
- Command [hooks](/features/hooks) are off unless `hooks.commands: true`.
- `/metrics` is off by default; give it a `metrics.token`.

## Report a vulnerability

Please report security issues privately through
[GitHub security advisories](https://github.com/mathieuHa/resticgui/security/advisories/new),
not in public issues.
