# Variables
DISTDIR=${CURDIR}/dist
BUILDDIR=${CURDIR}/build
WEBDIR=front
RESTICVERSION=0.15.2
RESTICSRCURL=https://github.com/restic/restic/releases/download/v${RESTICVERSION}/restic_${RESTICVERSION}_linux_${GOARCH}.bz2
TAG?=dev
CI_REGISTRY_IMAGE?=resticgui
CI?=false
BASEDOCKERIMAGE?=alpine:latest
NODEVERSION?=18.15.0-alpine
ALPINEDOCKERIMAGE?=alpine:latest
UBUNTUDOCKERIMAGE?=ubuntu:latest
GOVERSION?=1.20.5
GOOS?=linux
GOARCH?=amd64
TARGETARCH?=amd64
DOCKERFILE?=Dockerfile

# OS Detection
ifeq ($(OS),Windows_NT)
	os="windows"
	SHELL := pwsh.exe
	EXT=.exe
    .SHELLFLAGS := -NoProfile -Command
	RM=rm -Force -Recurse -ErrorAction Ignore
	MKDIR= mkdir -Force
	DLRESTIC=Invoke-RestMethod \
		 -Method 'GET' \
         -Uri $(RESTICSRCURL) \
         -OutFile ${BUILDDIR}/restic.tar.gz

	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		arch="amd64"
	endif
	ifeq ($(PROCESSOR_ARCHITECTURE),x86)
		arch="386"
	endif
else
	UNAME_S := $(shell uname -s)
	RM=rm -rf
	DLRESTIC=curl \
			--request GET -sL \
			--url ${RESTICSRCURL} \
			--output ${BUILDDIR}/restic.bz2
	MKDIR= mkdir -p
	ifeq ($(CI), true)
		BUILDGOARG= -ldflags '-w -s -extldflags "-static"'
	endif
	ifeq ($(UNAME_S),Linux)
		os="linux"
	endif
	ifeq ($(UNAME_S),Darwin)
		os="Darwin"
	endif
	UNAME_P := $(shell uname -p)
	ifeq ($(UNAME_P),x86_64)
		arch="amd64"
	endif
	ifneq ($(filter %86,$(UNAME_P)),)
		arch="386"
	endif
	ifneq ($(filter arm%,$(UNAME_P)),)
		arch="arm --goarm 7"
	endif
endif

# Recipes

test: testgo ## Launch Tests for the go backend

printvars: ## Print vars for debug
	 echo DISTDIR = ${DISTDIR}
	 echo WEBDIR = ${WEBDIR}
	 echo BUILDDIR = ${BUILDDIR}
	 echo os = ${os}
	 echo arch = ${arch}

all: build test  ## Run Build and launch Tests

install: installweb buildrestic ## Install the web dependancies and download restic binary

installwebdev: ## Install the web dependancies for Development
	cd $(WEBDIR) && npm install

installweb: createdir ## Install the web dependancies in the build dir
	cp -r $(WEBDIR) $(BUILDDIR)/
	npm i --prefix $(BUILDDIR)/$(WEBDIR)

installgo: ## Install go on Windows using choco
	choco install go

installswag: ## Install swag 1.8 on linux
	wget https://github.com/swaggo/swag/releases/download/v1.8.0/swag_1.8.0_Linux_x86_64.tar.gz
	tar -xvzf swag_1.8.0_Linux_x86_64.tar.gz -C /tmp/
	sudo mv /tmp/swag /usr/bin/
	rm swag*

createdir: ## Create the build and package dist dir if not present
ifeq ("$(wildcard $(DISTDIR))","")
	$(MKDIR) "$(DISTDIR)"
endif
ifeq ("$(wildcard $(BUILDDIR))","")
	$(MKDIR) "$(BUILDDIR)"
endif

buildrestic: createdir ## Download restic binary in build dir
	$(DLRESTIC)
	bzip2 -d ${BUILDDIR}/restic.bz2
	chmod +x "${BUILDDIR}/restic" 
	cp "${BUILDDIR}/restic" ${DISTDIR}

build: buildgo buildweb buildrestic ## Build go backend, Web and restic binary

buildgo: createdir ## Build the go backend in dist dir
	cp -r ${DISTDIR}/build/* ./internal/webui/ui/
	cp cmd/resticgui/config.yaml ${DISTDIR}/config.yaml
	cd cmd/resticgui && GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=1 go build -o ${DISTDIR}/resticgui-${GOARCH}${EXT} ${BUILDGOARG}

buildschema: ## Build the go ent schema backend
	go generate ./ent

showschema: ## Show the go ent schema
	go run entgo.io/ent/cmd/ent describe ./ent/schema

buildweb: installweb ## Build the frontend in dist dir
	npm run build --prefix $(BUILDDIR)/$(WEBDIR) && \
    cp -r $(BUILDDIR)/$(WEBDIR)/build $(DISTDIR)/

builddocker: ## Build the docker image (needs auth for gitlab registry)
ifeq ($(CI), true)
	docker pull ${CI_REGISTRY_IMAGE}:${TAG} || true
endif
	DOCKER_BUILDKIT=1 docker build -t $(CI_REGISTRY_IMAGE):$(TAG) . \
	-f $(DOCKERFILE) \
	--build-arg BASEDOCKERIMAGE=$(UBUNTUDOCKERIMAGE) \
	--build-arg GOVERSION=$(GOVERSION) \
	--build-arg TARGETARCH=$(TARGETARCH)

builddoc: ## Build the swagger openapi documenation
	swag init --parseInternal --parseDependency -g router.go  --dir internal/api

pushdocker: builddocker ## Push the built docker image to gitlab registry
ifeq ($(CI), true)
	docker push $(CI_REGISTRY_IMAGE):$(TAG)
endif

rundocker: ## Run the docker container locally
	docker run --rm -d --name resticgui -p "8000:3030" $(CI_REGISTRY_IMAGE):$(TAG)

run: ## Run resticgui binary listening on localhost:8080
	cd ${DISTDIR}/ && ./resticgui-${GOARCH}${EXT}

rundev: buildgo run ## Build and run resticgui binary listening on localhost:8080

rundevweb: installwebdev ## Run dev web listening on localhost:3000
	cd $(WEBDIR) && npm run dev
    
# testweb: installweb
# 	npm run lint

testgo: ## Run the go Tests
	go test ./internal/... -cover 

testgoci: ## Run the go Tests and export coverage
	go test ./internal/... -cover -coverprofile=coverage.out

clean: ## Clean build and temporary directory
	$(RM) ${DISTDIR}
	$(RM) $(BUILDDIR)
	$(RM) $(WEBDIR)/node_modules
	$(RM) $(WEBDIR)/build
	$(RM) $(WEBDIR)/.svelte-kit

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'