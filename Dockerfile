FROM docker.io/library/golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod .
COPY . .
RUN go build -o sentinel ./cmd/sentinel

FROM docker.io/library/alpine:3.20

WORKDIR /app
COPY --from=builder /build/sentinel .

ENTRYPOINT ["/app/sentinel"]
