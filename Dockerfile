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

# --- Stage 2: runtime
FROM ubuntu:22.04

RUN apt-get update && \
    # `ncurses-base` already installed, but add 256-colour variants
    apt-get install -y --no-install-recommends ncurses-term && \
    rm -rf /var/lib/apt/lists/* && \
    apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app
COPY --from=builder /app/main /app/main

ENV TERM=xterm-256color COLORTERM=truecolor

EXPOSE 2222

CMD ["/app/main"]
