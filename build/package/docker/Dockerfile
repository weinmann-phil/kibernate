# syntax=docker/dockerfile:1

FROM golang:1.19-alpine as builder
WORKDIR /app
COPY . ./
RUN GOOS=linux go build -ldflags="-w -s" -o koldstart ./cmd/koldstart
RUN chmod +x koldstart

FROM alpine:3.14
COPY --from=builder /app/koldstart /usr/local/bin/koldstart
USER nobody
ENTRYPOINT ["/usr/local/bin/koldstart"]