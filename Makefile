# Variables
DISTDIR=${CURDIR}/dist
BUILDDIR=${CURDIR}/build
RESTICVERSION=0.12.0
RESTICSRCURL=https://github.com/restic/restic/archive/refs/tags/v${RESTICVERSION}.tar.gz

# OS Detection
ifeq ($(OS),Windows_NT)
	os="windows"
	SHELL := pwsh.exe
	EXT=.exe
    .SHELLFLAGS := -NoProfile -Command
	RM=rm -Force -Recurse
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
			--output ${BUILDDIR}/restic.tar.gz
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

test: testgo testweb


printvars:
	 echo DISTDIR = ${DISTDIR}
	 echo os = ${os}
	 echo arch = ${arch}

all: build test

install: installweb buildrestic

installwebdev:
	cd web && npm install

installweb: createdir
	$(RM) web/node_modules
	cp -r web $(BUILDDIR)/
	npm ci --prefix $(BUILDDIR)/web

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
	tar -C ${BUILDDIR} -xzvf ${BUILDDIR}/restic.tar.gz
	cd ${BUILDDIR}/restic-${RESTICVERSION} && go run build.go --goos ${os} --goarch ${arch}
	cp "${BUILDDIR}/restic-${RESTICVERSION}/restic${EXT}" ${DISTDIR}

build: buildgo buildweb buildrestic

buildgo: createdir
	cp cmd/resticgui/config.yaml ${DISTDIR}/config.yaml
	cd cmd/resticgui && go build -o ${DISTDIR}/resticgui${EXT}

buildweb: installweb
	cd $(BUILDDIR)/web && ng build --output-path ${DISTDIR}/front/

run:
	cd ${DISTDIR}/ && ./resticgui${EXT}

testweb: installweb
	cd $(BUILDDIR)/web && ng test --progress false --watch false

testgo:
	cd cmd/resticgui && go test ./...

clean:
	$(RM) ${DISTDIR}
	$(RM) $(BUILDDIR)
