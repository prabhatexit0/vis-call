FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o viscall .

FROM alpine:latest

COPY --from=builder /app/viscall /usr/local/bin/viscall

ENTRYPOINT ["viscall"]
