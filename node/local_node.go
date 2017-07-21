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
)

const VolumesRootDir = "_volumes"

type LocalVolume struct {
	VolumeInfo
}

type LocalNode struct {
	filepath filepathshim.Filepath
	os       osshim.Os
	logger   lager.Logger
}

func NewLocalNode(os osshim.Os, filepath filepathshim.Filepath, logger lager.Logger) *LocalNode {
	return &LocalNode{
		os:       os,
		filepath: filepath,
		logger:   logger,
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

func createUnpublishVolumeErrorResponse(errorCode Error_NodeUnpublishVolumeError_NodeUnpublishVolumeErrorCode, errorDescription string) *NodeUnpublishVolumeResponse {
	return &NodeUnpublishVolumeResponse{
		Reply: &NodeUnpublishVolumeResponse_Error{
			Error: &Error{
				Value: &Error_NodeUnpublishVolumeError_{
					NodeUnpublishVolumeError: &Error_NodeUnpublishVolumeError{
						ErrorCode:        errorCode,
						ErrorDescription: errorDescription,
					}}}}}
}

func createUnpublishVolumeResultResponse() *NodeUnpublishVolumeResponse {
	return &NodeUnpublishVolumeResponse{
		Reply: &NodeUnpublishVolumeResponse_Result_{
			Result: &NodeUnpublishVolumeResponse_Result{},
		},
	}
}
func (ln *LocalNode) NodePublishVolume(ctx context.Context, in *NodePublishVolumeRequest) (*NodePublishVolumeResponse, error) {
	var volName string = in.GetVolumeId().GetValues()["volume_name"]

	if volName == "" {
		errorDescription := "Volume name is missing in request"
		return createPublishVolumeErrorResponse(Error_NodePublishVolumeError_INVALID_VOLUME_ID, errorDescription), errors.New(errorDescription)
	}
	volumePath := ln.volumePath(ln.logger, volName, in.GetTargetPath())
	mountPath := in.GetTargetPath()
	ln.logger.Info("mounting-volume", lager.Data{"id": volName, "mountpoint": mountPath})

	err := ln.mount(ln.logger, volumePath, mountPath)
	if err != nil {
		ln.logger.Error("mount-volume-failed", err)
		errorDescription := fmt.Sprintf("Error mounting volume %s", err.Error())
		return createPublishVolumeErrorResponse(Error_NodePublishVolumeError_MOUNT_ERROR, errorDescription), errors.New(errorDescription)
	}
	ln.logger.Info("volume-mounted", lager.Data{"name": volName})

	return createPublishVolumeResultResponse(), nil
}

func (ln *LocalNode) NodeUnpublishVolume(ctx context.Context, in *NodeUnpublishVolumeRequest) (*NodeUnpublishVolumeResponse, error) {
	name := in.GetVolumeId().GetValues()["volume_name"]
	if name == "" {
		errorMsg := "Volume name is missing in request"
		return createUnpublishVolumeErrorResponse(Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID, errorMsg), errors.New(errorMsg)
	}

	ln.logger.Info("unmount", lager.Data{"volume": name})

	mountPoint := in.GetTargetPath()
	fi, err := ln.os.Lstat(mountPoint)

	if ln.os.IsNotExist(err) {
		errorMsg := fmt.Sprintf("Mount point '%s' not found", mountPoint)
		return createUnpublishVolumeErrorResponse(Error_NodeUnpublishVolumeError_VOLUME_DOES_NOT_EXIST, errorMsg), errors.New(errorMsg)
	} else if fi.Mode()&os.ModeSymlink == 0 {
		errorMsg := fmt.Sprintf("Mount point '%s' is not a symbolic link", mountPoint)
		return createUnpublishVolumeErrorResponse(Error_NodeUnpublishVolumeError_VOLUME_DOES_NOT_EXIST, errorMsg), errors.New(errorMsg)
	}

	mountPath := in.GetTargetPath()
	if mountPath == "" {
		errorMsg := "Mount path is missing in the request"
		return createUnpublishVolumeErrorResponse(Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID, errorMsg), errors.New(errorMsg)
	}

	err = ln.unmount(ln.logger, mountPath)
	if err != nil {
		errorDescription := fmt.Sprintf("Error unmounting volume %s", err.Error())
		return createUnpublishVolumeErrorResponse(Error_NodeUnpublishVolumeError_UNMOUNT_ERROR, errorDescription), errors.New(errorDescription)
	}
	return createUnpublishVolumeResultResponse(), nil
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
			Capabilities: []*NodeServiceCapability{}}}}, nil
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

func (ns *LocalNode) unmount(logger lager.Logger, mountPath string) error {
	logger.Info("unlink", lager.Data{"tgt": mountPath})
	orig := syscall.Umask(000)
	defer syscall.Umask(orig)
	return ns.os.Remove(mountPath)
}
