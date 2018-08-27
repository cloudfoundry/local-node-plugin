package node

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	NODE_PLUGIN_ID = "org.cloudfoundry.code.local-node-plugin"
)

type LocalVolume struct {
	csi.Volume
}

//go:generate counterfeiter -o nodefakes/fake_os_helper.go . OsHelper
type OsHelper interface {
	Umask(mask int) (oldmask int)
	Mount(srcPath string, targetPath string) error
	IsMounted(targetPath string) (bool, error)
	Unmount(targetPath string) error
}

type LocalNode struct {
	filepath       filepathshim.Filepath
	os             osshim.Os
	logger         lager.Logger
	volumesRootDir string
	osHelper       OsHelper
	nodeId         string
}

func NewLocalNode(
	os osshim.Os,
	osHelper OsHelper,
	filepath filepathshim.Filepath,
	logger lager.Logger,
	volumeRootDir string,
	nodeId string,
) *LocalNode {
	return &LocalNode{
		os:             os,
		filepath:       filepath,
		logger:         logger,
		volumesRootDir: volumeRootDir,
		osHelper:       osHelper,
		nodeId:         nodeId,
	}
}

func (ln *LocalNode) NodePublishVolume(ctx context.Context, in *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	logger := ln.logger.Session("node-publish-volume")
	logger.Info("start")
	defer logger.Info("end")

	var volId string = in.GetVolumeId()
	if volId == "" {
		errorDescription := "Volume ID is missing in request"
		return nil, grpc.Errorf(codes.InvalidArgument, errorDescription)
	}

	volumePath := ln.volumePath(logger, volId)
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
	logger.Info("mounting-volume", lager.Data{"volume id": volId, "mount point": mountPath})

	mounted, err := ln.osHelper.IsMounted(mountPath)
	if err != nil {
		logger.Error("volume-is-mounted-failed", err)
		errorDescription := "Error checking if volume is mounted"
		return nil, grpc.Errorf(codes.Internal, errorDescription)
	}
	logger.Info("volume-mounted", lager.Data{"value": mounted})

	if mounted {
		logger.Info("unmount", lager.Data{"mountPath": mountPath})
		err := ln.osHelper.Unmount(mountPath)
		if err != nil {
			logger.Error("volume-unmount-failed", err)
			errorDescription := "Error unmounting volume"
			return nil, grpc.Errorf(codes.Internal, errorDescription)
		}
	}

	err = ln.mount(logger, volumePath, mountPath)
	if err != nil {
		logger.Error("mount-volume-failed", err)
		errorDescription := "Error mounting volume"
		return nil, grpc.Errorf(codes.Internal, errorDescription)
	}

	logger.Info("volume-mounted", lager.Data{"volume id": volId, "volume path": volumePath, "mount path": mountPath})
	return &csi.NodePublishVolumeResponse{}, nil
}

func (ln *LocalNode) NodeUnpublishVolume(ctx context.Context, in *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	var volId string = in.GetVolumeId()
	if volId == "" {
		errorDescription := "Volume ID is missing in request"
		return nil, grpc.Errorf(codes.InvalidArgument, errorDescription)
	}

	mountPath := in.GetTargetPath()
	if mountPath == "" {
		errorDescription := "Mount path is missing in request"
		return nil, grpc.Errorf(codes.InvalidArgument, errorDescription)
	}

	ln.logger.Info("unmount", lager.Data{"volume id": volId})

	mounted, err := ln.osHelper.IsMounted(mountPath)
	if err != nil {
		ln.logger.Error("volume-is-mounted-failed", err)
		errorDescription := "Error checking if volume is mounted"
		return nil, grpc.Errorf(codes.Internal, errorDescription)
	}

	ln.logger.Info("volume-mounted", lager.Data{"value": mounted})
	if !mounted {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	ln.logger.Info("umount", lager.Data{"mountPath": mountPath})

	err = ln.osHelper.Unmount(mountPath)
	if err != nil {
		ln.logger.Error("umount-volume-failed", err)
		errorDescription := "Error unmounting volume"
		return nil, grpc.Errorf(codes.Internal, errorDescription)
	}

	err = ln.os.Remove(mountPath)
	if err != nil {
		ln.logger.Error("remove-mount-path-failed", err)
		errorDescription := "Error removing volume mount directory"
		return nil, grpc.Errorf(codes.Internal, errorDescription)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ln *LocalNode) NodeGetId(ctx context.Context, in *csi.NodeGetIdRequest) (*csi.NodeGetIdResponse, error) {
	return &csi.NodeGetIdResponse{
		NodeId: ln.nodeId,
	}, nil
}

func (ln *LocalNode) Probe(ctx context.Context, in *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}

func (ln *LocalNode) NodeStageVolume(ctx context.Context, in *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (ln *LocalNode) NodeUnstageVolume(ctx context.Context, in *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ln *LocalNode) NodeGetCapabilities(ctx context.Context, in *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{Capabilities: []*csi.NodeServiceCapability{}}, nil
}

func (ln *LocalNode) NodeGetInfo(ctx context.Context, in *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: ln.nodeId,
	}, nil
}

func (ln *LocalNode) GetPluginCapabilities(ctx context.Context, in *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{Capabilities: []*csi.PluginCapability{}}, nil
}

func (ln *LocalNode) GetPluginInfo(ctx context.Context, in *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Name:          NODE_PLUGIN_ID,
		VendorVersion: "0.1.0",
	}, nil
}

func (ns *LocalNode) volumePath(logger lager.Logger, volumeId string) string {
	volumesPathRoot := filepath.Join(ns.volumesRootDir, volumeId)
	orig := ns.osHelper.Umask(000)
	defer ns.osHelper.Umask(orig)
	err := ns.os.MkdirAll(volumesPathRoot, os.ModePerm)
	if err != nil {
		panic(err)
	}

	return volumesPathRoot
}

func (ns *LocalNode) mount(logger lager.Logger, volumePath, mountPath string) error {
	err := ns.createVolumesRootifNotExist(logger, mountPath)
	if err != nil {
		logger.Error("create-volumes-root", err)
		return err
	}

	logger.Info("mount", lager.Data{"src": volumePath, "tgt": mountPath})
	return ns.osHelper.Mount(volumePath, mountPath)
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

func (ns *LocalNode) createVolumesRootifNotExist(logger lager.Logger, mountPath string) error {
	mountPath, err := ns.filepath.Abs(mountPath)
	if err != nil {
		logger.Error("abs-failed", err)
		return err
	}

	logger.Debug("mkdir", lager.Data{"mountPath": mountPath})
	orig := ns.osHelper.Umask(000)
	defer ns.osHelper.Umask(orig)

	return ns.os.MkdirAll(mountPath, os.ModePerm)
}
