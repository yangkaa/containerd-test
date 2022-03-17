export CONTEXT_DIR=/root/code/containerd-test
export TRANSIENT_MOUNT=/root/code/containerd-test:/root/code/containerd-test
export DOCKERFILE_NAME=/root/code/containerd-test/Dockerfile.test
export OUTPUT=docker.io/yangk/builah-test:1.0
go run main.go
