# Build stage
FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o garden ./cmd/garden
RUN ./garden
RUN cp -r static/* public/

# Runtime stage
FROM nginx:alpine
COPY --from=builder /app/public/ /usr/share/nginx/html/
