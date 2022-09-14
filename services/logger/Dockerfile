FROM golang:1.19
WORKDIR /go/src/github.com/dmitryzzz/grpc-example

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./proto/. ./proto
WORKDIR /go/src/github.com/dmitryzzz/grpc-example/services/logger
COPY ./services/logger/. .
RUN go build -o app .

ENTRYPOINT [ "./app" ]