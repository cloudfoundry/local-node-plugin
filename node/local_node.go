package node

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	. "github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

func NewLocalNode(os osshim.Os, filepath filepathshim.Filepath, logger lager.Logger, volumeRootDir string) NodeServer {
	return &LocalNode{
		os:             os,
		filepath:       filepath,
		logger:         logger,
		volumesRootDir: volumeRootDir,
	}
}
func (ln *LocalNode) NodePublishVolume(ctx context.Context, in *NodePublishVolumeRequest) (*NodePublishVolumeResponse, error) {
	logger := ln.logger.Session("node-publish-volume")
	logger.Info("start")
	defer logger.Info("end")

	var volId string = in.GetVolumeId()

	if volId == "" {
		errorDescription := "Volume ID is missing in request"
		return nil, grpc.Errorf(codes.InvalidArgument, errorDescription)
	}

	volumePath := ln.volumePath(ln.logger, volId)
	logger.Info("volume-path", lager.Data{"value": volumePath})

	vc := in.GetVolumeCapability()
	if vc == nil {
		errorDescription := "Volume capability is missing in request"
		return nil, grpc.Errorf(codes.InvalidArgument, errorDescription)
	}
	if vc.GetMount() == nil {
		errorDescription := "Volume mount capability is not specified"
		return nil, grpc.Errorf(codes.InvalidArgument, errorDescription)
	}

	mountPath := in.GetTargetPath()
	ln.logger.Info("mounting-volume", lager.Data{"volume id": volId, "mount point": mountPath})

	exists, _ := ln.exists(mountPath)
	ln.logger.Info("volume-exists", lager.Data{"value": exists})

	if !exists {
		err := ln.mount(ln.logger, volumePath, mountPath)
		if err != nil {
			ln.logger.Error("mount-volume-failed", err)
			errorDescription := "Error mounting volume"
			return nil, grpc.Errorf(codes.Internal, errorDescription)
		}
		ln.logger.Info("volume-mounted", lager.Data{"volume id": volId, "volume path": volumePath, "mount path": mountPath})
	}

	return &NodePublishVolumeResponse{}, nil
}

func (ln *LocalNode) NodeUnpublishVolume(ctx context.Context, in *NodeUnpublishVolumeRequest) (*NodeUnpublishVolumeResponse, error) {
	var volId string = in.GetVolumeId()

	if volId == "" {
		errorDescription := "Volume ID is missing in request"
		return nil, grpc.Errorf(codes.InvalidArgument, errorDescription)
	}

	ln.logger.Info("unmount", lager.Data{"volume id": volId})

	mountPath := in.GetTargetPath()
	if mountPath == "" {
		errorDescription := "Mount path is missing in the request"
		return nil, grpc.Errorf(codes.InvalidArgument, errorDescription)
	}

	fi, err := ln.os.Lstat(mountPath)

	if ln.os.IsNotExist(err) {
		return &NodeUnpublishVolumeResponse{}, nil
	} else if fi.Mode()&os.ModeSymlink == 0 {
		errorDescription := fmt.Sprintf("Mount point '%s' is not a symbolic link", mountPath)
		return nil, grpc.Errorf(codes.InvalidArgument, errorDescription)
	}

	err = ln.unmount(ln.logger, mountPath)
	if err != nil {
		errorDescription := err.Error()
		return nil, grpc.Errorf(codes.Internal, errorDescription)
	}
	return &NodeUnpublishVolumeResponse{}, nil
}

func (d *LocalNode) GetNodeID(ctx context.Context, in *GetNodeIDRequest) (*GetNodeIDResponse, error) {
	return &GetNodeIDResponse{}, nil
}

func (d *LocalNode) NodeProbe(ctx context.Context, in *NodeProbeRequest) (*NodeProbeResponse, error) {
	return &NodeProbeResponse{}, nil
}

func (d *LocalNode) NodeGetCapabilities(ctx context.Context, in *NodeGetCapabilitiesRequest) (*NodeGetCapabilitiesResponse, error) {
	return &NodeGetCapabilitiesResponse{Capabilities: []*NodeServiceCapability{}}, nil
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
