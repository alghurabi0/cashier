# coffeeshop-api Dockerfile (development)
FROM golang:1.25-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the API
RUN go build -o /bin/coffeeshop-api ./cmd/api/...

EXPOSE 8080

CMD ["/bin/coffeeshop-api"]
