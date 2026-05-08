FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/subscriptions ./cmd/subscriptions

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/subscriptions /app/subscriptions

USER app
EXPOSE 8080

ENTRYPOINT ["/app/subscriptions"]
