# Build stage
FROM golang:latest AS builder
RUN apt-get update && apt-get install -y git
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o garden ./cmd/garden
RUN ./garden
RUN cp -r static/* public/

# Runtime stage
FROM nginx:alpine
COPY --from=builder /app/public/ /usr/share/nginx/html/
