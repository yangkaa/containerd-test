package main

import (
	"bytes"
	"containerd-test/criutil"
	"context"
	"github.com/containers/buildah"
	"github.com/containers/buildah/define"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"
	"k8s.io/kubernetes/pkg/kubelet/kuberuntime/logs"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	_, runtimeConn, err := criutil.GetRuntimeClient(&ctx)
	if err != nil {
		logrus.Errorf("get runtime client failed %v", err)
		return
	}
	defer criutil.CloseConnection(runtimeConn)

	if buildah.InitReexec() {
		return
	}
	unshare.MaybeReexecUsingUserNamespace(false)

	output := &bytes.Buffer{}
	options := define.BuildOptions{
		ContextDirectory: os.Getenv("CONTEXT_DIR"),
		CommonBuildOpts:  &define.CommonBuildOptions{},
		TransientMounts:         []string{os.Getenv("TRANSIENT_MOUNT")},
		Output:                  os.Getenv("OUTPUT"),
		OutputFormat:            buildah.Dockerv2ImageManifest,
		Out:                     output,
		Err:                     output,
		Layers:                  true,
		NoCache:                 true,
		RemoveIntermediateCtrs:  true,
		ForceRmIntermediateCtrs: true,
	}
	storeOptions := storage.StoreOptions{}
	store, err := storage.GetStore(storeOptions)
	logrus.Info("Get Store Start")
	if err !=nil{
		logrus.Errorf("Get Store failed: %+v", err)
		os.Exit(1)
	}
	logrus.Info("Get Store Success")
	// build the image and gather output. log the output if the build part of the test failed
	imageID, imageName, err := imagebuildah.BuildDockerfiles(ctx,store , options, os.Getenv("DOCKERFILE_NAME"))
	if err != nil {
		logrus.Errorf("Build err %v", err)
		output.WriteString("\n" + err.Error())
		os.Exit(1)
	}
	logrus.Info("Build Success")
	outputString := output.String()
	logrus.Infof("imageID [%s] \nout [%s]", imageID, outputString)
	logrus.Infof("imageName : [%s]", imageName.Name())
	dest, err := alltransports.ParseImageName(imageName.Name())
	if err !=nil{
		logrus.Errorf("Parse image name err %v", err)
		os.Exit(1)
	}
	logrus.Infof("dest : [%s]", dest)
	ref, digest, err :=buildah.Push(context.Background(), imageName.Name(), dest, buildah.PushOptions{})
	if err !=nil{
		logrus.Errorf("Push image name err %v", err)
		os.Exit(1)
	}
	logrus.Infof("ref is %+v", ref)
	if ref != nil {
		logrus.Infof("pushed image %q with digest %s", ref, digest.String())
	}

	//pullImage(err, ctx)

	//createResp, err := runtimeClient.CreateContainer(context.Background(), &v1alpha2.CreateContainerRequest{
	//	Config: &v1alpha2.ContainerConfig{
	//		Image: imageSpec,
	//		Args:  []string{"run", "nginx"},
	//	},
	//})
	//if err != nil {
	//	logrus.Errorf("Create Container failed %v", err)
	//	return
	//}
	//logrus.Println("create container", createResp.String())

	//getContainerStatus(err)

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

func pullImage(err error, ctx context.Context) {
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
}

func getContainerStatus(err error) {
	runtimeService, err := remote.NewRemoteRuntimeService(criutil.RuntimeEndpoint, time.Second*3)
	if err != nil {
		logrus.Errorf("New Remote Runtime Service %v", err)
		return
	}
	status, err := runtimeService.ContainerStatus("8acbeb3d0c3a")
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
		Follow: false,
	}, time.Now())

	logs.ReadLogs(readLogCtx, logPath, status.GetId(), logOptions, runtimeService, os.Stdout, os.Stderr)
}
