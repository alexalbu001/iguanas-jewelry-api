FROM golang:1.24-alpine AS build

WORKDIR /app

# Install swag for Swagger documentation generation (pinned to match go.mod version)
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the source code
COPY . .

# Generate Swagger documentation
RUN swag init -g cmd/api/main.go -o docs

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

FROM alpine:latest AS production

# Add ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy binary from builder stage
COPY --from=build /app/main .

EXPOSE 8080

CMD ["./main"]