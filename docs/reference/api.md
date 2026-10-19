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

A repository with several backup plans needs the plan to back up,
`plan_id`, one of the ids its `plans` list:

```sh
curl -fsS -b cookies "$api/repositories/1" | jq '.plans[] | {id, name}'
curl -fsS -b cookies -H 'Content-Type: application/json' \
  -d '{"kind":"backup","plan_id":2}' "$api/repositories/1/jobs"
```

Add a plan that backs up a folder every hour:

```sh
curl -fsS -b cookies -H 'Content-Type: application/json' -d '{
  "name": "Documents",
  "paths": {"include": ["/home/me/Documents"], "exclude": []},
  "schedules": [{"job": "backup", "cron": "0 * * * *"}]
}' "$api/repositories/1/plans"
```

Check the repository and read a tenth of its data, then rebuild its index
and repair its snapshots (the kinds `repair_index`, `repair_snapshots` and
`migrate` are the [maintenance tools](/features/maintenance)):

```sh
curl -fsS -b cookies -H 'Content-Type: application/json' \
  -d '{"kind":"check","read_data_subset":"10%"}' "$api/repositories/1/jobs"
curl -fsS -b cookies -H 'Content-Type: application/json' \
  -d '{"kind":"repair_index"}' "$api/repositories/1/jobs"
curl -fsS -b cookies -H 'Content-Type: application/json' \
  -d '{"kind":"repair_snapshots"}' "$api/repositories/1/jobs"
```

Change the repository's password, read from a file rather than typed on the
command line:

```sh
jq -n --rawfile p new-password.txt '{password: ($p | rtrimstr("\n"))}' |
  curl -fsS -b cookies -H 'Content-Type: application/json' -d @- "$api/repositories/1/password"
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
