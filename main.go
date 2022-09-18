package main

import (
	"bytes"
	"containerd-test/criutil"
	"context"
	"encoding/json"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/typeurl"
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

const (
	DockerContainerdSock = "/var/run/docker/containerd/containerd.sock"
	ContainerdSock       = "/run/containerd/containerd.sock"
)

func main() {
	client, ctx, cancel, err := newClient("default", DockerContainerdSock)
	if err != nil {
		logrus.Errorf("new client failed %v", err)
		os.Exit(1)
	}
	defer cancel()
	eventsClient := client.EventService()
	eventsCh, errCh := eventsClient.Subscribe(ctx)
	for {
		var e *events.Envelope
		select {
		case e = <-eventsCh:
		case err = <-errCh:
			return
		}
		if e != nil {
			var out []byte
			if e.Event != nil {
				v, err := typeurl.UnmarshalAny(e.Event)
				if err != nil {
					log.G(ctx).WithError(err).Warn("cannot unmarshal an event from Any")
					continue
				}
				out, err = json.Marshal(v)
				if err != nil {
					log.G(ctx).WithError(err).Warn("cannot marshal Any into JSON")
					continue
				}
			}
			if _, err := fmt.Fprintln(
				os.Stdout,
				e.Timestamp,
				e.Namespace,
				e.Topic,
				string(out),
			); err != nil {
				return
			}

		}
	}
}

func newClient(namespace, address string, opts ...containerd.ClientOpt) (*containerd.Client, context.Context, context.CancelFunc, error) {
	ctx := context.Background()
	ctx = namespaces.WithNamespace(ctx, namespace)
	client, err := containerd.New(address, opts...)
	if err != nil {
		return nil, nil, nil, err
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	return client, ctx, cancel, nil
}

func old_main() bool {
	ctx := context.Background()
	_, runtimeConn, err := criutil.GetRuntimeClient(&ctx)
	if err != nil {
		logrus.Errorf("get runtime client failed %v", err)
		return true
	}
	defer criutil.CloseConnection(runtimeConn)

	if buildah.InitReexec() {
		return true
	}
	unshare.MaybeReexecUsingUserNamespace(false)

	output := &bytes.Buffer{}
	options := define.BuildOptions{
		ContextDirectory:        os.Getenv("CONTEXT_DIR"),
		CommonBuildOpts:         &define.CommonBuildOptions{},
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
	if err != nil {
		logrus.Errorf("Get Store failed: %+v", err)
		os.Exit(1)
	}
	logrus.Info("Get Store Success")
	// build the image and gather output. log the output if the build part of the test failed
	imageID, imageName, err := imagebuildah.BuildDockerfiles(ctx, store, options, os.Getenv("DOCKERFILE_NAME"))
	if err != nil {
		logrus.Errorf("Build err %v", err)
		output.WriteString("\n" + err.Error())
		os.Exit(1)
	}
	logrus.Info("Build Success")
	outputString := output.String()
	logrus.Infof("imageID [%s] \nout [%s]", imageID, outputString)
	logrus.Infof("imageName : [%s] imageName.String() [%s]", imageName.Name(), imageName.Name())
	dest, err := alltransports.ParseImageName(imageName.String())
	if err != nil {
		logrus.Errorf("Parse image name err %v", err)
		os.Exit(1)
	}
	logrus.Infof("dest : [%s]", dest)
	ref, digest, err := buildah.Push(context.Background(), imageName.Name(), dest, buildah.PushOptions{})
	if err != nil {
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
	return false
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
