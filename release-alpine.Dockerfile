ARG BASEDOCKERIMAGE=alpine:3.17.0

FROM docker.io/${BASEDOCKERIMAGE}

ARG RESTICGUIVERSION="0.1.1-test"
ARG ARCH=amd64

LABEL "hanotaux.mathieu.vendor"="MHX"
LABEL version=${RESTICGUIVERSION}
LABEL description="This image contains Restic and Resticgui webserver"

EXPOSE 3030

WORKDIR /opt

ENV GIN_MODE=release

RUN apk add --update --no-cache \
    curl

COPY cmd/resticgui/config-docker.yaml ./config.yaml

COPY --chmod=700 dist/restic /usr/bin/restic
COPY dist/build front
COPY dist/resticgui-${ARCH} resticgui

RUN mkdir /opt/db

CMD ["/opt/resticgui"]
