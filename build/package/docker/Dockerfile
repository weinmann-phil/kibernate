# syntax=docker/dockerfile:1

FROM golang:1.19-alpine as builder
WORKDIR /app
COPY . ./
RUN GOOS=linux go build -ldflags="-w -s" -o kibernate ./cmd/kibernate
RUN chmod +x kibernate

FROM alpine:3.14
COPY --from=builder /app/kibernate /usr/local/bin/kibernate
USER nobody
ENTRYPOINT ["/usr/local/bin/kibernate"]