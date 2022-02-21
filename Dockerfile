FROM golang:1.16 as go
WORKDIR /workspace
COPY . /workspace
RUN go mod vendor && go build -o main main.go
CMD ["/workspace/main"]

#FROM alpine:3.15
#WORKDIR /root
#COPY --from=go /workspace/main .
#RUN chmod +x /root/main && pwd && ls -a
#CMD ["/root/main"]