package node_test

import (

  "golang.org/x/net/context"
  "code.cloudfoundry.org/goshims/filepathshim/filepath_fake"
  "code.cloudfoundry.org/goshims/osshim/os_fake"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "google.golang.org/grpc"


  . "github.com/jeffpak/csi"
  "github.com/jeffpak/local-node-plugin/node"
  "time"
)

var _ = Describe("Node Client", func() {
  var (
    nc *node.LocalNode
    //cs *controller.Controller
		//
    //testLogger   lager.Logger
    context          context.Context
    fakeOs       *os_fake.FakeOs
    fakeFilepath *filepath_fake.FakeFilepath
    vc           *VolumeCapability
    volID        *VolumeID
    mountDir     string
    //volumeId     string
    volumeName   string
    err          error
  )
  BeforeEach(func() {
    //testLogger = lagertest.NewTestLogger("localdriver-local")
    context = &DummyContext{}
    clientConnection, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
    cc := *clientConnection
    //Expect(err).To(BeNil())
    fakeOs = &os_fake.FakeOs{}
    fakeFilepath = &filepath_fake.FakeFilepath{}
    nc = node.NewLocalNode(cc, fakeOs, fakeFilepath, mountDir)
		//
		//
    //mountDir = "/path/to/mount"
    volumeName = "abcd"
    volID = &VolumeID{Values: map[string]string{"volume_name": volumeName}}
    vc = &VolumeCapability{Value: &VolumeCapability_Mount{}}
    //fakeOs = &os_fake.FakeOs{}
    //fakeFilepath = &filepath_fake.FakeFilepath{}
    ////TODO: NEED TO FAKE THIS CLIENT CONNECTION
    //volumeId = "test-volume-id"
    //cs = controller.NewController(fakeOs, fakeFilepath, mountDir)
  })
  Describe("NodePublishVolume", func() {
    var (
      request *NodePublishVolumeRequest
      expectedResponse *NodePublishVolumeResponse
    )
    Context("when NodePublishVolume is called with a NodePublishVolume", func() {
      BeforeEach(func() {
        request = &NodePublishVolumeRequest{
          Version: &Version{Major: 0, Minor: 0, Patch: 1},
          VolumeId: volID,
          VolumeMetadata: &VolumeMetadata{},
          TargetPath: "unpublish-path",
          VolumeCapability: vc,
          Readonly: true,
        }
      })
      JustBeforeEach(func() {
        expectedResponse, err = nc.NodePublishVolume(context, request)
      })
      It("should return a NodePublishVolumeResponse", func() {
        Expect(expectedResponse).NotTo(BeNil())
      })
    })

  //Context("when the volume has been created", func() {
  //  BeforeEach(func() {
  //    createSuccessful(ctx, cs, fakeOs, volumeName, vc)
  //    mountSuccessful(ctx, nc, volID, vc[0])
  //  })
	//
  //  AfterEach(func() {
  //    deleteSuccessful(ctx, cs, *volID)
  //  })
	//
  //  FContext("when the volume exists", func() {
  //    AfterEach(func() {
  //      unmountSuccessful(ctx, nc, volID)
  //    })
	//
  //    It("should mount the volume on the local filesystem", func() {
  //      Expect(fakeFilepath.AbsCallCount()).To(Equal(3))
  //      Expect(fakeOs.MkdirAllCallCount()).To(Equal(4))
  //      Expect(fakeOs.SymlinkCallCount()).To(Equal(1))
  //      from, to := fakeOs.SymlinkArgsForCall(0)
  //      Expect(from).To(Equal("/path/to/mount/_volumes/test-volume-id"))
  //      Expect(to).To(Equal("/path/to/mount/_mounts/test-volume-id"))
  //    })
  //  })
	//
  //  Context("when the volume is missing", func() {
  //    BeforeEach(func() {
  //      fakeOs.StatReturns(nil, os.ErrNotExist)
  //    })
  //    AfterEach(func() {
  //      fakeOs.StatReturns(nil, nil)
  //    })
	//
  //    It("returns an error", func() {
  //      var path string = ""
  //      _, err := nc.NodePublishVolume(ctx, &NodePublishVolumeRequest{
  //        Version: &Version{},
  //        VolumeId: volID,
  //        TargetPath: &path,
  //        VolumeCapability: vc[0],
  //      })
  //      Expect(err).To(Equal("Volume 'test-volume-id' is missing"))
  //    })
  //  })
  //})
	//
  //Context("when the volume has not been created", func() {
  //  It("returns an error", func() {
  //    var path string = ""
  //    _, err := nc.NodePublishVolume(ctx, &NodePublishVolumeRequest{
  //      Version: &Version{},
  //      VolumeId: volID,
  //      TargetPath: &path,
  //      VolumeCapability: vc[0],
  //    })
  //    Expect(err).To(Equal("Volume 'bla' must be created before being mounted"))
  //  })
  //})
  })

  Describe("NodeUnpublishVolume", func() {
    var (
      request *NodeUnpublishVolumeRequest
      expectedResponse *NodeUnpublishVolumeResponse
    )
    Context("when NodeUnpublishVolume is called with a NodeUnpublishVolume", func() {
      BeforeEach(func() {
        request = &NodeUnpublishVolumeRequest{
          &Version{Major: 0, Minor: 0, Patch: 1},
          volID,
          &VolumeMetadata{},
          "unpublish-path",
        }
      })
      JustBeforeEach(func() {
        expectedResponse, err = nc.NodeUnpublishVolume(context, request)
      })
      It("should return a NodeUnpublishVolumeResponse", func() {
        Expect(expectedResponse).NotTo(BeNil())
      })
    })
  })

  Describe("GetNodeID", func() {
    var (
      request *GetNodeIDRequest
      expectedResponse *GetNodeIDResponse
    )
    Context("when GetNodeID is called with a GetNodeIDRequest", func() {
      BeforeEach(func() {
        request = &GetNodeIDRequest{
          &Version{Major: 0, Minor: 0, Patch: 1},
        }
      })
      JustBeforeEach(func() {
        expectedResponse, err = nc.GetNodeID(context, request)
      })
      It("should return a GetNodeIDResponse that has a result with no node ID", func() {
        Expect(expectedResponse).NotTo(BeNil())
        Expect(expectedResponse.GetResult()).NotTo(BeNil())
        Expect(expectedResponse.GetResult().GetNodeId()).To(BeNil())
        Expect(err).To(BeNil())
      })
    })
  })

  Describe("ProbeNode", func() {
    var (
      request *ProbeNodeRequest
      expectedResponse *ProbeNodeResponse
    )
    Context("when ProbeNode is called with a ProbeNodeRequest", func() {
      BeforeEach(func() {
        request = &ProbeNodeRequest{
          &Version{Major: 0, Minor: 0, Patch: 1},
        }
      })
      JustBeforeEach(func() {
        expectedResponse, err = nc.ProbeNode(context, request)
      })
      It("should return a ProbeNodeResponse", func() {
        Expect(expectedResponse).NotTo(BeNil())
        Expect(expectedResponse.GetResult()).NotTo(BeNil())
        Expect(err).To(BeNil())
      })
    })
  })

  Describe("NodeGetCapabilities", func() {
    var (
      request *NodeGetCapabilitiesRequest
      expectedResponse *NodeGetCapabilitiesResponse
    )
    Context("when NodeGetCapabilities is called with a NodeGetCapabilitiesRequest", func() {
      BeforeEach(func() {
        request = &NodeGetCapabilitiesRequest{
          &Version{Major: 0, Minor: 0, Patch: 1},
        }
      })
      JustBeforeEach(func() {
        expectedResponse, err = nc.NodeGetCapabilities(context, request)
      })

      It("should return a ControllerGetCapabilitiesResponse with UNKNOWN specified", func() {
        Expect(expectedResponse).NotTo(BeNil())
        capabilities := expectedResponse.GetResult().GetCapabilities()
        Expect(capabilities).To(HaveLen(1))
        Expect(capabilities[0].GetRpc().GetType()).To(Equal(NodeServiceCapability_RPC_UNKNOWN))
        Expect(err).To(BeNil())
      })
    })
  })

})

