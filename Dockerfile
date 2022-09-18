FROM yangk/containerd-test:v0.0.1 as go
WORKDIR /workspace
COPY . /workspace
ENV GOPROXY=https://goproxy.cn
RUN go mod vendor && CGO_ENABLED=1 go build -o main main.go

#ARG FILEDIR=/workspace
#ENV CONTEXT_DIR=${FILEDIR}
#ENV TRANSIENT_MOUNT=${FILEDIR}:${FILEDIR}
#ENV DOCKERFILE_NAME=${FILEDIR}/Dockerfile.test
#ENV OUTPUT=docker.io/yangk/builah-test:0.0.1
#CMD ["/workspace/main"]

#FROM goodrainapps/alpine:3.4
#WORKDIR /root
#COPY --from=go /workspace/main .
#RUN chmod +x /root/main && pwd && ls -a
#CMD ["/root/main"]