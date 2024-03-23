FROM golang:1.22.1 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o ./tmp/main ./cmd/multiplayer-server/main.go

FROM golang:1.22.1-alpine AS build-release-stage

COPY --from=build-stage /app/tmp/main /app/tmp/main
COPY --from=build-stage /app/configuration /app/configuration
COPY --from=build-stage /app/db/migrations /app/db/migrations

EXPOSE 8000

USER nonroot:nonroot

WORKDIR /app

ENTRYPOINT ["/app/tmp/main"]
