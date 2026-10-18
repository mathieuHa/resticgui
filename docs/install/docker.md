# With Docker

The image `ghcr.io/mathieuha/resticgui` holds ResticGUI and restic, for
`linux/amd64`, `linux/arm64` and `linux/arm/v7`. It is rebuilt every week
with the latest Alpine fixes, signed, and comes with an SBOM.

## 1. Start it

Mount the folders to back up under `/backup`, read-only, and a folder for
local repositories if you keep backups on a disk of this machine:

```sh
docker run -d --name resticgui --restart unless-stopped \
  -p 127.0.0.1:3001:3030 \
  -v resticgui:/opt/db \
  -v /home:/backup/home:ro \
  -v /srv/restic:/repositories \
  ghcr.io/mathieuha/resticgui:latest
```

Open <http://127.0.0.1:3001/ui/> and continue with
[your first backup](/guide/first-backup). In the container, the folders are
`/backup/home` and `/repositories`.

## With Compose

The same, as a `compose.yaml`:

```yaml
services:
  resticgui:
    image: ghcr.io/mathieuha/resticgui:latest
    restart: unless-stopped
    ports:
      - 127.0.0.1:3001:3030
    environment:
      # schedules follow this time zone
      TZ: Europe/Paris
    volumes:
      - resticgui:/opt/db
      - /home:/backup/home:ro
      - /srv/restic:/repositories

volumes:
  resticgui:
```

```sh
docker compose up -d
docker compose logs -f
```

## What is where

| In the container | What |
| --- | --- |
| `/opt/db` | The database and `resticgui.key`, the key that encrypts saved passwords and credentials. Keep this volume. |
| `/opt/config.yaml` | The [configuration](/reference/configuration). Mount your own file over it to change settings. |
| `/backup/...` | The folders to back up, by convention. |
| port `3030` | The web interface and the API. |

## Restoring files

The folders above are mounted read-only, so restoring into them fails. To
restore in place, mount them read-write, or mount a separate folder such as
`-v /srv/restore:/restore` and restore there.

## Settings and a password

Copy the image's configuration, change it, and mount it:

```sh
docker run --rm ghcr.io/mathieuha/resticgui:latest cat /opt/config.yaml > config.yaml
docker run --rm -it ghcr.io/mathieuha/resticgui:latest /opt/resticgui hash-password
```

Put the hash in `auth.password_hash`, then add
`-v ./config.yaml:/opt/config.yaml:ro` to `docker run`, or under `volumes:`
in Compose.

## Updates

```sh
docker pull ghcr.io/mathieuha/resticgui:latest
docker rm -f resticgui      # the data stays in the resticgui volume
# then the docker run command again, or: docker compose up -d
```

Use a version tag such as `ghcr.io/mathieuha/resticgui:0.2` to choose when
to move to a new minor version.

## Check the image

Images are signed with cosign, keyless, by the release workflow:

```sh
cosign verify ghcr.io/mathieuha/resticgui:latest \
  --certificate-identity-regexp 'https://github.com/mathieuHa/resticgui/' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```
