# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app .

# Runtime stage
FROM alpine:3.20
RUN adduser -D appuser
USER appuser
COPY --from=builder /app/app /app
EXPOSE 8080
CMD ["/app"]
