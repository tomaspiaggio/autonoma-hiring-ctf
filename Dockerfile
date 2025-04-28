# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
# CGO_ENABLED=0 builds a static binary, which is suitable for scratch or alpine images
# -o /app/main specifies the output path for the compiled binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main .

# Stage 2: Run the application
# Use a minimal base image
FROM alpine:latest

# Install terminfo database for color support etc.
RUN apk add --no-cache ncurses-terminfo

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Set the TERM environment variable
ENV TERM=xterm-256color

# Expose the correct port the application listens on
EXPOSE 2222

# Command to run the executable
CMD ["./main"]
