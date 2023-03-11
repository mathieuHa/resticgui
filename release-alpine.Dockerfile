ARG BASEDOCKERIMAGE=alpine:3.17.2

FROM docker.io/${BASEDOCKERIMAGE}

ARG RESTICGUIVERSION="0.1.3"
ARG TARGETARCH

LABEL "hanotaux.mathieu.vendor"="MHX"
LABEL version=${RESTICGUIVERSION}
LABEL description="This image contains Restic and Resticgui webserver"

EXPOSE 3030

WORKDIR /opt

ENV GIN_MODE=release

RUN apk add --update --no-cache \
    curl

COPY cmd/resticgui/config-docker.yaml ./config.yaml

COPY --chmod=700 dist/restic-$TARGETARCH /usr/bin/restic
COPY dist/build front
COPY dist/resticgui-$TARGETARCH resticgui

RUN mkdir /opt/db

CMD ["/opt/resticgui"]
