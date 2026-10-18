# Configuration

ResticGUI reads `config.yaml` from the folder it starts in: `/etc/resticgui`
with the [Linux service](/install/linux), `/opt` in the
[Docker image](/install/docker) and on [Kubernetes](/install/kubernetes).
It is read once at start: restart ResticGUI after a change.

Repositories, schedules, retention, hooks and notifiers are not in this
file: they are set in the web interface and kept in the database.

## Settings

<!-- settings: every key of cmd/resticgui/config.yaml, checked by a test -->

| Key | Default | What it does |
| --- | --- | --- |
| `server.ip` | `127.0.0.1` | Address to listen on. `0.0.0.0` listens on every address: set a password first. |
| `server.port` | `3001` | Port of the web interface and the API. The images use `3030`. |
| `store.path` | `resticgui.sqlite` | The SQLite database, created on the first start. |
| `store.key_file` | `resticgui.key` | The key that encrypts repository passwords and credentials, created on the first start. `$RESTICGUI_SECRET_KEY` (a base64 32-byte key) takes its place when set. |
| `store.legacy_path` | `resticgui.db` | A database of ResticGUI 0.1, imported once when `store.path` does not exist yet. `""` turns the import off. See [upgrading](/reference/upgrading). |
| `restic.bin` | `restic` | The restic binary, a path or a name found in `PATH`. restic 0.17 or later. |
| `jobs.workers` | `2` | How many restic commands may run at once, across repositories. Each repository runs one at a time. |
| `jobs.schedule` | `true` | Run the schedules. `false` leaves only the jobs started by hand. |
| `paths.roots` | `[]` | Folders that may be backed up, browsed and restored into. Empty allows every path. |
| `auth.username` | `admin` | The login name. |
| `auth.password_hash` | `""` | A bcrypt hash from `resticgui hash-password`. Empty turns the login off: anyone who reaches the server can use it. |
| `hooks.commands` | `false` | Allow command [hooks](/features/hooks), which run with ResticGUI's rights. Webhooks need no setting. |
| `metrics.enabled` | `false` | Serve [Prometheus metrics](/features/metrics) at `/metrics`. |
| `metrics.token` | `""` | When set, `/metrics` needs `Authorization: Bearer <token>`. |
| `logs.level` | `1` | `0` debug, `1` info, `2` warnings, `3` errors only. |
| `logs.debug` | `false` | Debug logs, whatever `logs.level` says. |

<!-- end of settings -->

## A complete file

<<< @/../cmd/resticgui/config.yaml

## Environment

| Variable | What it does |
| --- | --- |
| `RESTICGUI_SECRET_KEY` | The encryption key, base64, instead of `store.key_file`: `openssl rand -base64 32`. |
| `TZ` | The time zone of the schedules, such as `Europe/Paris`. |
| `RESTIC_CACHE_DIR` | Where restic keeps its cache; by default `~/.cache/restic`. |
| `HTTPS_PROXY`, `NO_PROXY` | A proxy for the cloud storages. |
| `GIN_MODE=release` | Quieter HTTP logs; the images set it. |

ResticGUI passes restic only the variables it needs (`PATH`, `HOME`, `TZ`,
the cache, certificate and proxy settings, `SSH_AUTH_SOCK`) and the
credentials of the repository at hand, never its own secrets.

## Commands

| Command | Does |
| --- | --- |
| `resticgui` | Starts the server. |
| `resticgui version` | Prints the version. |
| `resticgui hash-password` | Reads a password and prints its hash for `auth.password_hash`. |
| `resticgui migrate-legacy -from OLD.db -to NEW.db [-key-file FILE]` | Imports a ResticGUI 0.1 database by hand. |
