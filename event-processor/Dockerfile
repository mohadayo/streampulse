FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o event-processor .

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/event-processor .

EXPOSE 8081

CMD ["./event-processor"]
