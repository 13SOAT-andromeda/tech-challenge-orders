# ----------------------------------------------------------------------
# STAGE 1: PRODUCTION_BUILDER
# Builds the optimized, static binary for the final production image.
# We use a standard Go image to ensure all build tools are available.
# ----------------------------------------------------------------------
FROM golang:1.25-alpine AS production_builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files first (from root) to leverage Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code (including cmd/api)
COPY . .

# Build the application binary. We explicitly target the package containing main.go.
# CGO_ENABLED=0 ensures the binary is statically compiled for the minimal final image.
# -ldflags "-s -w" removes debugging information and symbol tables for size optimization.
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /usr/local/bin/main ./cmd/api

# ----------------------------------------------------------------------
# STAGE 2: DEVELOPMENT
# Sets up the environment for hot reloading using 'air'.
# ----------------------------------------------------------------------
FROM golang:1.25 AS development

WORKDIR /app

# Install git and ca-certificates so `go install` can fetch remote modules
RUN apt-get update 
RUN apt-get install -y --no-install-recommends build-essential ca-certificates git 
RUN rm -rf /var/lib/apt/lists/*

# Ensure go binaries installed with `go install` are on PATH
ENV PATH="/go/bin:${PATH}"

# Install 'air' for hot reloading (use the canonical repo)
RUN go install github.com/air-verse/air@latest

# Copy go mod and sum files (from root)
COPY go.mod go.sum ./

# Pre-download dependencies so 'air' doesn't do it on every start
RUN go mod download

# Copy the rest of the entire source code (including cmd/api and air.toml)
COPY . .

# Copy the swagger documentation files
COPY ./swagger ./swagger

# Ensure the tmp directory (used by air.toml) exists and is writable
RUN mkdir -p ./tmp && chmod -R 754 ./tmp

# Expose the application port
EXPOSE 8080

# Run air explicitly with the project config file
CMD ["air", "-c", "air.toml"]

# ----------------------------------------------------------------------
# STAGE 3: PRODUCTION
# Creates the minimal runtime image for the final deployment.
# It only copies the compiled binary from the 'production_builder' stage.
# ----------------------------------------------------------------------
FROM alpine:3.22.2 AS production

# Set the working directory
WORKDIR /app

# Create a non-root user to run the application
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot && apk add --no-cache ca-certificates curl

# Copy the optimized, statically linked binary from the builder stage
COPY --from=production_builder /usr/local/bin/main /app/main

COPY --from=production_builder /app/internal/adapter/email/templates \
     /app/internal/adapter/email/templates

COPY --from=production_builder /app/swagger \
/app/swagger

# Ensure the binary is executable
RUN chmod +x /app/main

# Switch to the non-root user
USER nonroot

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/app/main"]
