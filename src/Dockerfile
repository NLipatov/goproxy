# syntax=docker/dockerfile:1

FROM golang:latest

# Set destination for COPY
WORKDIR /src

# Copy files
COPY . ./

# Download dependencies
RUN go mod download
RUN go build -o main ./Infrastructure/main

# Run
CMD ["./main"]