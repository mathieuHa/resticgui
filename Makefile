# Variables
DISTDIR=${CURDIR}/dist
BUILDDIR=${CURDIR}/build
WEBDIR=front
RESTICVERSION=0.19.1
RESTICSRCURL=https://github.com/restic/restic/releases/download/v${RESTICVERSION}/restic_${RESTICVERSION}_linux_${GOARCH}.bz2
TAG?=dev
IMAGE?=resticgui
CI?=false
GOOS?=linux
GOARCH?=amd64

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
		BUILDGOARG= -ldflags '-w -s'
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
	cd cmd/resticgui && GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -o ${DISTDIR}/resticgui-${GOARCH}${EXT} ${BUILDGOARG}



buildweb: installweb ## Build the frontend in dist dir
	npm run build --prefix $(BUILDDIR)/$(WEBDIR) && \
    cp -r $(BUILDDIR)/$(WEBDIR)/build $(DISTDIR)/

builddocker: ## Build the container image for this machine as $(IMAGE):$(TAG)
	docker build -t $(IMAGE):$(TAG) --build-arg VERSION=$(TAG) .

smokedocker: builddocker ## Smoke-test the container image: UI, API and one backup
	scripts/smoke-image.sh $(IMAGE):$(TAG) "" $(TAG)

rundocker: ## Run the container image locally on http://localhost:8000/ui
	docker run --rm -d --name resticgui -p "8000:3030" $(IMAGE):$(TAG)

run: ## Run resticgui binary listening on localhost:8080
	cd ${DISTDIR}/ && ./resticgui-${GOARCH}${EXT}

rundev: buildgo run ## Build and run resticgui binary listening on localhost:8080

rundevweb: installwebdev ## Run dev web listening on localhost:3000
	cd $(WEBDIR) && npm run dev
    

generate: ## Regenerate the sqlc queries in internal/store/db
	sqlc generate -f internal/store/sqlc.yaml

testweb: ## Run the frontend checks and unit tests
	cd $(WEBDIR) && npm run check && npm run lint && npm test

e2ebin: createdir
	cd $(WEBDIR) && npm run build
	cp -r $(WEBDIR)/build/. internal/webui/ui/
	CGO_ENABLED=0 go build -o ${BUILDDIR}/resticgui-e2e ./cmd/resticgui

e2e: e2ebin ## Run the end-to-end tests: built UI in the binary, real restic (RESTIC_BIN)
	cd $(WEBDIR) && RESTICGUI_BIN=${BUILDDIR}/resticgui-e2e npx playwright test

# The Playwright image that matches the installed @playwright/test, so that
# screenshots render the same here and in CI
PLAYWRIGHT_IMAGE=mcr.microsoft.com/playwright:v$(shell node -p "require('./$(WEBDIR)/node_modules/@playwright/test/package.json').version" 2>/dev/null)-noble

e2edocker: e2ebin ## Run the end-to-end tests in the Playwright container and compare the screenshots (ARGS=--update-snapshots to refresh them)
	docker run --rm --init --ipc=host --user $(shell id -u):$(shell id -g) -e HOME=/tmp -e CI \
		-e PLAYWRIGHT_SCREENSHOTS=1 -e RESTICGUI_BIN=/src/build/resticgui-e2e -e RESTIC_BIN=/usr/local/bin/restic \
		-v "$(CURDIR)":/src -v "$(shell command -v $(or $(RESTIC_BIN),restic))":/usr/local/bin/restic:ro \
		-w /src/$(WEBDIR) $(PLAYWRIGHT_IMAGE) npx playwright test $(ARGS)

genapi: ## Regenerate the frontend API types from internal/server/openapi.yaml
	cd $(WEBDIR) && npm run gen:api

testgo: ## Run the go Tests
	go test ./cmd/... ./internal/... -cover

testgoci: ## Run the go Tests and export coverage
	go test ./cmd/... ./internal/... -cover -coverprofile=coverage.out

docs: ## Build the documentation site in docs/.vitepress/dist
	cd docs && npm ci && npm run build

docsdev: ## Serve the documentation site with live reload on http://localhost:5173
	cd docs && npm ci && npm run dev

clean: ## Clean build and temporary directory
	$(RM) ${DISTDIR}
	$(RM) $(BUILDDIR)
	$(RM) $(WEBDIR)/node_modules
	$(RM) $(WEBDIR)/build
	$(RM) $(WEBDIR)/.svelte-kit

.PHONY: help e2ebin e2e e2edocker docs docsdev
.DEFAULT_GOAL := help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'