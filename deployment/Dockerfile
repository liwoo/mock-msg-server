FROM golang:1.19.3-alpine

RUN apk add --no-cache git

ENV SECRET=secret \
    SGW_BASE_URL=http://localhost:8080 \
    DISTRICT_ID=1 \
    SERVER_PORT=8080 \
    COUCHBASE_URL=couchbase://localhost \
    COUCHBASE_READS_DB=reads \
    COUCHBASE_WRITES_DB=writes \
    COUCHBASE_USER=username \
    COUCHBASE_PASS=password

# Set the Current Working Directory inside the container
WORKDIR /app/main

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -o ./out/main .


# This container exposes port from env file to the outside world
EXPOSE $SERVER_PORT

# Run the binary program produced by `go install`
CMD ["./out/main"]
