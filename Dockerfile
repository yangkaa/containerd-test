FROM golang:1.16-alpine as go
WORKDIR /workspace
COPY go.mod go.mod
COPY main.go main.go
RUN go mod download && go env -w GO111MODULE=on && go build main.go -o main

FROM alpine:3.15
COPY --from=go /workspace/main .
CMD ["./main"]
