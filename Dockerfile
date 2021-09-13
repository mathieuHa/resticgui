FROM node:14.17.6 as BuilderNode

COPY web .

RUN npm ci

RUN ng build --output-path /front

FROM golang:1.17.1 as BuilderGo

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

FROM ubuntu:20.04

LABEL "hanotaux.mathieu.vendor"="MHX"
LABEL version="1.0"
LABEL description="This image contain restic and resticgui webserver"

#RUN apt-get update && apt-get install -y \ 
#    && rm -rf /var/lib/apt/lists/*

COPY --from=BuilderGo /restic /usr/bin/restic
COPY --from=BuilderNode /front front

