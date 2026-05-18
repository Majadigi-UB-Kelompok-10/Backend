FROM golang:1.26.2-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /build/api .

EXPOSE 8888

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8888/health || exit 1

ENV PORT=8888

CMD ["./api"]
