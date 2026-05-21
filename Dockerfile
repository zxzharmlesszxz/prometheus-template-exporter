FROM golang:1.26 AS build

WORKDIR /src

ARG VERSION=dev
ARG BRANCH=dev
ARG REVISION=dev
ARG BUILD_USER=docker
ARG BUILD_DATE=unknown

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build \
  -trimpath \
  -ldflags "-s -w \
  -X github.com/prometheus/common/version.Version=${VERSION} \
  -X github.com/prometheus/common/version.Branch=${BRANCH} \
  -X github.com/prometheus/common/version.Revision=${REVISION} \
  -X github.com/prometheus/common/version.BuildUser=${BUILD_USER} \
  -X github.com/prometheus/common/version.BuildDate=${BUILD_DATE}" \
  -o /out/prometheus-exporter-framework \
  ./cmd

FROM debian:bookworm-slim

COPY --from=build /out/prometheus-exporter-framework /usr/local/bin/prometheus-exporter-framework

ENTRYPOINT ["/usr/local/bin/prometheus-exporter-framework"]
CMD ["--web.listen-address=:9900"]
