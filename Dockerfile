FROM golang:1.19.2

WORKDIR /app/

COPY cmd /app/cmd/
COPY internal /app/internal/
COPY migrations /app/migrations/
COPY go.mod go.sum /app/

RUN go install github.com/pressly/goose/v3/cmd/goose@latest