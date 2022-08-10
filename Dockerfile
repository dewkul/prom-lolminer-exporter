ARG APP_VERSION=0.0.0-SNAPSHOT
ARG APP_GID=5000
ARG APP_UID=5000

## Build stage
FROM golang:1.18-alpine AS build
ENV CGO_ENABLED=0
WORKDIR /app

# Download deps
COPY . .
RUN go mod download

# Build app
ARG APP_VERSION
RUN go build -v -ldflags="-X 'main.appVersion=${APP_VERSION}'" -o prom-lolminer-exporter

# Test
RUN go test -v .

## Runtime stage
FROM alpine:3 AS runtime
WORKDIR /app

ARG APP_GID
ARG APP_UID
RUN addgroup -g $APP_GID -S app && adduser -G app -u $APP_UID -S app

COPY --from=build /app/prom-lolminer-exporter ./
RUN chown app:app prom-lolminer-exporter

USER app
ENTRYPOINT ["./prom-lolminer-exporter"]
CMD [""]
