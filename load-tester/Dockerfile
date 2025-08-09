# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/load-tester ./main.go

FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
COPY --from=builder /bin/load-tester /usr/local/bin/load-tester
ENTRYPOINT ["/usr/local/bin/load-tester"]


