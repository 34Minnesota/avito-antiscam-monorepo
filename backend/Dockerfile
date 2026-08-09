FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/antiscam ./cmd/antiscam

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/antiscam /app/antiscam
COPY --from=builder /src/docs /app/docs

RUN mkdir -p /app/logs

EXPOSE 8080

CMD ["/app/antiscam"]