# Build
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /src

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /dep-dashboard ./cmd

# Runtime
FROM alpine:3.23

RUN adduser -D -u 1000 app
RUN mkdir -p /data && chown app:app /data
USER app
WORKDIR /data
COPY --from=builder /dep-dashboard /usr/local/bin/dep-dashboard
EXPOSE 8080

ENTRYPOINT ["dep-dashboard"]