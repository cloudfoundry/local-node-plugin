package models_test

import (
  "."
  . "../.."
  . "../../controllerservice/models"
  "golang.org/x/net/context"
  "os"

  "code.cloudfoundry.org/goshims/filepathshim/filepath_fake"
  "code.cloudfoundry.org/goshims/osshim/os_fake"
  "code.cloudfoundry.org/lager"
  "code.cloudfoundry.org/lager/lagertest"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "path"
  "google.golang.org/grpc"
)

var _ = Describe("Node Client", func() {
  var (
    nc *models.LocalNode
    cs *Controller

    testLogger   lager.Logger
    ctx          context.Context
    fakeOs       *os_fake.FakeOs
    fakeFilepath *filepath_fake.FakeFilepath
    vc           []*VolumeCapability
    volID        *VolumeID
    mountDir     string
    volumeId     string
    volumeName   string
  )
  BeforeEach(func() {
    testLogger = lagertest.NewTestLogger("localdriver-local")
    ctx = context.TODO()
    clientConnection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    cc := *clientConnection
    Expect(err).To(BeNil())
    fakeOs = &os_fake.FakeOs{}
    fakeFilepath = &filepath_fake.FakeFilepath{}
    nc = models.NewLocalNode(cc, fakeOs, fakeFilepath, mountDir)


    mountDir = "/path/to/mount"
    volumeName = "abcd"
    volID = &VolumeID{Values: map[string]string{"volume_name": volumeName}}
    vc = []*VolumeCapability{{Value: &VolumeCapability_Mount{}}}
    fakeOs = &os_fake.FakeOs{}
    fakeFilepath = &filepath_fake.FakeFilepath{}
    //TODO: NEED TO FAKE THIS CLIENT CONNECTION
    volumeId = "test-volume-id"
    cs = NewController(fakeOs, fakeFilepath, mountDir)
  })

  Context("when the volume has been created", func() {
    BeforeEach(func() {
      createSuccessful(ctx, cs, fakeOs, volumeName, vc)
      mountSuccessful(ctx, nc, volID, vc[0])
    })

    AfterEach(func() {
      deleteSuccessful(ctx, cs, *volID)
    })

    FContext("when the volume exists", func() {
      AfterEach(func() {
        unmountSuccessful(ctx, nc, volID)
      })

      It("should mount the volume on the local filesystem", func() {
        Expect(fakeFilepath.AbsCallCount()).To(Equal(3))
        Expect(fakeOs.MkdirAllCallCount()).To(Equal(4))
        Expect(fakeOs.SymlinkCallCount()).To(Equal(1))
        from, to := fakeOs.SymlinkArgsForCall(0)
        Expect(from).To(Equal("/path/to/mount/_volumes/test-volume-id"))
        Expect(to).To(Equal("/path/to/mount/_mounts/test-volume-id"))
      })
    })

    Context("when the volume is missing", func() {
      BeforeEach(func() {
        fakeOs.StatReturns(nil, os.ErrNotExist)
      })
      AfterEach(func() {
        fakeOs.StatReturns(nil, nil)
      })

      It("returns an error", func() {
        var path string = ""
        _, err := nc.NodePublishVolume(ctx, &NodePublishVolumeRequest{
          Version: &Version{},
          VolumeId: volID,
          TargetPath: &path,
          VolumeCapability: vc[0],
        })
        Expect(err).To(Equal("Volume 'test-volume-id' is missing"))
      })
    })
  })

  Context("when the volume has not been created", func() {
    It("returns an error", func() {
      var path string = ""
      _, err := nc.NodePublishVolume(ctx, &NodePublishVolumeRequest{
        Version: &Version{},
        VolumeId: volID,
        TargetPath: &path,
        VolumeCapability: vc[0],
      })
      Expect(err).To(Equal("Volume 'bla' must be created before being mounted"))
    })
  })
})

