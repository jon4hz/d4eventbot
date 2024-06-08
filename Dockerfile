FROM golang:1.20-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o d4eventbot .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/d4eventbot .
COPY ./msg.tmpl .
CMD ["./d4eventbot", "bot"]