FROM golang:1.17 as go
WORKDIR /workspace
COPY . /workspace
RUN gpg --recv-keys 0x018BA5AD9DF57A4448F0E6CF8BECF1637AD8C79D \
    && sudo gpg --export 0x018BA5AD9DF57A4448F0E6CF8BECF1637AD8C79D >> /usr/share/keyrings/projectatomic-ppa.gpg \
    && sudo echo 'deb [signed-by=/usr/share/keyrings/projectatomic-ppa.gpg] http://ppa.launchpad.net/projectatomic/ppa/ubuntu zesty main' > /etc/apt/sources.list.d/projectatomic-ppa.list \
    && sudo apt update \
    && sudo apt -y install -t stretch-backports \
    && sudo apt -y install bats btrfs-tools git libapparmor-dev libdevmapper-dev libglib2.0-dev libgpgme11-dev libseccomp-dev libselinux1-dev skopeo-containers go-md2man \
    && go mod vendor && go build -o main main.go
CMD ["/workspace/main"]

#FROM goodrainapps/alpine:3.4
#WORKDIR /root
#COPY --from=go /workspace/main .
#RUN chmod +x /root/main && pwd && ls -a
#CMD ["/root/main"]