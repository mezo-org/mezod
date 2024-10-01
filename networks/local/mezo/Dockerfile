FROM golang:1.22.6-bullseye AS build-env

# Install minimum necessary dependencies
ENV PACKAGES curl make git libc-dev bash gcc
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y $PACKAGES

# Set working directory for the build
WORKDIR /go/src/github.com/mezo-org/mezod

# Add source files
COPY . .

# build Ethermint
RUN make build-linux

# Final image
FROM golang:1.22.6-bullseye as final

WORKDIR /

RUN apt-get update

# Copy over binaries from the build-env
COPY --from=build-env /go/src/github.com/mezo-org/mezod/build/mezod /

EXPOSE 26656 26657 1317 9090 8545 8546

# Run mezod by default, omit entrypoint to ease using container with mezod
ENTRYPOINT ["/bin/bash", "-c"]