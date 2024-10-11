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
# Distroless images are tagged with commit hash of the repository
# https://github.com/GoogleContainerTools/distroless
# ab72257043915c56b78b53d91a8e0d11d31c4699: 2024-10-09
FROM gcr.io/distroless/static-debian12:nonroot-ab72257043915c56b78b53d91a8e0d11d31c4699

COPY --from=build /go/src/github.com/mezo-org/mezod/build/mezod /usr/bin/mezod

EXPOSE 26656 26657 1317 9090 8545 8546

CMD ["mezod"]
