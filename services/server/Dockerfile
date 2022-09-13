FROM golang:1.19
WORKDIR /go/src/github.com/dmitryzzz/grpc-example

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./proto/. ./proto
WORKDIR /go/src/github.com/dmitryzzz/grpc-example/services/server
COPY ./services/server/. .
RUN go build -o app .

ENTRYPOINT [ "./app" ]


# FROM golang:alpine AS builder
# WORKDIR /go/src/github.com/dmitryzzz/grpc-example
# COPY go.mod go.sum ./
# COPY ./proto/. ./proto
# RUN go mod download && go mod verify

# WORKDIR /go/src/github.com/dmitryzzz/grpc-example/services/server
# COPY ./services/server/. .

# RUN go build -o app .

# FROM alpine
# WORKDIR /build

# COPY --from=builder /go/src/github.com/dmitryzzz/grpc-example/services/server/app /build/app

# ENTRYPOINT [ "./app" ]