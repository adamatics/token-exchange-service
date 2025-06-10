# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download

# Install swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy the source code
COPY . .

# Generate Swagger documentation
RUN swag init

# Build the application
# CGO_ENABLED=0 for a statically linked binary (required for scratch)
# -ldflags="-w -s" to strip debug symbols and reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o /app/token-exchange-service .

# Stage 2: Create the minimal scratch image
FROM scratch

WORKDIR /

# Copy the binary and certs from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/token-exchange-service /token-exchange-service

# Copy the docs and static directories
COPY --from=builder /app/docs /docs
COPY --from=builder /app/static /static

# Expose port (make sure this matches the PORT env var your app uses)
EXPOSE 9000

# Default non-sensitive configuration
ENV PORT=9000

# Required environment variables that MUST be provided at runtime:
# ENV CLIENT_ID - Azure AD client ID
# ENV CLIENT_SECRET - Azure AD client secret
# ENV TENANT_ID - Azure AD tenant ID

# Optional environment variables that can be provided at runtime:
# ENV PORT - Port to run the application on
# ENV DEFAULT_SCOPE - Default scope to use for token refresh if none is provided

# Command to run the application
ENTRYPOINT ["/token-exchange-service"]
