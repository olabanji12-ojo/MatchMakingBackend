FROM golang:1.23-alpine

WORKDIR /app

# Install dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
