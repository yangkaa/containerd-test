package main

import (
	"containerd-test/criutil"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"
	"os"
	"time"
	"k8s.io/kubernetes/pkg/kubelet/kuberuntime/logs"
)

func main() {
	ctx := context.Background()
	runtimeClient, runtimeConn, err := criutil.GetRuntimeClient(&ctx)
	if err != nil {
		logrus.Errorf("get runtime client failed %v", err)
		return
	}
	defer criutil.CloseConnection(runtimeConn)

	imageClient, imageConn, err := criutil.GetImageClient(&ctx)
	if err != nil {
		logrus.Errorf("get runtime client failed %v", err)
		return
	}
	defer criutil.CloseConnection(imageConn)

	imageSpec := &v1alpha2.ImageSpec{
		Image: "nginx",
	}
	resp, err := imageClient.PullImage(context.Background(), &v1alpha2.PullImageRequest{
		Image: imageSpec,
	})
	if err != nil {
		logrus.Errorf("Pull Image failed %v", err)
		return
	}
	logrus.Println(resp.String())

	createResp, err :=runtimeClient.CreateContainer(context.Background(), &v1alpha2.CreateContainerRequest{
		Config: &v1alpha2.ContainerConfig{
			Image: imageSpec,
			Args: []string{"run"},
		},
	})
	if err  != nil {
		logrus.Errorf("Create Container failed %v", err)
		return
	}
	logrus.Println("create container", createResp.String())


	runtimeService, err := remote.NewRemoteRuntimeService(criutil.RuntimeEndpoint, time.Second *3 )
	if err != nil {
		logrus.Errorf("New Remote Runtime Service %v", err)
		return
	}
	status, err := runtimeService.ContainerStatus(createResp.GetContainerId())
	if err != nil {
		logrus.Errorf("Get Container Status %v", err)
		return
	}
	logPath := status.GetLogPath()
	if logPath == "" {
		logrus.Errorf("The container has not set log path")
		return
	}
	readLogCtx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	logOptions := logs.NewLogOptions(&v1.PodLogOptions{
		Follow:     false,
	}, time.Now())

	logs.ReadLogs(readLogCtx, logPath, status.GetId(), logOptions, runtimeService, os.Stdout, os.Stderr)




	//logrus.Info("get runtime client success")
	//resp, err:=runtimeClient.ListContainers(context.Background(), &v1.ListContainersRequest{})
	//if err != nil {
	//	logrus.Errorf("List Container failed %v", err)
	//	return
	//}
	//logrus.Println(resp.String())
	//conn, err := bindings.NewConnection(context.Background(), "unix://run/user/1000/podman/podman.sock")
	//defer conn.Done()
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//client, err := containerd.New("/run/containerd/containerd.sock")
	//if err != nil {
	//	fmt.Println("new containerd cli failed %+v", err)
	//	os.Exit(1)
	//}
	//defer client.Close()
}
