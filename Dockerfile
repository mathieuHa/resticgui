FROM node:14.17.6 as BuilderWeb

COPY front/ /front

WORKDIR /front

RUN npm ci

RUN npm run build

FROM ubuntu:focal-20220105 as BuilderRestic

ARG RESTICVERSION=0.12.1
ARG RESTICURL=https://github.com/restic/restic/releases/download/v${RESTICVERSION}/restic_${RESTICVERSION}_linux_amd64.bz2
ARG BUILDDIR=./build

RUN apt-get update && apt-get install -y \
    curl \
    openssl

RUN mkdir -p ${BUILDDIR}
RUN curl \
    --request GET -sL \
    --url ${RESTICURL} \
    --output ${BUILDDIR}/restic.bz2

RUN bzip2 -d ${BUILDDIR}/restic.bz2

RUN	cp "${BUILDDIR}/restic" /restic

FROM golang:1.17.6 as BuilderResticGui

ARG BUILDDIR=./build

RUN mkdir -p ${BUILDDIR}

COPY . ${BUILDDIR}

WORKDIR ${BUILDDIR}

RUN cd cmd/resticgui && go build -o /resticgui

FROM ubuntu:focal-20220105

LABEL "hanotaux.mathieu.vendor"="MHX"
LABEL version="1.0"
LABEL description="This image contain restic and resticgui webserver"

EXPOSE 8080

WORKDIR /opt

ENV GIN_MODE=release

RUN apt-get update && apt-get install -y \
    net-tools \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY cmd/resticgui .

COPY --from=BuilderRestic /restic /usr/bin/restic
COPY --from=BuilderWeb /build front
COPY --from=BuilderResticGui /resticgui resticgui

CMD ["/opt/resticgui"]
