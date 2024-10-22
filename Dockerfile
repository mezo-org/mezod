#
# Build layer
#
FROM golang:1.22.8-bullseye AS build

WORKDIR /go/src/github.com/mezo-org/mezod

RUN apt-get update -y && \
    apt-get install git -y

COPY . .

RUN make build

#
# Busybox layer as source of shell binary
#
FROM busybox:stable AS shell

#
# Production layer
#
# Refs.:
# https://github.com/GoogleContainerTools/distroless/blob/main/base/README.md
#
FROM gcr.io/distroless/base-nossl:nonroot AS production

COPY --from=shell /bin/sh /bin/sh

COPY --from=build /go/src/github.com/mezo-org/mezod/build/mezod /usr/bin/mezod

CMD ["mezod"]
