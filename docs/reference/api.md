# API

Everything the web interface does goes through a JSON API under `/api/v2`,
described by an OpenAPI document that the server serves at
`/api/v2/openapi.yaml`.

**[Open the API reference](/api.html){target="_self"}**, rendered from the
document of this version.

## Logging in

With a password set, log in once and keep the session cookie:

```sh
curl -c cookies -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"..."}' \
  http://127.0.0.1:3001/api/v2/auth/login
curl -b cookies http://127.0.0.1:3001/api/v2/repositories
```

## Examples

Start a backup of repository 1 and follow it:

```sh
api=http://127.0.0.1:3001/api/v2
job=$(curl -fsS -b cookies -H 'Content-Type: application/json' \
  -d '{"kind":"backup"}' "$api/repositories/1/jobs" | jq .id)
curl -fsS -b cookies "$api/jobs/$job" | jq '{status, summary}'
```

List the snapshots, and download a folder of the newest one as a zip:

```sh
curl -fsS -b cookies "$api/repositories/1/snapshots" | jq '.[-1].id'
curl -fsS -b cookies -o docs.zip \
  "$api/repositories/1/snapshots/latest/download?path=/home/me/Documents&archive=zip"
```

Live job events, as the interface receives them:

```sh
curl -N -b cookies "$api/events"
```
