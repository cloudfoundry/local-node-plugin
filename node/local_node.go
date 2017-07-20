package node

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	. "github.com/jeffpak/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const VolumesRootDir = "_volumes"

type LocalVolume struct {
	VolumeInfo
}

type LocalNode struct {
	clientConnection grpc.ClientConn
	filepath         filepathshim.Filepath
	os               osshim.Os
	logger           lager.Logger
}

func NewLocalNode(conn grpc.ClientConn, os osshim.Os, filepath filepathshim.Filepath, logger lager.Logger) *LocalNode {
	return &LocalNode{
		clientConnection: conn,
		os:               os,
		filepath:         filepath,
		logger:           logger,
	}
}
func createPublishVolumeErrorResponse(errorCode Error_NodePublishVolumeError_NodePublishVolumeErrorCode, errorDescription string) *NodePublishVolumeResponse {
	return &NodePublishVolumeResponse{
		Reply: &NodePublishVolumeResponse_Error{
			Error: &Error{
				Value: &Error_NodePublishVolumeError_{
					NodePublishVolumeError: &Error_NodePublishVolumeError{
						ErrorCode:        errorCode,
						ErrorDescription: errorDescription,
					}}}}}
}

func createPublishVolumeResultResponse() *NodePublishVolumeResponse {
	return &NodePublishVolumeResponse{
		Reply: &NodePublishVolumeResponse_Result_{
			Result: &NodePublishVolumeResponse_Result{},
		},
	}
}

func (ns *LocalNode) NodePublishVolume(ctx context.Context, in *NodePublishVolumeRequest) (*NodePublishVolumeResponse, error) {
	logger := lager.NewLogger("node-publish-volume")
	var volName string = in.GetVolumeId().GetValues()["volume_name"]

	if volName == "" {
		errorDescription := "Volume name is missing in request"
		return createPublishVolumeErrorResponse(Error_NodePublishVolumeError_INVALID_VOLUME_ID, errorDescription), errors.New(errorDescription)
	}
	volumePath := ns.volumePath(logger, volName, in.GetTargetPath())
	mountPath := in.GetTargetPath()
	logger.Info("mounting-volume", lager.Data{"id": volName, "mountpoint": mountPath})

	err := ns.mount(logger, volumePath, mountPath)
	if err != nil {
		logger.Error("mount-volume-failed", err)
		errorDescription := fmt.Sprintf("Error mounting volume %s", err.Error())
		return createPublishVolumeErrorResponse(Error_NodePublishVolumeError_MOUNT_ERROR, errorDescription), errors.New(errorDescription)
	}
	logger.Info("volume-mounted", lager.Data{"name": volName})

	return createPublishVolumeResultResponse(), nil
}

func (d *LocalNode) NodeUnpublishVolume(ctx context.Context, unmountRequest *NodeUnpublishVolumeRequest) (*NodeUnpublishVolumeResponse, error) {
	//  logger := env.Logger().Session("unmount", lager.Data{"volume": unmountRequest.Name})
	//
	//  if unmountRequest.Name == "" {
	//    return voldriver.ErrorResponse{Err: "Missing mandatory 'volume_name'"}
	//  }
	//
	//  mountPath, err := d.get(logger, unmountRequest.Name)
	//  if err != nil {
	//    logger.Error("failed-no-such-volume-found", err, lager.Data{"mountpoint": mountPath})
	//
	//    return voldriver.ErrorResponse{Err: fmt.Sprintf("Volume '%s' not found", unmountRequest.Name)}
	//  }
	//
	//  if mountPath == "" {
	//    errText := "Volume not previously mounted"
	//    logger.Error("failed-mountpoint-not-assigned", errors.New(errText))
	//    return voldriver.ErrorResponse{Err: errText}
	//  }
	//
	//  return d.unmount(logger, unmountRequest.Name, mountPath)
	return &NodeUnpublishVolumeResponse{}, nil
}

func (d *LocalNode) GetNodeID(ctx context.Context, in *GetNodeIDRequest) (*GetNodeIDResponse, error) {
	return &GetNodeIDResponse{
		Reply: &GetNodeIDResponse_Result_{
			Result: &GetNodeIDResponse_Result{}}}, nil
}

func (d *LocalNode) ProbeNode(ctx context.Context, in *ProbeNodeRequest) (*ProbeNodeResponse, error) {
	return &ProbeNodeResponse{
		Reply: &ProbeNodeResponse_Result_{
			Result: &ProbeNodeResponse_Result{}}}, nil
}

func (d *LocalNode) NodeGetCapabilities(ctx context.Context, in *NodeGetCapabilitiesRequest) (*NodeGetCapabilitiesResponse, error) {
	return &NodeGetCapabilitiesResponse{Reply: &NodeGetCapabilitiesResponse_Result_{
		Result: &NodeGetCapabilitiesResponse_Result{
			Capabilities: []*NodeServiceCapability{{
				Type: &NodeServiceCapability_Rpc{
					Rpc: &NodeServiceCapability_RPC{
						Type: NodeServiceCapability_RPC_UNKNOWN}}}}}}}, nil
}

func (ns *LocalNode) volumePath(logger lager.Logger, volumeId string, mountPath string) string {
	dir, err := ns.filepath.Abs(mountPath)
	if err != nil {
		logger.Fatal("abs-failed", err)
	}
	volumesPathRoot := filepath.Join(dir, VolumesRootDir)
	orig := syscall.Umask(000)
	defer syscall.Umask(orig)
	ns.os.MkdirAll(volumesPathRoot, os.ModePerm)

	return filepath.Join(volumesPathRoot, volumeId)
}

func (ns *LocalNode) mount(logger lager.Logger, volumePath, mountPath string) error {
	logger.Info("link", lager.Data{"src": volumePath, "tgt": mountPath})
	orig := syscall.Umask(000)
	defer syscall.Umask(orig)
	return ns.os.Symlink(volumePath, mountPath)
}
