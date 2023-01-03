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
- local
- _AWS S3_ comming soon [#104](https://gitlab.com/resticgui/resticgui/-/issues/104)
- _B2_ comming soon [#105](https://gitlab.com/resticgui/resticgui/-/issues/105)
- _Minio S3_ comming soon [#106](https://gitlab.com/resticgui/resticgui/-/issues/106)

This ResticGUI version works with Restic version `0.12.1` or more.  
Restic must be installed and present in the $PATH if run as a single binary.

## ScreenShots

### File Tree Navigation
![file tree navigation](/images/resticgui-file-tree-navigation.png "File Tree Navigation")

### Repository General View
![general repository view](/images/resticgui-general-display.png "Repository General View")

### Repository Snapshot List
![snapshot list view](/images/resticgui-snapshot-list.png "Snapshot List View")

## Install

### Run as binary

Download the last stable release [0.0.1](https://gitlab.com/resticgui/resticgui/-/releases/v0.0.1)

> Uncompress  

```bash
mkdir /opt
tar -xvzf resticgui-ce54653c.tar.gz -C /opt/resticgui
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
docker pull resticgui/resticgui:8eb47db5
```
> You can bind mount the directories you want to backup in the container
```bash
docker run -d resticgui/resticgui:8eb47db5 -v /home/myuser:/backup
```  

## Configure

Configuration is made in the `config.yaml` file.  
The `config.yaml` must be place in the same directory as the resticgui binary.  

### Variables

- `logs.debug` 
  - bool
  - default: true
  - enable the debug logs
- `logs.level`
  - string
  - default: 1, expected values are: 1, 2
- `db.name`
  - string
  - default: `resticgui.db`
  - name of the resticgui sqlite database
- `server.ip`
  - string
  - default: `127.0.0.1`
  - IP listening for the web server
- `server.port`
  - string
  - default: `3001`
  - tcp port for the web server
- `server.disable`
  - bool
  - default: `false`
  - disable web server
- `front.path`
  - string
  - default: `./build`
  - path of the webserver static files

### Example
```yaml
# configuration
logs:
  debug: true
  level: 1

db:
  name: resticgui.db

server:
  ip: "127.0.0.1"
  port: 3001
  disable: false

front:
  path: ./build
```

When the backups have been configured and the webserver is not needed to be exposed even in localhost, you can turn off the server with `server.disable: true`. Restart ResticGui and backup will happen in the background while the webserver is not exposed.


### Demo environment

A public demo environment is available at [demo.resticgui.com](https://demo.resticgui.com)  
It is reinitialized every hour and do not persist data between restarts. 

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
> Listen on http://localhost:3000

#### Start container

```
make rundocker
```
> Listen on http://localhost:8000

## Dependancies

### Frontend

#### Install node 16

```
curl -sL https://deb.nodesource.com/setup_16.x | sudo bash -
sudo apt -y install nodejs
```

### Backend

#### Install go

__Linux__
```bash
sudo apt-get install golang-go
# Ou 
wget https://go.dev/dl/go1.19.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.19.4.linux-amd64.tar.gz
```

__Windows__
```bash
choco install go mingw make
```
#### Documentation

http://localhost:8000/swagger/index.html

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
