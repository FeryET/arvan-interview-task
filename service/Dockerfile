# syntax=docker.arvancloud.ir/docker/dockerfile:1

FROM docker.arvancloud.ir/golang:1.21.5-bullseye AS builder

ARG GOCACHE=/root/.cache/go-build

WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=./go/go.sum,target=go.sum \
    --mount=type=bind,source=./go/go.mod,target=go.mod \
    go mod download -x

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="${GOCACHE}" \
    --mount=type=bind,source=./go,target=./ \
    CGO_ENABLED=0 GOOS=linux go build -o /service

FROM scratch

COPY --from=builder --chmod="0755" /service /service

ENTRYPOINT [ "/service" ]
