#
# Build the image
#
FROM golang:1.22.6-bullseye AS build

WORKDIR /go/src/github.com/mezo-org/mezod

RUN apt-get update -y && \
    apt-get install git -y

COPY . .

RUN make build

#
# Runtime image
#
FROM alpine:3.20.3

# Install glibc compatibility for alpine
RUN apk add --no-cache gcompat

COPY --from=build /go/src/github.com/mezo-org/mezod/build/mezod /usr/bin/mezod

CMD ["mezod"]
