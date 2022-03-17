FROM golang:1.17 as go
WORKDIR /workspace
COPY . /workspace
RUN apt update \
    && apt -y install bats golang-github-containerd-btrfs-dev git libapparmor-dev libdevmapper-dev libglib2.0-dev libgpgme11-dev libseccomp-dev libselinux1-dev go-md2man \
    && go mod vendor && CGO_ENABLED=1 go build -o main main.go

#FROM goodrainapps/alpine:3.4
#WORKDIR /root
#COPY --from=go /workspace/main .
#RUN chmod +x /root/main && pwd && ls -a
#CMD ["/root/main"]