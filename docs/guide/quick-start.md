# Quick start

ResticGUI is one program: a web interface and a scheduler on top of
[restic](https://restic.net). It keeps its settings in a small SQLite
database and runs restic for every backup, restore and check. Your backups
are plain restic repositories, which the `restic` command line can always
read too.

Pick how you want to run it. Each guide takes a few minutes and ends with a
first backup.

| | Best for | You get |
| --- | --- | --- |
| [On Linux](/install/linux) | A server or a desktop you want to back up directly | A systemd service, with access to every file of the machine |
| [With Docker](/install/docker) | A home server or a NAS running containers | A container with restic inside; you mount the folders to back up |
| [On Kubernetes](/install/kubernetes) | Backing up volumes of a cluster | A Deployment, a PersistentVolumeClaim and a Service |

All three serve the interface at a URL such as `http://127.0.0.1:3001/ui/`.
Then follow [your first backup](/guide/first-backup).

::: warning Before you open it to a network
Without a password, anyone who reaches ResticGUI can read your backups and
restore files on the machine. The guides start it on `127.0.0.1` only. Set a
[password](/reference/security#a-login) before you listen on another
address or put it behind a reverse proxy.
:::

## What you need

- A machine with Linux (amd64, arm64 or armv7), or Docker, or Kubernetes.
  Release binaries also exist for macOS and Windows.
- A place for the backups: a disk, a NAS share, or a cloud bucket. See
  [storage](/features/storage).
- restic 0.17 or later. The Docker image and the Kubernetes manifests include
  it.
