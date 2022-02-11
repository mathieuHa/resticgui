# Variables
DISTDIR=${CURDIR}/dist
BUILDDIR=${CURDIR}/build
WEBDIR=front
RESTICVERSION=0.12.1
RESTICSRCURL=https://github.com/restic/restic/releases/download/v${RESTICVERSION}/restic_${RESTICVERSION}_linux_amd64.bz2
TAG?=dev
CI_REGISTRY_IMAGE?=resticgui
CI?=false

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

test: testgo

printvars:
	 echo DISTDIR = ${DISTDIR}
	 echo os = ${os}
	 echo arch = ${arch}

all: build test

install: installweb buildrestic

installwebdev:
	cd $(WEBDIR) && npm install

installweb: createdir
	cp -r $(WEBDIR) $(BUILDDIR)/
	npm ci --prefix $(BUILDDIR)/$(WEBDIR)

installgo:
	choco install go

createdir:
ifeq ("$(wildcard $(DISTDIR))","")
	$(MKDIR) "$(DISTDIR)"
endif
ifeq ("$(wildcard $(BUILDDIR))","")
	$(MKDIR) "$(BUILDDIR)"
endif

buildrestic: createdir
	$(DLRESTIC)
	bzip2 -d ${BUILDDIR}/restic.bz2
	chmod "${BUILDDIR}/restic" 
	cp "${BUILDDIR}/restic" ${DISTDIR}

build: buildgo buildweb buildrestic

buildgo: createdir
	cp cmd/resticgui/config.yaml ${DISTDIR}/config.yaml
	cd cmd/resticgui && go build -o ${DISTDIR}/resticgui${EXT}


buildweb: installweb
	npm run build --prefix $(BUILDDIR)/$(WEBDIR) && \
    cp -r $(BUILDDIR)/$(WEBDIR)/build $(DISTDIR)/

builddocker:
ifeq ($(CI), true)
	docker pull ${CI_REGISTRY_IMAGE}:${TAG} || true
endif
	DOCKER_BUILDKIT=1 docker build -t $(CI_REGISTRY_IMAGE):$(TAG) .

pushdocker: builddocker
ifeq ($(CI), true)
	docker push $(CI_REGISTRY_IMAGE):$(TAG)
endif

rundocker: builddocker
	docker run --rm -d --name resticgui -p "8000:8080" $(CI_REGISTRY_IMAGE):$(TAG)

run:
	cd ${DISTDIR}/ && ./resticgui${EXT}

rundevweb: installwebdev
	cd $(WEBDIR) && npm run dev
    

# testweb: installweb
# 	cd $(BUILDDIR)/web && ng test --progress false --watch false

testgo:
	cd cmd/resticgui && go test ./...

clean:
	$(RM) ${DISTDIR}
	$(RM) $(BUILDDIR)
	$(RM) $(WEBDIR)/node_modules
	$(RM) $(WEBDIR)/build
	$(RM) $(WEBDIR)/.svelte-kit