//func createSuccessful(ctx context.Context, cs ControllerServer, fakeOs *os_fake.FakeOs, volumeName string, vc []*VolumeCapability) *CreateVolumeResponse {
//  createResponse, err := cs.CreateVolume(ctx, &CreateVolumeRequest{
//    Version: &Version{},
//    Name: &volumeName,
//    VolumeCapabilities: vc,
//  })
//  Expect(err).To(BeNil())
//  Expect(fakeOs.MkdirAllCallCount()).Should(Equal(2))
//
//  volumeDir, fileMode := fakeOs.MkdirAllArgsForCall(1)
//  Expect(path.Base(volumeDir)).To(Equal(volumeName))
//  Expect(fileMode).To(Equal(os.ModePerm))
//  return createResponse
//}
//
//func mountSuccessful(ctx context.Context, ns NodeServer, volID *VolumeID, volCapability *VolumeCapability) {
//  //fakeFilepath.AbsReturns("/path/to/mount/", nil)
//  var path string = "/path/to/mount"
//  var mountResponse node.NodePublishVolumeResponseResult
//  mountResponse, err := ns.NodePublishVolume(ctx, &NodePublishVolumeRequest{
//    Version: &Version{},
//    VolumeId: volID,
//    TargetPath: &path,
//    VolumeCapability: volCapability,
//  })
//  Expect(err).To(BeNil())
//  Expect(mountResponse.GetMountPath()).To(Equal("/path/to/mount/_mounts/" + volID.String()))
//}
//
//func unmountSuccessful(ctx context.Context, ns NodeServer, volID *VolumeID) {
//  var path string = "/path/to/mount"
//  unmountResponse, err := ns.NodeUnpublishVolume(ctx, &NodeUnpublishVolumeRequest{
//    Version: &Version{},
//    VolumeId: volID,
//    TargetPath: &path,
//  })
//  Expect(unmountResponse.GetError()).To(BeNil())
//  Expect(err).To(BeNil())
//}
//
//func deleteSuccessful(ctx context.Context, cs ControllerServer, volumeID VolumeID) *DeleteVolumeResponse{
//  deleteResponse, err := cs.DeleteVolume(ctx, &DeleteVolumeRequest{
//    Version: &Version{},
//    VolumeId: &volumeID,
//  })
//  Expect(err).To(BeNil())
//  return deleteResponse
//}

type DummyContext struct {}

func (*DummyContext) Deadline() (deadline time.Time, ok bool) { return time.Time{}, false }

func (*DummyContext) Done() <-chan struct{} {return nil}

func (*DummyContext) Err() (error){ return nil }

func (*DummyContext) Value(key interface{}) interface{} {return nil}
