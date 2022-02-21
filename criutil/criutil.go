package criutil

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/kubernetes/pkg/kubelet/util"
	"time"
)

const (
	defaultTimeout = 2 * time.Second
	// use same message size as cri remote client in kubelet.
	maxMsgSize = 1024 * 1024 * 16
)
var defaultRuntimeEndpoints = []string{"unix:///var/run/dockershim.sock", "unix:///run/containerd/containerd.sock", "unix:///run/crio/crio.sock", "unix:///var/run/cri-dockerd.sock"}

func GetRuntimeClient(context *context.Context) (pb.RuntimeServiceClient, *grpc.ClientConn, error) {
	// Set up a connection to the server.
	conn, err := getRuntimeClientConnection(context)
	if err != nil {
		return nil, nil, errors.Wrap(err, "connect")
	}
	runtimeClient := pb.NewRuntimeServiceClient(conn)
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
		logrus.Debugf("connect using endpoint '%s' with '%s' timeout", endPoint, time.Second*10)
		addr, dialer, err := util.GetAddressAndDialer(endPoint)
		if err != nil {
			if indx == endPointsLen-1 {
				return nil, err
			}
			logrus.Error(err)
			continue
		}
		conn, err = grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*10), grpc.WithDialer(dialer), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)))
		if err != nil {
			errMsg := errors.Wrapf(err, "connect endpoint '%s', make sure you are running as root and the endpoint has been started", endPoint)
			if indx == endPointsLen-1 {
				return nil, errMsg
			}
			logrus.Error(errMsg)
		} else {
			logrus.Debugf("connected successfully using endpoint: %s", endPoint)
			break
		}
	}
	return conn, nil
}
