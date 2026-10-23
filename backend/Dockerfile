FROM golang:1.26.2-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sitani-api ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S sitani \
    && adduser -S -G sitani sitani
COPY --from=builder /out/sitani-api /usr/local/bin/sitani-api

USER sitani
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/sitani-api"]
