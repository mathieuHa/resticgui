# ResticGui

A ResticGUI for Restic https://github.com/restic/restic

This resticGUI version works with restic version 0.12.0 or more

# Documentation

## dev environment

### Backend

#### Install go

```sudo apt-get install golang-go```

#### Installation of go packages

```
go get github.com/gorilla/mux
...
```

#### launch server

```
go backend/test/main.go
```

### Frontend

Install node 14 and angular CLI

```
curl -sL https://deb.nodesource.com/setup_14.x | sudo bash -
sudo apt -y install nodejs
sudo npm install -g @angular/cli
```

#### Start dev web server 

```
ng serve --open
```