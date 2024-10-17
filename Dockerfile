#
# Build the image
#
FROM golang:1.22.8-bullseye AS build

WORKDIR /go/src/github.com/mezo-org/mezod

RUN apt-get update -y && \
    apt-get install git -y

COPY . .

RUN make build

#
# Debug image (busybox)
#
# Refs.:
# https://github.com/GoogleContainerTools/distroless/blob/main/base/README.md
# The images is tagged using the commit hash from 2024-09-24
FROM gcr.io/distroless/base-nossl:debug-ab72257043915c56b78b53d91a8e0d11d31c4699 AS debug
COPY --from=build /go/src/github.com/mezo-org/mezod/build/mezod /usr/bin/mezod

#
# Production image
#
# Refs.:
# https://github.com/GoogleContainerTools/distroless/blob/main/base/README.md
# The images is tagged using the commit hash from 2024-09-24
FROM gcr.io/distroless/base-nossl:ab72257043915c56b78b53d91a8e0d11d31c4699 AS production

COPY --from=build /go/src/github.com/mezo-org/mezod/build/mezod /usr/bin/mezod

# Required to pass passhrase using pipe
COPY --from=debug /busybox/sh /usr/bin/sh

# For simple health check
COPY --from=debug /busybox/cat /usr/bin/cat

CMD ["mezod"]
