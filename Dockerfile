FROM golang:1.22.1-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 go build -o main ./cmd/multiplayer-server/main.go

FROM golang:1.22.1-alpine 

COPY --from=build-stage /app/main /app/main
COPY --from=build-stage /app/configuration /app/configuration
COPY --from=build-stage /app/db/migrations /app/db/migrations

EXPOSE 8000

USER root

WORKDIR /app

ENTRYPOINT ["/app/main"]
