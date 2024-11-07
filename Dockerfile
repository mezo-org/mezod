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
# Layer for building tomledit
#
FROM golang:1.22.8-bullseye AS build-tomledit

WORKDIR /go/src/github.com/creachadair/

# hadolint ignore=DL3003
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
FROM gcr.io/distroless/base-nossl:nonroot AS production

COPY --from=busybox /bin/sh /bin/cat /bin/test /bin/stty /bin/ls /bin/
COPY --from=build-tomledit /usr/bin/tomledit /usr/bin/tomledit
COPY --from=build /go/src/github.com/mezo-org/mezod/build/mezod /usr/bin/mezod
COPY deployment/docker/entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]

CMD ["mezod"]
