FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY config/ ./config/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main ./cmd/server

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
RUN mkdir -p /app/uploads

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
