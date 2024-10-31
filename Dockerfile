FROM golang:alpine as builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download
RUN go get github.com/eatmoreapple/openwechat

# Copy the source code
COPY . .

# Ensure all dependencies are correctly processed
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main .

FROM alpine

COPY --from=builder /app/main /app/main
COPY --from=builder /app/.env /app/.env

CMD ["/app/main"]
