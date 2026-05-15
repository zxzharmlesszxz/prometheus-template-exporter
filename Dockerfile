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
  -o /out/prometheus-template-exporter \
  ./cmd

FROM debian:bookworm-slim

COPY --from=build /out/prometheus-template-exporter /usr/local/bin/prometheus-template-exporter

ENTRYPOINT ["/usr/local/bin/prometheus-template-exporter"]
CMD ["--web.listen-address=:9900"]
