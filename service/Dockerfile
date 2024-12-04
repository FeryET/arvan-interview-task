# syntax=docker.arvancloud.ir/docker/dockerfile:1

FROM docker.arvancloud.ir/golang:1.21.5-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /service

FROM scratch

COPY --from=builder --chmod="0755" /service /service

ENTRYPOINT [ "/service" ]
