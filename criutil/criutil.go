package criutil

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/kubernetes/pkg/kubelet/util"
	"time"
)

const (
	defaultTimeout = 2 * time.Second
	// use same message size as cri remote client in kubelet.
	maxMsgSize = 1024 * 1024 * 16
)

var RuntimeEndpoint string
var defaultRuntimeEndpoints = []string{"unix:///var/run/dockershim.sock", "unix:///run/containerd/containerd.sock", "unix:///run/crio/crio.sock", "unix:///var/run/cri-dockerd.sock"}

func GetImageClient(context *context.Context) (v1alpha2.ImageServiceClient, *grpc.ClientConn, error) {
	// Set up a connection to the server.
	conn, err := getImageClientConnection(context)
	if err != nil {
		return nil, nil, errors.Wrap(err, "connect")
	}
	imageClient := v1alpha2.NewImageServiceClient(conn)
	return imageClient, conn, nil
}

func getImageClientConnection(context *context.Context) (*grpc.ClientConn, error) {
	//if ImageEndpoint == "" {
	//	if RuntimeEndpointIsSet && RuntimeEndpoint == "" {
	//		return nil, fmt.Errorf("--image-endpoint is not set")
	//	}
	//	ImageEndpoint = RuntimeEndpoint
	//	ImageEndpointIsSet = RuntimeEndpointIsSet
	//}
	//logrus.Debugf("get image connection")
	//// If no EP set then use the default endpoint types
	//if !ImageEndpointIsSet {
	//	logrus.Warningf("image connect using default endpoints: %v. "+
	//		"As the default settings are now deprecated, you should set the "+
	//		"endpoint instead.", defaultRuntimeEndpoints)
	//	logrus.Debug("Note that performance maybe affected as each default " +
	//		"connection attempt takes n-seconds to complete before timing out " +
	//		"and going to the next in sequence.")
	return getConnection(defaultRuntimeEndpoints)
	//}
	//return getConnection([]string{ImageEndpoint})
}

func GetRuntimeClient(context *context.Context) (v1alpha2.RuntimeServiceClient, *grpc.ClientConn, error) {
	// Set up a connection to the server.
	conn, err := getRuntimeClientConnection(context)
	if err != nil {
		return nil, nil, errors.Wrap(err, "connect")
	}
	runtimeClient := v1alpha2.NewRuntimeServiceClient(conn)
	return runtimeClient, conn, nil
}

func CloseConnection(conn *grpc.ClientConn) error {
	if conn == nil {
		return nil
	}
	return conn.Close()
}

func getRuntimeClientConnection(context *context.Context) (*grpc.ClientConn, error) {
	//if RuntimeEndpointIsSet && RuntimeEndpoint == "" {
	//	return nil, fmt.Errorf("--runtime-endpoint is not set")
	//}
	//logrus.Debug("get runtime connection")
	//// If no EP set then use the default endpoint types
	//if !RuntimeEndpointIsSet {
	//	logrus.Warningf("runtime connect using default endpoints: %v. "+
	//		"As the default settings are now deprecated, you should set the "+
	//		"endpoint instead.", defaultRuntimeEndpoints)
	//	logrus.Debug("Note that performance maybe affected as each default " +
	//		"connection attempt takes n-seconds to complete before timing out " +
	//		"and going to the next in sequence.")
	return getConnection(defaultRuntimeEndpoints)
	//}
	//return getConnection([]string{RuntimeEndpoint})
}

func getConnection(endPoints []string) (*grpc.ClientConn, error) {
	if endPoints == nil || len(endPoints) == 0 {
		return nil, fmt.Errorf("endpoint is not set")
	}
	endPointsLen := len(endPoints)
	var conn *grpc.ClientConn
	for indx, endPoint := range endPoints {
		logrus.Debugf("connect using endpoint '%s' with '%s' timeout", endPoint, time.Second*3)
		addr, dialer, err := util.GetAddressAndDialer(endPoint)
		if err != nil {
			if indx == endPointsLen-1 {
				return nil, err
			}
			logrus.Error(err)
			continue
		}
		conn, err = grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*3), grpc.WithContextDialer(dialer), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)))
		if err != nil {
			errMsg := errors.Wrapf(err, "connect endpoint '%s', make sure you are running as root and the endpoint has been started", endPoint)
			if indx == endPointsLen-1 {
				return nil, errMsg
			}
			logrus.Error(errMsg)
		} else {
			RuntimeEndpoint = endPoint
			logrus.Infof("connected successfully using endpoint: %s", endPoint)
			break
		}
	}
	return conn, nil
}
