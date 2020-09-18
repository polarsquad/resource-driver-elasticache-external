FROM golang:1.14 AS builder

WORKDIR /app

# Copy just the files needed for download modules to take advantage of caching in Docker for local development
# go.sum is used for cache invalidation
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download


COPY . .

RUN GOOS=linux go build -o /bin/driver humanitec.io/resources/driver-aws-external/cmd/driver

FROM ubuntu:20.04 AS final

WORKDIR /bin

# Required allows x509 certificate to be properly checked
RUN apt-get update && apt-get install ca-certificates -y

COPY --from=builder /bin/driver .

ENTRYPOINT ["/bin/driver"]