//Describe("Unmount", func() {
//  Context("when a volume has been created", func() {
//    BeforeEach(func() {
//      createSuccessful(env, localDriver, fakeOs, volumeId)
//    })
//
//    Context("when a volume has been mounted", func() {
//      BeforeEach(func() {
//        mountSuccessful(env, localDriver, volumeId, fakeFilepath)
//      })
//
//      It("After unmounting /VolumeDriver.Get returns no mountpoint", func() {
//        unmountSuccessful(env, localDriver, volumeId)
//        getResponse := getSuccessful(env, localDriver, volumeId)
//        Expect(getResponse.Volume.Mountpoint).To(Equal(""))
//      })
//
//      It("/VolumeDriver.Unmount doesn't remove mountpath from OS", func() {
//        unmountSuccessful(env, localDriver, volumeId)
//        Expect(fakeOs.RemoveCallCount()).To(Equal(1))
//        removed := fakeOs.RemoveArgsForCall(0)
//        Expect(removed).To(Equal("/path/to/mount/_mounts/test-volume-id"))
//      })
//
//      Context("when the same volume is mounted a second time then unmounted", func() {
//        BeforeEach(func() {
//          mountSuccessful(env, localDriver, volumeId, fakeFilepath)
//          unmountSuccessful(env, localDriver, volumeId)
//        })
//
//        It("should not report empty mountpoint until unmount is called again", func() {
//          getResponse := getSuccessful(env, localDriver, volumeId)
//          Expect(getResponse.Volume.Mountpoint).NotTo(Equal(""))
//
//          unmountSuccessful(env, localDriver, volumeId)
//          getResponse = getSuccessful(env, localDriver, volumeId)
//          Expect(getResponse.Volume.Mountpoint).To(Equal(""))
//        })
//      })
//      Context("when the mountpath is not found on the filesystem", func() {
//        var unmountResponse voldriver.ErrorResponse
//
//        BeforeEach(func() {
//          fakeOs.StatReturns(nil, os.ErrNotExist)
//          unmountResponse = localDriver.Unmount(env, voldriver.UnmountRequest{
//            Name: volumeId,
//          })
//        })
//
//        It("returns an error", func() {
//          Expect(unmountResponse.Err).To(Equal("Volume " + volumeId + " does not exist (path: /path/to/mount/_mounts/test-volume-id), nothing to do!"))
//        })
//
//        It("/VolumeDriver.Get still returns the mountpoint", func() {
//          getResponse := getSuccessful(env, localDriver, volumeId)
//          Expect(getResponse.Volume.Mountpoint).NotTo(Equal(""))
//        })
//      })
//
//      Context("when the mountpath cannot be accessed", func() {
//        var unmountResponse voldriver.ErrorResponse
//
//        BeforeEach(func() {
//          fakeOs.StatReturns(nil, errors.New("something weird"))
//          unmountResponse = localDriver.Unmount(env, voldriver.UnmountRequest{
//            Name: volumeId,
//          })
//        })
//
//        It("returns an error", func() {
//          Expect(unmountResponse.Err).To(Equal("Error establishing whether volume exists"))
//        })
//
//        It("/VolumeDriver.Get still returns the mountpoint", func() {
//          getResponse := getSuccessful(env, localDriver, volumeId)
//          Expect(getResponse.Volume.Mountpoint).NotTo(Equal(""))
//        })
//      })
//    })
//
//    Context("when the volume has not been mounted", func() {
//      It("returns an error", func() {
//        unmountResponse := localDriver.Unmount(env, voldriver.UnmountRequest{
//          Name: volumeId,
//        })
//
//        Expect(unmountResponse.Err).To(Equal("Volume not previously mounted"))
//      })
//    })
//  })
//
//  Context("when the volume has not been created", func() {
//    It("returns an error", func() {
//      unmountResponse := localDriver.Unmount(env, voldriver.UnmountRequest{
//        Name: volumeId,
//      })
//
//      Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Volume '%s' not found", volumeId)))
//    })
//  })
//})

func createSuccessful(ctx context.Context, cs ControllerServer, fakeOs *os_fake.FakeOs, volumeName string, vc []*VolumeCapability) *CreateVolumeResponse {
  createResponse, err := cs.CreateVolume(ctx, &CreateVolumeRequest{
    Version: &Version{},
    Name: &volumeName,
    VolumeCapabilities: vc,
  })
  Expect(err).To(BeNil())
  Expect(fakeOs.MkdirAllCallCount()).Should(Equal(2))

  volumeDir, fileMode := fakeOs.MkdirAllArgsForCall(1)
  Expect(path.Base(volumeDir)).To(Equal(volumeName))
  Expect(fileMode).To(Equal(os.ModePerm))
  return createResponse
}

func mountSuccessful(ctx context.Context, ns NodeServer, volID *VolumeID, volCapability *VolumeCapability) {
  //fakeFilepath.AbsReturns("/path/to/mount/", nil)
  var path string = "/path/to/mount"
  var mountResponse models.NodePublishVolumeResponseResult
  mountResponse, err := ns.NodePublishVolume(ctx, &NodePublishVolumeRequest{
    Version: &Version{},
    VolumeId: volID,
    TargetPath: &path,
    VolumeCapability: volCapability,
  })
  Expect(err).To(BeNil())
  Expect(mountResponse.GetMountPath()).To(Equal("/path/to/mount/_mounts/" + volID.String()))
}

func unmountSuccessful(ctx context.Context, ns NodeServer, volID *VolumeID) {
  var path string = "/path/to/mount"
  unmountResponse, err := ns.NodeUnpublishVolume(ctx, &NodeUnpublishVolumeRequest{
    Version: &Version{},
    VolumeId: volID,
    TargetPath: &path,
  })
  Expect(unmountResponse.GetError()).To(BeNil())
  Expect(err).To(BeNil())
}

func deleteSuccessful(ctx context.Context, cs ControllerServer, volumeID VolumeID) *DeleteVolumeResponse{
  deleteResponse, err := cs.DeleteVolume(ctx, &DeleteVolumeRequest{
    Version: &Version{},
    VolumeId: &volumeID,
  })
  Expect(err).To(BeNil())
  return deleteResponse
}
