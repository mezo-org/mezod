#
# Build the image
#
FROM golang:1.22.8-bullseye AS build

WORKDIR /go/src/github.com/mezo-org/mezod

RUN apt-get update -y && \
    apt-get install git -y

COPY . .

RUN make build

# Busybox image as source for shell binary
FROM busybox:stable AS shell

#
# Production image
#
# Refs.:
# https://github.com/GoogleContainerTools/distroless/blob/main/base/README.md
# The images is tagged using the commit hash from 2024-09-24
# Debug image contains shell which is used in Docker Compose
FROM gcr.io/distroless/base-nossl:ab72257043915c56b78b53d91a8e0d11d31c4699 AS production

COPY --from=shell /bin/sh /bin/sh

COPY --from=build /go/src/github.com/mezo-org/mezod/build/mezod /usr/bin/mezod

CMD ["mezod"]
