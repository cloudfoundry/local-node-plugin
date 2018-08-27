package node_test

import (
	"errors"
	"path/filepath"

	"code.cloudfoundry.org/goshims/filepathshim/filepath_fake"
	"code.cloudfoundry.org/goshims/osshim/os_fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"

	"os"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/local-node-plugin/node"
	"code.cloudfoundry.org/local-node-plugin/node/nodefakes"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Node Client", func() {
	var (
		context          context.Context
		err              error
		fakeFilepath     *filepath_fake.FakeFilepath
		fakeOs           *os_fake.FakeOs
		fakeOsHelper     *nodefakes.FakeOsHelper
		fileInfo         *FakeFileInfo
		localNode        *node.LocalNode
		mountPath        string
		publishResp      *csi.NodePublishVolumeResponse
		testLogger       lager.Logger
		volumeCapability *csi.VolumeCapability
		volumeId         string
		volumesRoot      string
	)

	BeforeEach(func() {
		volumesRoot = "/tmp/_volumes"
		volumeId = "test-volume-id"
		mountPath = "/path/to/mount/_mounts/test-volume-id"

		testLogger = lagertest.NewTestLogger("localdriver-local")
		context = &DummyContext{}

		fakeOs = &os_fake.FakeOs{}
		fakeFilepath = &filepath_fake.FakeFilepath{}
		fakeFilepath.AbsReturns(mountPath, nil)

		fakeOsHelper = &nodefakes.FakeOsHelper{}

		localNode = node.NewLocalNode(fakeOs, fakeOsHelper, fakeFilepath, testLogger, volumesRoot, "some-node-id")
		volumeCapability = &csi.VolumeCapability{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{}}

		fileInfo = newFakeFileInfo()
		fileInfo.StubMode(os.ModeSymlink)
	})

	Describe("NodePublishVolume", func() {
		Context("when the volume and mount directories do not exist", func() {
			BeforeEach(func() {
				fakeOsHelper.IsMountedReturns(false, nil)
			})

			It("creates the required directories and mounts the volume", func() {
				publishResp, err = localNode.NodePublishVolume(context, &csi.NodePublishVolumeRequest{
					VolumeId:         volumeId,
					TargetPath:       mountPath,
					VolumeCapability: volumeCapability,
					Readonly:         false,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(*publishResp).To(Equal(csi.NodePublishVolumeResponse{}))

				Expect(fakeOsHelper.UnmountCallCount()).To(Equal(0))

				Expect(fakeOs.MkdirAllCallCount()).To(Equal(2))
				srcPath, _ := fakeOs.MkdirAllArgsForCall(0)
				Expect(srcPath).To(Equal(filepath.Join(volumesRoot, volumeId)))
				tgtPath, _ := fakeOs.MkdirAllArgsForCall(1)
				Expect(tgtPath).To(Equal(mountPath))

				Expect(fakeOsHelper.MountCallCount()).To(Equal(1))
				from, to := fakeOsHelper.MountArgsForCall(0)
				Expect(from).To(Equal(filepath.Join(volumesRoot, volumeId)))
				Expect(to).To(Equal(mountPath))
			})
		})

		Context("when the mount path is mounted", func() {
			BeforeEach(func() {
				fakeOsHelper.IsMountedReturns(true, nil)
			})

			It("unmounts the destination directory, creates the volume directory and bind mounts the volume path to the mount path", func() {
				publishResp, err = localNode.NodePublishVolume(context, &csi.NodePublishVolumeRequest{
					VolumeId:         volumeId,
					TargetPath:       mountPath,
					VolumeCapability: volumeCapability,
					Readonly:         false,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(*publishResp).To(Equal(csi.NodePublishVolumeResponse{}))

				Expect(fakeOsHelper.UnmountCallCount()).To(Equal(1))
				tgtPath := fakeOsHelper.UnmountArgsForCall(0)
				Expect(tgtPath).To(Equal(mountPath))

				Expect(fakeOs.MkdirAllCallCount()).To(Equal(2))
				srcPath, _ := fakeOs.MkdirAllArgsForCall(0)
				Expect(srcPath).To(Equal(filepath.Join(volumesRoot, volumeId)))
				tgtPath, _ = fakeOs.MkdirAllArgsForCall(1)
				Expect(tgtPath).To(Equal(mountPath))

				Expect(fakeOsHelper.MountCallCount()).To(Equal(1))
				from, to := fakeOsHelper.MountArgsForCall(0)
				Expect(from).To(Equal(filepath.Join(volumesRoot, volumeId)))
				Expect(to).To(Equal(mountPath))
			})
		})

		Context("failure cases", func() {
			Context("when the volume id is missing", func() {
				BeforeEach(func() {
					volumeId = ""
				})

				It("returns an error", func() {
					_, err = localNode.NodePublishVolume(context, &csi.NodePublishVolumeRequest{
						VolumeId:         volumeId,
						TargetPath:       mountPath,
						VolumeCapability: volumeCapability,
						Readonly:         false,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
					Expect(grpcStatus.Message()).To(Equal("Volume ID is missing in request"))
				})
			})

			Context("when the volume capability is missing", func() {
				BeforeEach(func() {
					volumeCapability = nil
				})

				It("returns an error", func() {
					_, err = localNode.NodePublishVolume(context, &csi.NodePublishVolumeRequest{
						VolumeId:         volumeId,
						TargetPath:       mountPath,
						VolumeCapability: volumeCapability,
						Readonly:         false,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
					Expect(grpcStatus.Message()).To(Equal("Volume capability is missing in request"))
				})
			})

			Context("When the volume capability is not mount capability", func() {
				BeforeEach(func() {
					volumeCapability = &csi.VolumeCapability{}
				})

				It("returns an error", func() {
					_, err = localNode.NodePublishVolume(context, &csi.NodePublishVolumeRequest{
						VolumeId:         volumeId,
						TargetPath:       mountPath,
						VolumeCapability: volumeCapability,
						Readonly:         false,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
					Expect(grpcStatus.Message()).To(Equal("Volume mount capability is not specified"))
				})
			})

			Context("when checking if mounted fails", func() {
				BeforeEach(func() {
					fakeOsHelper.IsMountedReturns(false, errors.New("failed to check if mounted"))
				})

				It("returns an error", func() {
					_, err = localNode.NodePublishVolume(context, &csi.NodePublishVolumeRequest{
						VolumeId:         volumeId,
						TargetPath:       mountPath,
						VolumeCapability: volumeCapability,
						Readonly:         false,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.Internal))
					Expect(grpcStatus.Message()).To(Equal("Error checking if volume is mounted"))
				})
			})

			Context("when the volume cannot be unmounted", func() {
				BeforeEach(func() {
					fakeOsHelper.IsMountedReturns(true, nil)
					fakeOsHelper.UnmountReturns(errors.New("failed to unmount volume"))
				})

				It("returns an error", func() {
					_, err = localNode.NodePublishVolume(context, &csi.NodePublishVolumeRequest{
						VolumeId:         volumeId,
						TargetPath:       mountPath,
						VolumeCapability: volumeCapability,
						Readonly:         false,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.Internal))
					Expect(grpcStatus.Message()).To(Equal("Error unmounting volume"))
				})
			})

			Context("When the volume mount fails", func() {
				BeforeEach(func() {
					fakeOsHelper.MountReturns(errors.New("failed to mount volume"))
				})

				It("returns an error", func() {
					_, err = localNode.NodePublishVolume(context, &csi.NodePublishVolumeRequest{
						VolumeId:         volumeId,
						TargetPath:       mountPath,
						VolumeCapability: volumeCapability,
						Readonly:         false,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.Internal))
					Expect(grpcStatus.Message()).To(Equal("Error mounting volume"))
				})
			})
		})
	})

	Describe("NodeUnpublishVolume", func() {
		Context("when a volume has been mounted", func() {
			BeforeEach(func() {
				fakeOsHelper.IsMountedReturns(true, nil)
			})

			It("Unmounts the volume", func() {
				unpublishResp, err := localNode.NodeUnpublishVolume(context, &csi.NodeUnpublishVolumeRequest{
					VolumeId:   volumeId,
					TargetPath: mountPath,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(*unpublishResp).To(Equal(csi.NodeUnpublishVolumeResponse{}))

				Expect(fakeOsHelper.UnmountCallCount()).To(Equal(1))
				dstPath := fakeOsHelper.UnmountArgsForCall(0)
				Expect(dstPath).To(Equal(mountPath))

				Expect(fakeOs.RemoveCallCount()).To(Equal(1))
				dstPath = fakeOs.RemoveArgsForCall(0)
				Expect(dstPath).To(Equal(mountPath))
			})
		})

		Context("when a volume is not mounted", func() {
			BeforeEach(func() {
				fakeOsHelper.IsMountedReturns(false, nil)
			})

			It("exits early and does not return an error", func() {
				unpublishResp, err := localNode.NodeUnpublishVolume(context, &csi.NodeUnpublishVolumeRequest{
					VolumeId:   volumeId,
					TargetPath: mountPath,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(*unpublishResp).To(Equal(csi.NodeUnpublishVolumeResponse{}))

				Expect(fakeOsHelper.UnmountCallCount()).To(Equal(0))

				Expect(fakeOs.RemoveCallCount()).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			Context("when the volume id is missing", func() {
				BeforeEach(func() {
					volumeId = ""
				})

				It("returns an error", func() {
					_, err := localNode.NodeUnpublishVolume(context, &csi.NodeUnpublishVolumeRequest{
						VolumeId:   volumeId,
						TargetPath: mountPath,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
					Expect(grpcStatus.Message()).To(Equal("Volume ID is missing in request"))
				})
			})

			Context("when the mount path is missing", func() {
				BeforeEach(func() {
					mountPath = ""
				})

				It("returns an error", func() {
					_, err := localNode.NodeUnpublishVolume(context, &csi.NodeUnpublishVolumeRequest{
						VolumeId:   volumeId,
						TargetPath: mountPath,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
					Expect(grpcStatus.Message()).To(Equal("Mount path is missing in request"))
				})
			})

			Context("when checking if mounted fails", func() {
				BeforeEach(func() {
					fakeOsHelper.IsMountedReturns(false, errors.New("failed to check if mounted"))
				})

				It("returns an error", func() {
					_, err := localNode.NodeUnpublishVolume(context, &csi.NodeUnpublishVolumeRequest{
						VolumeId:   volumeId,
						TargetPath: mountPath,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.Internal))
					Expect(grpcStatus.Message()).To(Equal("Error checking if volume is mounted"))
				})
			})

			Context("when unmounting the volume fails", func() {
				BeforeEach(func() {
					fakeOsHelper.IsMountedReturns(true, nil)
					fakeOsHelper.UnmountReturns(errors.New("failed to unmount"))
				})

				It("returns an error", func() {
					_, err := localNode.NodeUnpublishVolume(context, &csi.NodeUnpublishVolumeRequest{
						VolumeId:   volumeId,
						TargetPath: mountPath,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.Internal))
					Expect(grpcStatus.Message()).To(Equal("Error unmounting volume"))
				})
			})

			Context("when removing the mount path fails", func() {
				BeforeEach(func() {
					fakeOsHelper.IsMountedReturns(true, nil)
					fakeOs.RemoveReturns(errors.New("failed to remove"))
				})

				It("returns an error", func() {
					_, err := localNode.NodeUnpublishVolume(context, &csi.NodeUnpublishVolumeRequest{
						VolumeId:   volumeId,
						TargetPath: mountPath,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.Internal))
					Expect(grpcStatus.Message()).To(Equal("Error removing volume mount directory"))
				})
			})
		})
	})

	Describe("GetNodeID", func() {
		Context("when GetNodeID is called with a GetNodeIDRequest", func() {
			It("should return a GetNodeIDResponse that has a result with a node ID", func() {
				expectedResponse, err := localNode.NodeGetId(context, &csi.NodeGetIdRequest{})
				Expect(err).NotTo(HaveOccurred())
				Expect(expectedResponse).NotTo(BeNil())
				Expect(expectedResponse.GetNodeId()).To(Equal("some-node-id"))
			})
		})
	})

	Describe("NodeProbe", func() {
		Context("when NodeProbe is called with a NodeProbeRequest", func() {
			It("should return a NodeProbeResponse", func() {
				expectedResponse, err := localNode.Probe(context, &csi.ProbeRequest{})
				Expect(err).NotTo(HaveOccurred())
				Expect(expectedResponse).NotTo(BeNil())
			})
		})
	})

	Describe("NodeGetCapabilities", func() {
		Context("when NodeGetCapabilities is called with a NodeGetCapabilitiesRequest", func() {
			It("should return an empty NodeGetCapabilitiesResponse", func() {
				expectedResponse, err := localNode.NodeGetCapabilities(context, &csi.NodeGetCapabilitiesRequest{})
				Expect(err).NotTo(HaveOccurred())
				Expect(expectedResponse).NotTo(BeNil())
				capabilities := expectedResponse.GetCapabilities()
				Expect(capabilities).To(HaveLen(0))
			})
		})
	})

	Describe("NodeGetInfo", func() {
		Context("when NodeGetinfo is called with a NodeGetInfoRequest", func() {
			It("should return an empty NodeGetCapabilitiesResponse", func() {
				expectedResponse, err := localNode.NodeGetInfo(context, &csi.NodeGetInfoRequest{})
				Expect(err).NotTo(HaveOccurred())
				Expect(expectedResponse).NotTo(BeNil())
				Expect(*expectedResponse).To(Equal(csi.NodeGetInfoResponse{NodeId: "some-node-id"}))
			})
		})
	})

	Describe("GetPluginInfo", func() {
		Context("when provided with a GetPluginInfoRequest", func() {
			It("returns the plugin info", func() {
				expectedResponse, err := localNode.GetPluginInfo(context, &csi.GetPluginInfoRequest{})
				Expect(err).NotTo(HaveOccurred())
				Expect(expectedResponse).NotTo(BeNil())
				Expect(expectedResponse.GetName()).To(Equal(node.NODE_PLUGIN_ID))
				Expect(expectedResponse.GetVendorVersion()).To(Equal("0.1.0"))
			})
		})
	})
})

type DummyContext struct{}

func (*DummyContext) Deadline() (deadline time.Time, ok bool) { return time.Time{}, false }

func (*DummyContext) Done() <-chan struct{} { return nil }

func (*DummyContext) Err() error { return nil }

func (*DummyContext) Value(key interface{}) interface{} { return nil }

type FakeFileInfo struct {
	FileMode os.FileMode
}

func (FakeFileInfo) Name() string                { return "" }
func (FakeFileInfo) Size() int64                 { return 0 }
func (fs *FakeFileInfo) Mode() os.FileMode       { return fs.FileMode }
func (fs *FakeFileInfo) StubMode(fm os.FileMode) { fs.FileMode = fm }
func (FakeFileInfo) ModTime() time.Time          { return time.Time{} }
func (fs *FakeFileInfo) IsDir() bool             { return fs.FileMode == os.ModeDir }
func (FakeFileInfo) Sys() interface{}            { return nil }

func newFakeFileInfo() *FakeFileInfo {
	return &FakeFileInfo{}
}
