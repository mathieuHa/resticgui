ARG NODEVERSION=18.12.1-alpine
ARG BASEDOCKERIMAGE=ubuntu:jammy-20221130
ARG GOVERSION=1.20.2

FROM docker.io/node:${NODEVERSION} as BuilderWeb

COPY front/ /front

WORKDIR /front

RUN npm ci

RUN npm run build

ARG BASEDOCKERIMAGE

FROM docker.io/${BASEDOCKERIMAGE} as BuilderRestic

ARG TARGETARCH
ARG RESTICVERSION=0.12.1
ARG RESTICURL=https://github.com/restic/restic/releases/download/v${RESTICVERSION}/restic_${RESTICVERSION}_linux_${TARGETARCH}.bz2
ARG BUILDDIR=./build

RUN apt-get update && apt-get install -y \
    curl \
    openssl \
    bzip2

RUN mkdir -p ${BUILDDIR}
RUN curl \
    --request GET -sL \
    --url ${RESTICURL} \
    --output ${BUILDDIR}/restic.bz2

RUN bzip2 -d ${BUILDDIR}/restic.bz2

RUN	cp "${BUILDDIR}/restic" /restic

ARG GOVERSION

FROM docker.io/golang:${GOVERSION} as BuilderResticGui

ARG BUILDDIR=./build

RUN mkdir -p ${BUILDDIR}

COPY . ${BUILDDIR}

WORKDIR ${BUILDDIR}

RUN cd cmd/resticgui && go build -o /resticgui

ARG BASEDOCKERIMAGE

FROM docker.io/${BASEDOCKERIMAGE}

ARG RESTICGUIVERSION="0.2.0"

LABEL "hanotaux.mathieu.vendor"="MHX"
LABEL version=${RESTICGUIVERSION}
LABEL description="This image contains Restic and Resticgui webserver"

EXPOSE 3030

WORKDIR /opt

ENV GIN_MODE=release

RUN apt-get update && apt-get install -y \
    # net-tools \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY cmd/resticgui/config-docker.yaml ./config.yaml

COPY --from=BuilderRestic --chmod=700 /restic /usr/bin/restic
COPY --from=BuilderWeb /front/build front
COPY --from=BuilderResticGui /resticgui resticgui

RUN mkdir /opt/db

CMD ["/opt/resticgui"]
