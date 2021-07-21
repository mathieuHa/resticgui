# ResticGui

A ResticGUI for Restic https://github.com/restic/restic

This resticGUI version works with restic version 0.12.0 or more

# Documentation

## dev environment

### Backend

#### Install go

```sudo apt-get install golang-go```

```choco install go mingw make```


#### Compilation du projet

```
go build
...
```

#### Installation of powershell 7.1

Run msi from https://github.com/PowerShell/PowerShell/releases/tag/v7.1.3


#### Logging documentation

https://github.com/rs/zerolog

#### Configuration documentation

https://github.com/spf13/viper

#### launch server

```
make run
```

### Frontend dev

Install node 14 and angular CLI

```
curl -sL https://deb.nodesource.com/setup_14.x | sudo bash -
sudo apt -y install nodejs
sudo npm install -g @angular/cli
sudo npm install --save-dev @angular-devkit/build-angular
```

#### Start dev web server 

```
ng serve --open
```
