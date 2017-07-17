package models

import (
  . "../.."
  "golang.org/x/net/context"
  "google.golang.org/grpc"
  "code.cloudfoundry.org/lager"
  "fmt"
  "path/filepath"
  "syscall"
  "os"
  "code.cloudfoundry.org/goshims/filepathshim"
  "code.cloudfoundry.org/goshims/osshim"
  "errors"
)

const VolumesRootDir = "_volumes"

type LocalVolume struct {
  VolumeInfo
}

type LocalNode struct {
  clientConnection grpc.ClientConn
  filepath      filepathshim.Filepath
  mountPathRoot string
  os            osshim.Os
}

func NewLocalNode(conn grpc.ClientConn, os osshim.Os, filepath filepathshim.Filepath, mountPathRoot string) *LocalNode {
  return &LocalNode{
    clientConnection: conn,
    os:            os,
    filepath:      filepath,
    mountPathRoot: mountPathRoot,
  }
}

func (ns *LocalNode) NodePublishVolume(ctx context.Context, in *NodePublishVolumeRequest) (*NodePublishVolumeResponse, error) {
  logger := lager.NewLogger("node-publish-volume")
  var volName string = in.GetVolumeId().GetValues()["volume_name"]
  //if mountRequest.Name == "" {
  //  return voldriver.MountResponse{Err: "Missing mandatory 'volume_name'"}
  //}

  //var vol *LocalVolume
  //var ok bool
  //if vol, ok = d.volumes[mountRequest.Name]; !ok {
  //  return voldriver.MountResponse{Err: fmt.Sprintf("Volume '%s' must be created before being mounted", mountRequest.Name)}
  //}

  volumePath := ns.volumePath(logger, volName)
  fmt.Println(volumePath)

  //exists, err := d.exists(volumePath)
  //if err != nil {
  //  logger.Error("mount-volume-failed", err)
  //  return voldriver.MountResponse{Err: err.Error()}
  //}
  //
  //if !exists {
  //  logger.Error("mount-volume-failed", errors.New("Volume '"+mountRequest.Name+"' is missing"))
  //  return voldriver.MountResponse{Err: "Volume '" + mountRequest.Name + "' is missing"}
  //}

  mountPath := in.GetTargetPath()
  logger.Info("mounting-volume", lager.Data{"id": volName, "mountpoint": mountPath})

  //if vol.MountCount < 1 {
    err := ns.mount(logger, volumePath, mountPath)
  fmt.Println("mounting")
    if err != nil {
      logger.Error("mount-volume-failed", err)
      return &NodePublishVolumeResponse{}, errors.New(fmt.Sprintf("Error mounting volume: %s", err.Error()))
    }
    //vol.Mountpoint = mountPath
  //}

  //vol.MountCount++
  fmt.Println("Response time")
  logger.Info("volume-mounted", lager.Data{"name": volName})

  mountResponse := &NodePublishVolumeResponse{Reply: &NodePublishVolumeResponseResult{
    MountPath: mountPath,
  }}
  return mountResponse, nil
}

//func (d *LocalDriver) Path(env voldriver.Env, pathRequest voldriver.PathRequest) voldriver.PathResponse {
//  logger := env.Logger().Session("path", lager.Data{"volume": pathRequest.Name})
//
//  if pathRequest.Name == "" {
//    return voldriver.PathResponse{Err: "Missing mandatory 'volume_name'"}
//  }
//
//  mountPath, err := d.get(logger, pathRequest.Name)
//  if err != nil {
//    logger.Error("failed-no-such-volume-found", err, lager.Data{"mountpoint": mountPath})
//
//    return voldriver.PathResponse{Err: fmt.Sprintf("Volume '%s' not found", pathRequest.Name)}
//  }
//
//  if mountPath == "" {
//    errText := "Volume not previously mounted"
//    logger.Error("failed-mountpoint-not-assigned", errors.New(errText))
//    return voldriver.PathResponse{Err: errText}
//  }
//
//  return voldriver.PathResponse{Mountpoint: mountPath}
//}

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

func (ns *LocalNode) volumePath(logger lager.Logger, volumeId string) string {
  dir, err := ns.filepath.Abs(ns.mountPathRoot)
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

//func (d *LocalDriver) unmount(logger lager.Logger, name string, mountPath string) voldriver.ErrorResponse {
//  logger = logger.Session("unmount")
//  logger.Info("start")
//  defer logger.Info("end")
//
//  exists, err := d.exists(mountPath)
//  if err != nil {
//    logger.Error("failed-retrieving-mount-info", err, lager.Data{"mountpoint": mountPath})
//    return voldriver.ErrorResponse{Err: "Error establishing whether volume exists"}
//  }
//
//  if !exists {
//    errText := fmt.Sprintf("Volume %s does not exist (path: %s), nothing to do!", name, mountPath)
//    logger.Error("failed-mountpoint-not-found", errors.New(errText))
//    return voldriver.ErrorResponse{Err: errText}
//  }
//
//  d.volumes[name].MountCount--
//  if d.volumes[name].MountCount > 0 {
//    logger.Info("volume-still-in-use", lager.Data{"name": name, "count": d.volumes[name].MountCount})
//    return voldriver.ErrorResponse{}
//  } else {
//    logger.Info("unmount-volume-folder", lager.Data{"mountpath": mountPath})
//    err := d.os.Remove(mountPath)
//    if err != nil {
//      logger.Error("unmount-failed", err)
//      return voldriver.ErrorResponse{Err: fmt.Sprintf("Error unmounting volume: %s", err.Error())}
//    }
//  }
//
//  logger.Info("unmounted-volume")
//
//  d.volumes[name].Mountpoint = ""
//
//  return voldriver.ErrorResponse{}
//}

type NodePublishVolumeResponseResult struct {
  MountPath string
}

func (*NodePublishVolumeResponseResult) isNodePublishVolumeResponse_Reply() {}

func (res *NodePublishVolumeResponseResult) GetMountPath() string {
  return res.MountPath
}