package main

import (
	"context"
	"fmt"
	"github.com/containers/podman/v3/pkg/bindings"
	"os"
)
func main() {
	conn, err := bindings.NewConnection(context.Background(), "unix://run/user/1000/podman/podman.sock")
	defer conn.Done()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//client, err := containerd.New("/run/containerd/containerd.sock")
	//if err != nil {
	//	fmt.Println("new containerd cli failed %+v", err)
	//	os.Exit(1)
	//}
	//defer client.Close()
}
