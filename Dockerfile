FROM golang:1.17 as go
WORKDIR /workspace
COPY . /workspace
RUN go mod vendor && go build -o main main.go
CMD ["/workspace/main"]

FROM goodrainapps/alpine:3.4
WORKDIR /root
COPY --from=go /workspace/main .
RUN chmod +x /root/main && pwd && ls -a
CMD ["/root/main"]