FROM golang:1.16-alpine as go
WORKDIR .
RUN go build main.go -o main

FROM alpine:3.15
COPY --from=go ./main ./main
CMD ["./main"]
