package main

import (
	"fmt"
	"github.com/containerd/containerd"
	"os"
)

func main() {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Println("new containerd cli failed")
		os.Exit(1)
	}
	defer client.Close()

}
