FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service-a

FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=builder /app/service-a /service-a
ENV PORT=8081
USER 65532:65532
EXPOSE 8081
ENTRYPOINT ["/service-a"]


