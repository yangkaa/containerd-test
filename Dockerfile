FROM golang:1.16-alpine as go
WORKDIR /workspace
COPY go.mod go.mod
COPY main.go main.go
RUN go mod vendor && go env -w GO111MODULE=on && go build -o main main.go

FROM alpine:3.15
COPY --from=go /workspace/main .
CMD ["./main"]
