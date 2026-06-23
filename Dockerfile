FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/go-chat ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
WORKDIR /app

COPY --from=build /bin/go-chat /app/go-chat
COPY static /app/static

USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/go-chat"]
