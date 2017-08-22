package node

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	. "github.com/paulcwarren/spec"
	"golang.org/x/net/context"
)

const (
	NODE_PLUGIN_ID = "com.github.jeffpak.local-node-plugin"
)

type LocalVolume struct {
	VolumeInfo
}

type LocalNode struct {
	filepath       filepathshim.Filepath
	os             osshim.Os
	logger         lager.Logger
	volumesRootDir string
}

func NewLocalNode(os osshim.Os, filepath filepathshim.Filepath, logger lager.Logger, volumeRootDir string) *LocalNode {
	return &LocalNode{
		os:             os,
		filepath:       filepath,
		logger:         logger,
		volumesRootDir: volumeRootDir,
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
	logger := ln.logger.Session("node-publish-volume")
	logger.Info("start")
	defer logger.Info("end")

	volID := in.GetVolumeId()
	if volID == nil {
		errorDescription := "Volume id is missing in request"
		return createPublishVolumeErrorResponse(Error_NodePublishVolumeError_INVALID_VOLUME_ID, errorDescription), nil
	}

	var volName string = in.GetVolumeId().GetValues()["volume_name"]

	if volName == "" {
		errorDescription := "Volume name is missing in request"
		return createPublishVolumeErrorResponse(Error_NodePublishVolumeError_INVALID_VOLUME_ID, errorDescription), nil
	}
	volumePath := ln.volumePath(ln.logger, volName)
	logger.Info("volume-path", lager.Data{"value": volumePath})

	vc := in.GetVolumeCapability()
	if vc == nil {
		errorDescription := "Volume capability is missing in request"
		return createPublishVolumeErrorResponse(Error_NodePublishVolumeError_MOUNT_ERROR, errorDescription), nil
	}
	if vc.GetMount() == nil {
		errorDescription := "Volume mount capability is not specified"
		return createPublishVolumeErrorResponse(Error_NodePublishVolumeError_MOUNT_ERROR, errorDescription), nil
	}

	mountPath := in.GetTargetPath()
	ln.logger.Info("mounting-volume", lager.Data{"id": volName, "mountpoint": mountPath})

	exists, _ := ln.exists(mountPath)
	ln.logger.Info("volume-exists", lager.Data{"value": exists})

	if !exists {
		err := ln.mount(ln.logger, volumePath, mountPath)
		if err != nil {
			ln.logger.Error("mount-volume-failed", err)
			errorDescription := fmt.Sprintf("Error mounting volume %s", err.Error())
			return createPublishVolumeErrorResponse(Error_NodePublishVolumeError_MOUNT_ERROR, errorDescription), nil
		}
		ln.logger.Info("volume-mounted", lager.Data{"name": volName, "volume path": volumePath, "mount path": mountPath})
	}

	return createPublishVolumeResultResponse(), nil
}

func (ln *LocalNode) NodeUnpublishVolume(ctx context.Context, in *NodeUnpublishVolumeRequest) (*NodeUnpublishVolumeResponse, error) {
	volID := in.GetVolumeId()
	if volID == nil {
		errorDescription := "Volume id is missing in request"
		return createUnpublishVolumeErrorResponse(Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID, errorDescription), nil
	}

	name := in.GetVolumeId().GetValues()["volume_name"]
	if name == "" {
		errorDescription := "Volume name is missing in request"
		return createUnpublishVolumeErrorResponse(Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID, errorDescription), nil
	}

	ln.logger.Info("unmount", lager.Data{"volume": name})

	mountPath := in.GetTargetPath()
	if mountPath == "" {
		errorDescription := "Mount path is missing in the request"
		return createUnpublishVolumeErrorResponse(Error_NodeUnpublishVolumeError_INVALID_VOLUME_ID, errorDescription), nil
	}

	fi, err := ln.os.Lstat(mountPath)

	if ln.os.IsNotExist(err) {
		return createUnpublishVolumeResultResponse(), nil
	} else if fi.Mode()&os.ModeSymlink == 0 {
		errorDescription := fmt.Sprintf("Mount point '%s' is not a symbolic link", mountPath)
		return createUnpublishVolumeErrorResponse(Error_NodeUnpublishVolumeError_VOLUME_DOES_NOT_EXIST, errorDescription), nil
	}

	err = ln.unmount(ln.logger, mountPath)
	if err != nil {
		errorDescription := err.Error()
		return createUnpublishVolumeErrorResponse(Error_NodeUnpublishVolumeError_UNMOUNT_ERROR, errorDescription), nil
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

func (ns *LocalNode) volumePath(logger lager.Logger, volumeId string) string {
	volumesPathRoot := filepath.Join(ns.volumesRootDir, volumeId)
	orig := syscall.Umask(000)
	defer syscall.Umask(orig)
	err := ns.os.MkdirAll(volumesPathRoot, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return volumesPathRoot
}

func (ns *LocalNode) mount(logger lager.Logger, volumePath, mountPath string) error {
	mountRoot := filepath.Dir(mountPath)
	err := createVolumesRootifNotExist(logger, mountRoot, ns.os)

	if err != nil {
		logger.Error("create-volumes-root", err)
		return err
	}

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

func (ns *LocalNode) exists(path string) (bool, error) {
	_, err := ns.os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func createVolumesRootifNotExist(logger lager.Logger, mountPath string, osShim osshim.Os) error {
	mountPath, err := filepath.Abs(mountPath)
	if err != nil {
		logger.Fatal("abs-failed", err)
	}

	logger.Debug(mountPath)
	_, err = osShim.Stat(mountPath)

	if err != nil {
		if osShim.IsNotExist(err) {
			// Create the directory if not exist
			orig := syscall.Umask(000)
			defer syscall.Umask(orig)
			err = osShim.MkdirAll(mountPath, os.ModePerm)
			if err != nil {
				logger.Error("mkdirall", err)
				return err
			}
		}
	}
	return nil
}
