ARG BASEDOCKERIMAGE=ubuntu:jammy-20221130

FROM docker.io/${BASEDOCKERIMAGE}

ARG RESTICGUIVERSION="0.1.1-test"
ARG ARCH=amd64

LABEL "hanotaux.mathieu.vendor"="MHX"
LABEL version=${RESTICGUIVERSION}
LABEL description="This image contains Restic and Resticgui webserver"

EXPOSE 3030

WORKDIR /opt

ENV GIN_MODE=release

RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY cmd/resticgui/config-docker.yaml ./config.yaml

COPY --chmod=700 dist/restic /usr/bin/restic
COPY dist/build front
COPY dist/resticgui-${ARCH} resticgui

RUN mkdir /opt/db

CMD ["/opt/resticgui"]
