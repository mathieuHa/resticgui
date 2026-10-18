# Security review

A review of the whole application before 1.0: the Go server, the web
interface, the deployment files and the dependencies. Every finding below
is fixed in the same release unless it says otherwise, with a test that
fails without the fix.

## Threat model

**Who can reach it.** By default ResticGUI listens on `127.0.0.1`, so only
programs on the machine, and web pages open in a browser on it, can send it
requests. In Docker and Kubernetes it listens on every address of its
container, and the published port or the Service decides who reaches it.

**What a logged-in user can do.** Everything ResticGUI can: read any file
under `paths.roots` (by backing it up and downloading it), write any file
there (by restoring), send its backups to any server, and, with
`hooks.commands: true`, run commands. ResticGUI often runs as root, so a
login is close to a root shell on the files in `paths.roots`. There is one
user; there are no roles.

**What the attacker has.** The review looked at four attackers:

| Attacker | Can | Must not |
| --- | --- | --- |
| A web page open in the user's browser | Send requests to `127.0.0.1` and to a domain it controls | Change anything or read anything through ResticGUI |
| Someone on the network, without the password | Reach the port | Get in, or guess the password at speed |
| A repository shared with others | Make restic print anything: file names, sizes, errors | Run code, break the interface or exhaust memory in the browser |
| A logged-in user, with `paths.roots` and command hooks off | Everything above | Run programs, or reach files outside `paths.roots` |

## Findings

| Id | Severity | Finding | Fix |
| --- | --- | --- | --- |
| RG-01 | High | restic options `sftp.command`, `sftp.args`, `rclone.program` and `rclone.args`, an rclone location defining its backend inline, and `RCLONE_CONFIG` let a logged-in user run any program, even with command hooks off. | Refused unless `hooks.commands: true`, by the API and again by the restic runner. |
| RG-02 | High without a login | A web page could make the browser send `text/plain` requests to the API that start jobs, restore files or test notifiers (CSRF). Without a login, with RG-01, that was code execution from any web page. | Writes need `Content-Type: application/json`, and are refused when `Sec-Fetch-Site` or `Origin` shows another site. |
| RG-03 | High without a login | DNS rebinding: a page on a domain that then resolves to `127.0.0.1` could read and change everything. | Without a login, only IP addresses, `localhost` and `server.allowed_hosts` are answered. |
| RG-04 | Medium | Sessions were signed tokens: a logout did not end a copied cookie, and there was no idle timeout. | Random tokens stored hashed in the database; logout deletes them; sessions end after 24 hours unused and 7 days at most. |
| RG-05 | Medium | No limit on login attempts. | 10 failures per client in 15 minutes, then `429`; two password checks at most at once. |
| RG-06 | Medium | A newline in a backup path passed the `paths.roots` check and added a second path, since restic reads them one per line. | Paths with control characters are refused. |
| RG-07 | Medium | A symbolic link inside a root could point backup paths, restore targets and local repositories outside of `paths.roots`. | Links are followed, as far as the path exists, before checking. |
| RG-08 | Low | Command hooks inherited ResticGUI's environment, including `RESTICGUI_SECRET_KEY`, which decrypts every stored password. | Hooks get the environment without any `RESTICGUI_` variable but their own. |
| RG-09 | Low | No security headers: the interface could be framed (clickjacking). | A Content-Security-Policy allowing only the interface's own scripts (inline ones by hash), `frame-ancestors 'none'`, `X-Frame-Options`, `nosniff`, `Referrer-Policy: no-referrer`; API answers are sandboxed. |
| RG-10 | Low | Webhooks followed redirects, so a hook could be bounced to an address inside the server's network. | Redirects are not followed. |
| RG-11 | Low | A hook that timed out left its child processes running, and its output was kept whole in memory. | The whole process group is killed; the last 256 KiB of output are kept. restic's stderr lines are capped too. |
| RG-12 | Low | The Kubernetes manifests let every pod of the cluster reach ResticGUI, and ran it with every capability. | A NetworkPolicy refusing all connections (port-forward still works), a read-only root, seccomp, and only the capabilities to read and restore files. The systemd unit is sandboxed the same way. |
| RG-13 | Low | The `/metrics` token was compared in variable time. | Constant-time comparison. |
| RG-14 | Low | A password inside a repository location, such as `rest:https://user:password@host/`, was shown in the interface. | Refused, with a pointer to `RESTIC_REST_PASSWORD`. |
| RG-15 | Info | Downloads named files with Go quoting, which garbled non-ASCII names. | RFC 2231 encoding. A download is always `application/octet-stream` and an attachment, so a backed-up page never runs. |

### Accepted

| Id | Severity | Finding | Why it stays |
| --- | --- | --- | --- |
| RG-16 | Low | Testing a notifier tells whether an address of the server's network answers. | Only a logged-in user can do it, and they can already send backups to any address. The response body is never shown. |
| RG-17 | Info | Command hooks are shown in full in the interface. | They are commands, not secrets: read tokens from a file or an environment variable instead of writing them in the command. |
| RG-18 | Info | restic's standard output is read line by line without a limit; a huge line from a crafted repository takes memory. | restic itself holds the same data in memory. The decoding is fuzzed. |
| RG-19 | Info | The image runs as root. | To read the files of every owner and restore them with their owners. Kubernetes drops every capability but those. |
| RG-20 | Info | The docs site's build tools have advisories for their development server. | They only run on a developer's machine; the published site is static files. |

### Checked and fine

- Passwords: bcrypt, at least 8 characters, the same time for a wrong
  user name as for a wrong password.
- The session cookie is `HttpOnly`, `SameSite=Strict`, and `Secure` behind
  HTTPS.
- Repository passwords, credentials, notifier, Healthchecks and webhook
  URLs are encrypted with AES-256-GCM, a fresh nonce each time, and never
  returned by the API or logged. The key file is created `0600`, and refused
  when others can read it.
- restic gets one repository's credentials in its environment, never on
  its command line, and cannot be redirected by them
  (`RESTIC_PASSWORD_COMMAND` and the like are blanked). Snapshot ids, paths
  and patterns are separate arguments, after `--` where they could look like
  flags. Backend variables are allowed per backend.
- `paths.roots` is checked on every path the API takes; `/home` does not
  match `/homeevil`.
- All SQL goes through generated, parameterized queries.
- The interface never renders data as HTML: file names, logs and errors are
  text.

## Tools

Every pull request runs:

- **govulncheck**: known vulnerabilities in the Go code ResticGUI calls.
- **gosec** and **Semgrep**: insecure patterns in the code. Each accepted
  gosec finding carries a `#nosec` comment saying why.
- **npm audit**: the web interface's dependencies.
- **Trivy**: the container image.
- **Fuzzing**: the path checks and the decoding of restic's output.
- **End-to-end checks**: a hostile file name shows as text, every page runs
  under the Content-Security-Policy, and a request from another site is
  refused.

[Report a vulnerability](https://github.com/mathieuHa/resticgui/security/advisories/new)
privately, as [SECURITY.md](https://github.com/mathieuHa/resticgui/blob/main/SECURITY.md)
explains.
