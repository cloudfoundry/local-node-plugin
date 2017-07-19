package node

import (
  "golang.org/x/net/context"
  "google.golang.org/grpc"
  "code.cloudfoundry.org/goshims/filepathshim"
  "code.cloudfoundry.org/goshims/osshim"
  . "github.com/jeffpak/csi"
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
  //logger := lager.NewLogger("node-publish-volume")
  //var volName string = in.GetVolumeId().GetValues()["volume_name"]
  ////if mountRequest.Name == "" {
  ////  return voldriver.MountResponse{Err: "Missing mandatory 'volume_name'"}
  ////}
	//
  ////var vol *LocalVolume
  ////var ok bool
  ////if vol, ok = d.volumes[mountRequest.Name]; !ok {
  ////  return voldriver.MountResponse{Err: fmt.Sprintf("Volume '%s' must be created before being mounted", mountRequest.Name)}
  ////}
	//
  //volumePath := ns.volumePath(logger, volName)
  //fmt.Println(volumePath)
	//
  ////exists, err := d.exists(volumePath)
  ////if err != nil {
  ////  logger.Error("mount-volume-failed", err)
  ////  return voldriver.MountResponse{Err: err.Error()}
  ////}
  ////
  ////if !exists {
  ////  logger.Error("mount-volume-failed", errors.New("Volume '"+mountRequest.Name+"' is missing"))
  ////  return voldriver.MountResponse{Err: "Volume '" + mountRequest.Name + "' is missing"}
  ////}
	//
  //mountPath := in.GetTargetPath()
  //logger.Info("mounting-volume", lager.Data{"id": volName, "mountpoint": mountPath})
	//
  ////if vol.MountCount < 1 {
  //  err := ns.mount(logger, volumePath, mountPath)
  //fmt.Println("mounting")
  //  if err != nil {
  //    logger.Error("mount-volume-failed", err)
  //    return &NodePublishVolumeResponse{}, errors.New(fmt.Sprintf("Error mounting volume: %s", err.Error()))
  //  }
  //  //vol.Mountpoint = mountPath
  ////}
	//
  ////vol.MountCount++
  //fmt.Println("Response time")
  //logger.Info("volume-mounted", lager.Data{"name": volName})
	//
  //mountResponse := &NodePublishVolumeResponse{Reply: &NodePublishVolumeResponseResult{
  //  MountPath: mountPath,
  //}}
  //return mountResponse, nil
  return &NodePublishVolumeResponse{}, nil
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
  return &ProbeNodeResponse{}, nil
}

func (d *LocalNode) NodeGetCapabilities(ctx context.Context, in *NodeGetCapabilitiesRequest) (*NodeGetCapabilitiesResponse, error) {
  return &NodeGetCapabilitiesResponse{Reply: &NodeGetCapabilitiesResponse_Result_{
    Result: &NodeGetCapabilitiesResponse_Result{
      Capabilities: []*NodeServiceCapability{{
        Type: &NodeServiceCapability_Rpc{
          Rpc: &NodeServiceCapability_RPC{
            Type: NodeServiceCapability_RPC_UNKNOWN}}}}}}}, nil
}

//func (ns *LocalNode) volumePath(logger lager.Logger, volumeId string) string {
//  dir, err := ns.filepath.Abs(ns.mountPathRoot)
//  if err != nil {
//    logger.Fatal("abs-failed", err)
//  }
//
//  volumesPathRoot := filepath.Join(dir, VolumesRootDir)
//  orig := syscall.Umask(000)
//  defer syscall.Umask(orig)
//  ns.os.MkdirAll(volumesPathRoot, os.ModePerm)
//
//  return filepath.Join(volumesPathRoot, volumeId)
//}
//
//func (ns *LocalNode) mount(logger lager.Logger, volumePath, mountPath string) error {
//  logger.Info("link", lager.Data{"src": volumePath, "tgt": mountPath})
//  orig := syscall.Umask(000)
//  defer syscall.Umask(orig)
//  return ns.os.Symlink(volumePath, mountPath)
//}
