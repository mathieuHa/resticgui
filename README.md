# ResticGui

A smart GUI for Restic [github.com/restic/restic](https://github.com/restic/restic)

It features a complete web interface that ease configuration of:
- scheduling backups
- retention policy
- storage provider configuration
- selecting paths from file tree navigation

The following actions are supported:
- listing Snapshots
- removing specific Snapshot
- show statistics of a specific Snapshot
- search for files in a Snapshot
- dowloading files or directories from a Snapshot

The following storage providers are supported:
- `local`
- `AWS S3`
- `B2`
- `Minio S3`

This ResticGUI version works with Restic version `0.12.1` or more.  
Restic must be installed and present in the $PATH if run as a single binary.

## ScreenShots

### File Tree Navigation
![file tree navigation](/images/resticgui-file-tree-navigation-v011.png "File Tree Navigation")

### Repository General View
![general repository view](/images/resticgui-general-view-v011.png "Repository General View")

### Repository Snapshot List
![snapshot list view](/images/resticgui-snapshot-list.png "Snapshot List View")

### Repository Snapshot Explore
![snapshot files view](/images/resticgui-snapshot-files-view-v011.png "Snapshot Explore View")

## Install

### Run as binary

Download the last stable release [0.1.2](https://gitlab.com/resticgui/resticgui/-/releases/v0.1.2)

> Uncompress  

```bash
mkdir /opt
tar -xvzf resticgui-ce54653c-amd64.tar.gz -C /opt/resticgui
```

> Run

```bash
/opt/resticgui/resticgui
```

### Run with docker

Pull the latest stable image from dockerhub registry or gitlab registry:
- [gitlab](https://gitlab.com/resticgui/resticgui/container_registry)
- [dockerhub](https://hub.docker.com/r/resticgui/resticgui)

Example:

```bash
docker pull resticgui/resticgui:v0.1.2
```
> You can bind mount the directories you want to backup in the container
```bash
docker run -d resticgui/resticgui:v0.1.2 -v /home/myuser:/backup
```  

## Configure

Configuration is made in the `config.yaml` file.  
The `config.yaml` must be place in the same directory as the resticgui binary.  

### Variables

- `logs.debug` (bool, default `false`): enable the debug logs (overrides `logs.level`)
- `logs.level` (int, default `1`): zerolog levels, 0 debug, 1 info, 2 warn, 3 error
- `server.ip` (default `127.0.0.1`) and `server.port` (default `3001`): where the web server listens
- `store.path` (default `resticgui.sqlite`): the SQLite database, created on the first start
- `store.key_file` (default `resticgui.key`): the key that encrypts repository passwords and credentials, created if missing; `$RESTICGUI_SECRET_KEY` takes precedence. Keep a copy: without it the stored passwords cannot be read.
- `store.legacy_path` (default `resticgui.db`): a ResticGUI v0.1 database, imported once when `store.path` does not exist yet. The import can also be run by hand with `resticgui migrate-legacy -from resticgui.db -to resticgui.sqlite`.
- `restic.bin` (default `restic`): the restic binary
- `jobs.workers` (default `2`): how many restic commands may run at once
- `jobs.schedule` (default `true`): run the scheduled backups, forget/prune and checks
- `paths.roots` (default empty, meaning everything): directories that may be backed up, restored into and browsed
- `auth.username` and `auth.password_hash`: the login. Make the hash with `resticgui hash-password` (it reads the password on stdin). With no hash the API needs no login, so only do that on a trusted network.

### Example
```yaml
logs:
  debug: false
  level: 1

server:
  ip: "127.0.0.1"
  port: 3001

store:
  path: resticgui.sqlite
  key_file: resticgui.key

paths:
  roots: [/home/myuser]

auth:
  username: admin
  password_hash: "$2a$10$..."
```


### Demo environment

A public demo environment is available at [demo.resticgui.com](https://demo.resticgui.com)  
It is reinitialized every hour and do not persist data between restarts. 

## Roadmap

Features we plan to work on, roughly in this order. Suggestions are welcome in the issues.

- **Dashboard**: last backup status and age per repository, next run, size over time, success/failure history.
- **Notifications**: alerts on failure, success or a missed schedule through [Shoutrrr](https://github.com/containrrr/shoutrrr) URLs (ntfy, Gotify, Discord, Slack, Telegram, email) and Healthchecks.io pings.
- **Hooks**: a command or webhook before and after a backup, for example to dump a database.
- **Import an existing restic repository** and index its snapshots.
- **Prometheus metrics** at `/metrics`: last backup time and status, duration, data added, repository size.
- **Snapshot diff**: what changed between two snapshots.
- **Backup plans**: several sets of paths, each with its own schedule and retention, into one repository; exclude patterns and per-plan restic options.
- **Copy to a second repository** (`restic copy`) for 3-2-1 backups.
- **Maintenance tools**: unlock, repair, partial data check, password change.
- **Configuration from a file or environment variables** for container deployments.

## Glossary

The Design of Restic component is detailed in [github.com/restic/restic/blob/master/doc/design.rst](https://github.com/restic/restic/blob/master/doc/design.rst)

Extract from this documentation:

> Repository: All data produced during a backup is sent to and stored in a repository in a structured form, for example in a file system hierarchy with several subdirectories. A repository implementation must be able to fulfill a number of operations, e.g. list the contents.

> Snapshot: A Snapshot stands for the state of a file or directory that has been backed up at some point in time. The state here means the content and meta data like the name and modification time for the file or the directory and its contents.

## License

ResticGUI is licensed under [GPL 3.0 License](https://www.gnu.org/licenses/gpl-3.0.txt). You can find the
complete text in [``COPYING``](COPYING).

## Dev tools

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://gitlab.com/resticgui/resticgui)


#### Launch Backend server

```
make run
```
> Listen on http://localhost:3001

#### Start dev web server 

```
make rundevweb
```
> Listen on http://localhost:3000/ui, and sends `/api` to the backend on port 3001

#### Tests

```
make testgo     # Go unit tests
make testweb    # svelte-check, eslint and Vitest
make e2e        # Playwright against the built binary; needs restic, or RESTIC_BIN=/path/to/restic
make genapi     # after changing internal/server/openapi.yaml
```

#### Start container

```
make rundocker
```
> Listen on http://localhost:8000

## Dependancies

### Frontend

#### Install node 24

```
curl -sL https://deb.nodesource.com/setup_24.x | sudo bash -
sudo apt -y install nodejs
```

### Backend

#### Install go

__Linux__
```bash
sudo apt-get install golang-go
# Ou 
wget https://go.dev/dl/go1.27.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.27.2.linux-amd64.tar.gz
```

__Windows__
```bash
choco install go mingw make
```
#### Documentation

The API is described by its OpenAPI document at http://localhost:3001/api/v2/openapi.yaml

### Makefile help

```
all                            Run Build and launch Tests
build                          Build go backend, Web and restic binary
builddocker                    Build the docker image (needs auth for gitlab registry)
buildgo                        Build the go backend in dist dir
buildrestic                    Download restic binary in build dir
buildweb                       Build the frontend in dist dir
clean                          Clean build and temporary directory
createdir                      Create the build and package dist dir if not present
install                        Install the web dependancies and download restic binary
installgo                      Install go on Windows using choco
installweb                     Install the web dependancies in the build dir
installwebdev                  Install the web dependancies for Development
printvars                      Print vars for debug
pushdocker                     Push the built docker image to gitlab registry
rundev                         Build and run resticgui binary listening on localhost:8080
run                            Run resticgui binary listening on localhost:8080
rundevweb                      Run dev web listening on localhost:3000
rundocker                      Run the docker container locally
test                           Launch Tests for the go backend
testgo                         Run the go Tests
```

### Library Doc

#### GO Logging documentation

https://github.com/rs/zerolog

#### Configuration documentation

https://github.com/spf13/viper

# About

[maxlerebourg](https://github.com/maxlerebourg) and [I](https://github.com/mathieuHa) have been using Restic since 2020 at [Primadviz](https://primadviz.com).
We come from a web development and security engineer background and wanted to add a wrapper in the form a web UI that facilitate configuration and usage of Restic.
