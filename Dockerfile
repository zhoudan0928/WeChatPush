FROM golang:alpine as builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download
RUN go get github.com/eatmoreapple/openwechat

# Copy the source code, excluding .env
COPY . .
RUN rm -f .env

# Ensure all dependencies are correctly processed
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main .

FROM alpine

# Install ca-certificates
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .

# Expose the port the app runs on
EXPOSE 8080

# Set environment variable for the port
ENV PORT=8080

CMD ["/app/main"]
