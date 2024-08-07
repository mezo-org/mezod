FROM golang:1.22.6-bullseye AS build-env

WORKDIR /go/src/github.com/mezo-org/mezod

RUN apt-get update -y && \
    apt-get install git -y

COPY . .

RUN make build

FROM golang:1.22.6-bullseye

RUN apt-get update -y && \
    apt-get install ca-certificates jq -y

WORKDIR /root

COPY --from=build-env /go/src/github.com/mezo-org/mezod/build/mezod /usr/bin/mezod

EXPOSE 26656 26657 1317 9090 8545 8546

CMD ["mezod"]
