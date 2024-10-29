# syntax=docker/dockerfile:1.7-labs

#
# Build layer
#
FROM golang:1.22.8-bullseye AS build

WORKDIR /go/src/github.com/mezo-org/mezod

COPY go.mod go.sum ./
RUN go mod download

COPY Makefile .
COPY --parents ./**/*.txt ./
COPY --parents ./**/.keep ./
COPY --parents ./**/*.json ./
COPY --parents ./**/*.go ./
COPY --parents ethereum/bindings/portal/gen/_address/BitcoinBridge ./
RUN make build

#
# Layer for building tomledit
#
FROM golang:1.22.8-bullseye AS build-tomledit

WORKDIR /go/src/github.com/creachadair/

RUN git clone https://github.com/creachadair/tomledit.git \
    && cd tomledit/cmd/tomledit \
    && go build -o /usr/bin/tomledit

#
# Busybox layer as source of shell commands
#
FROM busybox:stable AS busybox

#
# Production layer
#
# Refs.:
# https://github.com/GoogleContainerTools/distroless/blob/main/base/README.md
#
FROM gcr.io/distroless/base-nossl:nonroot AS production

COPY --from=busybox /bin/sh /bin/cat /bin/test /bin/
COPY --from=build-tomledit /usr/bin/tomledit /usr/bin/tomledit
COPY --from=build /go/src/github.com/mezo-org/mezod/build/mezod /usr/bin/mezod
COPY deployment/docker/init.sh /init.sh
COPY deployment/docker/start.sh /start.sh

CMD ["mezod"]
