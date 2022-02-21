package main

import (
	"containerd-test/criutil"
	"context"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func main() {
	ctx := context.Background()
	_, runtimeConn, err := criutil.GetRuntimeClient(&ctx)
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

	resp, err := imageClient.PullImage(context.Background(), &v1.PullImageRequest{
		Image: &v1.ImageSpec{
			Image: "nginx",
		},
	})
	if err != nil {
		logrus.Errorf("Pull Image failed %v", err)
		return
	}
	logrus.Println(resp.String())

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
