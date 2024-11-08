# syntax=docker/dockerfile:1.7-labs

#
# Build layer
#
FROM golang:1.22.8-bullseye AS build

WORKDIR /go/src/github.com/mezo-org/mezod

RUN apt-get update -y && \
    apt-get install jq -y

RUN curl -sL https://deb.nodesource.com/setup_18.x | bash && \
    apt-get update -y && \
    apt-get install -y nodejs

COPY go.mod go.sum ./
RUN go mod download

COPY --parents ./**/Makefile ./
COPY --parents ./**/*.txt ./
COPY --parents ./**/.keep ./
COPY --parents ./**/*.json ./
COPY --parents ./**/*.go ./
COPY --parents ethereum/bindings/portal/gen/_address/BitcoinBridge ./

RUN make bindings
RUN make build

#
# Busybox layer as source of shell commands
#
FROM busybox:stable AS busybox

#
# Production layer
#
FROM gcr.io/distroless/base-nossl:nonroot AS production

ADD --chmod=755 https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux-amd64 /bin/jq

COPY --from=busybox /bin/sh /bin/cat /bin/test /bin/stty /bin/ls /bin/grep /bin/awk /bin/tail /bin/
COPY --from=build /go/src/github.com/mezo-org/mezod/build/mezod /usr/bin/mezod
COPY deployment/docker/entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]

CMD ["mezod"]
