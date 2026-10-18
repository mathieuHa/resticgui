# On Linux

ResticGUI runs as a systemd service. As root it can read every file it backs
up and restore files with their owners.

## 1. Install restic and ResticGUI

Install restic 0.17 or later from your distribution or from the
[restic releases](https://github.com/restic/restic/releases):

```sh
sudo apt install restic          # Debian, Ubuntu
sudo dnf install restic          # Fedora
restic version                   # 0.17.0 or later
```

Then download ResticGUI from the
[releases](https://github.com/mathieuHa/resticgui/releases). Set `VERSION` to
the newest one and `ARCH` to `amd64`, `arm64` or `armv7`:

```sh
VERSION=0.2.0
ARCH=amd64
curl -fLO "https://github.com/mathieuHa/resticgui/releases/download/v$VERSION/resticgui_${VERSION}_linux_$ARCH"
curl -fLO "https://github.com/mathieuHa/resticgui/releases/download/v$VERSION/SHA256SUMS"
sha256sum --ignore-missing -c SHA256SUMS
sudo install -m 755 "resticgui_${VERSION}_linux_$ARCH" /usr/local/bin/resticgui
resticgui version
```

## 2. Configure it

ResticGUI reads `config.yaml` from the folder it starts in. Put it in
`/etc/resticgui`:

```sh
base=https://raw.githubusercontent.com/mathieuHa/resticgui/main/deploy/linux
sudo mkdir -p /etc/resticgui
sudo curl -fL -o /etc/resticgui/config.yaml "$base/config.yaml"
sudo curl -fL -o /etc/systemd/system/resticgui.service "$base/resticgui.service"
```

The configuration keeps the interface on this machine and lists the
folders it may back up. Change `paths.roots` to yours, and `restic.bin` if
restic is not in `/usr/local/bin` (`command -v restic` tells):

<<< @/../deploy/linux/config.yaml

The [configuration reference](/reference/configuration) lists every setting.

## 3. Start it

<<< @/../deploy/linux/resticgui.service{ini}

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now resticgui
systemctl status resticgui
```

Open <http://127.0.0.1:3001/ui/>. On a server without a desktop, reach it
through SSH from your computer:

```sh
ssh -L 3001:127.0.0.1:3001 you@server
```

then open the same URL on your computer. Continue with
[your first backup](/guide/first-backup).

## Logs and updates

```sh
journalctl -u resticgui -f       # what it does
sudo systemctl stop resticgui    # then install the new binary as in step 1
sudo systemctl start resticgui
```

The database and its encryption key are in `/var/lib/resticgui`. Back up
`resticgui.key` somewhere safe: without it, the saved repository passwords
and credentials cannot be read. The repositories themselves only need their
own password.

## Open it to your network

Set a password first, then listen on every address:

```sh
sudo resticgui hash-password     # type the password, copy the hash
sudoedit /etc/resticgui/config.yaml
```

```yaml
server:
  ip: 0.0.0.0
auth:
  username: admin
  password_hash: "$2a$10$..."    # the hash
```

```sh
sudo systemctl restart resticgui
```

Prefer HTTPS through a reverse proxy, as [security](/reference/security)
explains.
