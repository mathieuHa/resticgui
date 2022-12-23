# ResticGui

A ResticGUI for Restic https://github.com/restic/restic

This resticGUI version works with restic version 0.12.0 or more

# Documentation

# License

ResticGUI is licensed under [GPL 3.0 License](https://www.gnu.org/licenses/gpl-3.0.txt). You can find the
complete text in [``COPYING``](COPYING).

### Dev tools

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
wget https://go.dev/dl/go1.17.7.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.17.7.linux-amd64.tar.gz
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