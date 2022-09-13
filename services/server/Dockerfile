FROM golang:alpine AS builder
#WORKDIR /go/src/github.com/dmitryzzz/grpc-example
WORKDIR /build
COPY go.mod go.sum ./

#WORKDIR /go/src/github.com/dmitryzzz/grpc-example/services/server
COPY ./services/server/. .

RUN go mod download && go mod tidy && go mod verify
RUN go build -o app .

FROM alpine
WORKDIR /build

COPY --from=builder /go/src/github.com/dmitryzzz/grpc-example/services/server/app /build/app

CMD [". /app"]
