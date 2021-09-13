FROM node:14.17.6 as BuilderWeb

COPY web/ /web

WORKDIR /web

RUN npm ci && npm install -g @angular/cli

RUN ng build --output-path /front

FROM golang:1.17.1 as BuilderRestic

ARG RESTICVERSION=0.12.1
ARG RESTICSRCURL=https://github.com/restic/restic/archive/refs/tags/v${RESTICVERSION}.tar.gz
ARG BUILDDIR=./build

RUN mkdir -p ${BUILDDIR}
RUN curl \
    --request GET -sL \
    --url ${RESTICSRCURL} \
    --output ${BUILDDIR}/restic.tar.gz

RUN tar -C ${BUILDDIR} -xzvf ${BUILDDIR}/restic.tar.gz
RUN	cd ${BUILDDIR}/restic-${RESTICVERSION} && go run build.go --goos linux --goarch amd64
RUN	cp "${BUILDDIR}/restic-${RESTICVERSION}/restic" /restic

FROM golang:1.17.1 as BuilderResticGui

ARG BUILDDIR=./build

RUN mkdir -p ${BUILDDIR}

COPY . ${BUILDDIR}

WORKDIR ${BUILDDIR}

RUN cd cmd/resticgui && go build -o /resticgui 

FROM ubuntu:20.04

LABEL "hanotaux.mathieu.vendor"="MHX"
LABEL version="1.0"
LABEL description="This image contain restic and resticgui webserver"

WORKDIR /opt

#RUN apt-get update && apt-get install -y \ 
#    && rm -rf /var/lib/apt/lists/*

COPY cmd/resticgui .

COPY --from=BuilderRestic /restic /usr/bin/restic
COPY --from=BuilderWeb /front front
COPY --from=BuilderResticGui /resticgui resticgui

CMD ["sleep", "infinity"]
# CMD ["/opt/resticgui"]