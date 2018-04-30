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
	. "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/jeffpak/local-node-plugin/node"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Node Client", func() {
	var (
		nc           *node.LocalNode
		testLogger   lager.Logger
		context      context.Context
		fakeOs       *os_fake.FakeOs
		fakeFilepath *filepath_fake.FakeFilepath
		vc           *VolumeCapability
		volumeId     string
		err          error
		fileInfo     *FakeFileInfo
		publishResp  *NodePublishVolumeResponse
		volumesRoot  string
	)

	BeforeEach(func() {
		volumesRoot = "/tmp/_volumes"
		testLogger = lagertest.NewTestLogger("localdriver-local")
		context = &DummyContext{}

		fakeOs = &os_fake.FakeOs{}
		fakeFilepath = &filepath_fake.FakeFilepath{}
		fakeFilepath.AbsReturns("/path/to/mount", nil)

		nc = node.NewLocalNode(fakeOs, fakeFilepath, testLogger, volumesRoot)
		volumeId = "test-volume-id"
		vc = &VolumeCapability{AccessType: &VolumeCapability_Mount{Mount: &VolumeCapability_MountVolume{}}, AccessMode: &VolumeCapability_AccessMode{}}

		fileInfo = newFakeFileInfo()
		fakeOs.LstatReturns(fileInfo, nil)
		fileInfo.StubMode(os.ModeSymlink)
	})

	Describe("NodePublishVolume", func() {

		Context("when mount path exist", func() {
			var (
				mount_path = "/path/to/mount/_mounts/test-volume-id"
			)

			BeforeEach(func() {
				fakeOs.StatReturns(nil, nil)
			})

			JustBeforeEach(func() {
				publishResp, err = nodePublish(context, nc, volumeId, vc, mount_path)
			})

			Context("when the mount path is a directory", func() {
				BeforeEach(func() {
					fileInfo := newFakeFileInfo()
					fakeOs.LstatReturns(fileInfo, nil)
					fileInfo.StubMode(os.ModeDir)
				})

				It("deletes a destination directory, creates volumesRoot directory and Send publish request to CSI node server", func() {
					Expect(err).To(BeNil())
					Expect(publishResp).NotTo(BeNil())

					Expect(fakeOs.RemoveCallCount()).To(Equal(1))
					path := fakeOs.RemoveArgsForCall(0)
					Expect(path).To(Equal(mount_path))

					Expect(fakeOs.MkdirAllCallCount()).To(Equal(1))
					path, _ = fakeOs.MkdirAllArgsForCall(0)
					Expect(path).To(Equal(filepath.Join(volumesRoot, volumeId)))

					Expect(fakeOs.SymlinkCallCount()).To(Equal(1))
					from, to := fakeOs.SymlinkArgsForCall(0)
					Expect(from).To(Equal(filepath.Join(volumesRoot, "test-volume-id")))
					Expect(to).To(Equal(mount_path))
				})
			})

			//TODO
			Context("when the mount path is a symbolic link", func() {
			})
		})
	})

	Context("when the volume has been created", func() {
		var (
			mount_path        = "/path/to/mount/_mounts/test-volume-id"
			mount_path_parent = filepath.Dir(mount_path)
		)

		BeforeEach(func() {
			fakeOs.StatReturns(nil, os.ErrNotExist)
		})

		JustBeforeEach(func() {
			publishResp, err = nodePublish(context, nc, volumeId, vc, mount_path)
		})

		Context("when the volume exists", func() {
			AfterEach(func() {
				fileInfo := newFakeFileInfo()
				fakeOs.LstatReturns(fileInfo, nil)
				fileInfo.StubMode(os.ModeSymlink)
				unmountResponse, err := nodeUnpublish(context, nc, volumeId, mount_path)
				Expect(err).To(BeNil())
				Expect(unmountResponse).NotTo(BeNil())
			})

			Context("when volume is a symlink", func() {
				BeforeEach(func() {
					fileInfo := newFakeFileInfo()
					fakeOs.LstatReturns(fileInfo, nil)
					fileInfo.StubMode(os.ModeSymlink)
				})

				It("should mount the volume on the local filesystem", func() {
					// Expect(fakeFilepath.AbsCallCount()).To(Equal(1))
					Expect(err).To(BeNil())
					Expect(publishResp).NotTo(BeNil())

					Expect(fakeOs.MkdirAllCallCount()).To(Equal(1))
					path, _ := fakeOs.MkdirAllArgsForCall(0)
					Expect(path).To(Equal(filepath.Join(volumesRoot, volumeId)))
					Expect(fakeOs.SymlinkCallCount()).To(Equal(1))
					from, to := fakeOs.SymlinkArgsForCall(0)
					Expect(from).To(Equal(filepath.Join(volumesRoot, "test-volume-id")))
					Expect(to).To(Equal(mount_path))
				})
			})

			Context("when the volume's base directory doesn't exist", func() {
				BeforeEach(func() {
					fileInfo = newFakeFileInfo()
					err = os.ErrNotExist
					fakeOs.StatReturns(fileInfo, err)
					fakeOs.IsNotExistReturns(true)
				})

				It("Create volumesRoot directory and Send publish request to CSI node server", func() {
					Expect(err).To(BeNil())
					Expect(publishResp).NotTo(BeNil())

					Expect(fakeOs.MkdirAllCallCount()).To(Equal(2))
					path, _ := fakeOs.MkdirAllArgsForCall(0)
					Expect(path).To(Equal(filepath.Join(volumesRoot, volumeId)))

					path, _ = fakeOs.MkdirAllArgsForCall(1)
					Expect(path).To(ContainSubstring(mount_path_parent))

					Expect(fakeOs.SymlinkCallCount()).To(Equal(1))
					from, to := fakeOs.SymlinkArgsForCall(0)
					Expect(from).To(Equal(filepath.Join(volumesRoot, "test-volume-id")))
					Expect(to).To(Equal(mount_path))
				})
			})

			Context("when the volume is node published a second time", func() {
				JustBeforeEach(func() {
					fakeOs.StatReturns(nil, nil)
					publishResp, err = nodePublish(context, nc, volumeId, vc, mount_path)
				})

				It("should succeed", func() {
					Expect(err).To(BeNil())
					Expect(publishResp).NotTo(BeNil())
					Expect(fakeOs.SymlinkCallCount()).To(Equal(1))
				})
			})
		})

		Context("when the volume id is missing", func() {
			BeforeEach(func() {
				fakeOs.StatReturns(nil, os.ErrNotExist)
				fakeOs.LstatReturns(fileInfo, nil)
				fileInfo.StubMode(os.ModeDir)
				volumeId = ""
			})
			AfterEach(func() {
				fakeOs.StatReturns(nil, nil)
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				grpcStatus, _ := status.FromError(err)
				Expect(grpcStatus).NotTo(BeNil())
				Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
				Expect(grpcStatus.Message()).To(Equal("Volume ID is missing in request"))
			})
		})

		Context("when the volume capability is missing", func() {
			BeforeEach(func() {
				vc = nil
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				grpcStatus, _ := status.FromError(err)
				Expect(grpcStatus).NotTo(BeNil())
				Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
				Expect(grpcStatus.Message()).To(Equal("Volume capability is missing in request"))
			})
		})

		Context("When the volume capability is not mount capability", func() {
			BeforeEach(func() {
				vc = &VolumeCapability{}
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				grpcStatus, _ := status.FromError(err)
				Expect(grpcStatus).NotTo(BeNil())
				Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
				Expect(grpcStatus.Message()).To(Equal("Volume mount capability is not specified"))
			})
		})
	})

	Describe("NodeUnpublishVolume", func() {
		var (
			request          *NodeUnpublishVolumeRequest
			expectedResponse *NodeUnpublishVolumeResponse
		)
		Context("when NodeUnpublishVolume is called with a NodeUnpublishVolume", func() {
			BeforeEach(func() {
				request = &NodeUnpublishVolumeRequest{
					volumeId,
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

		Context("when a volume has been mounted", func() {
			var (
				mount_path = "/path/to/mount/_mounts/test-volume-id"
			)

			JustBeforeEach(func() {
				publish_resp, publish_err := nodePublish(context, nc, volumeId, vc, mount_path)
				Expect(publish_err).ToNot(HaveOccurred())
				Expect(publish_resp).ToNot(BeNil())
			})

			It("Unmount the volume", func() {
				unmountResponse, err := nodeUnpublish(context, nc, volumeId, mount_path)
				Expect(err).To(BeNil())
				Expect(unmountResponse).NotTo(BeNil())
				des := fakeOs.RemoveArgsForCall(0)
				Expect(des).To(Equal(mount_path))
			})

			Context("when the mountpath is not found on the filesystem", func() {
				It("returns a success", func() {
					fileInfo = newFakeFileInfo()
					err = os.ErrNotExist

					fakeOs.LstatReturns(fileInfo, err)
					fakeOs.IsNotExistReturns(true)
					path := "/not-found"
					resp, err := nc.NodeUnpublishVolume(context, &NodeUnpublishVolumeRequest{
						VolumeId:   "abcd",
						TargetPath: path,
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(resp).ToNot(BeNil())
				})
			})

			Context("when the mountpath is not a symbolic link", func() {
				It("returns an error", func() {
					fileInfo := newFakeFileInfo()
					err = os.ErrNotExist
					fakeOs.LstatReturns(fileInfo, err)
					fileInfo.StubMode(os.ModeTemporary)

					path := "/not-symbolic-link"
					resp, err := nc.NodeUnpublishVolume(context, &NodeUnpublishVolumeRequest{
						VolumeId:   "abcd",
						TargetPath: path,
					})

					errorMsg := "Mount point '/not-symbolic-link' is not a symbolic link"
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
					Expect(grpcStatus.Message()).To(Equal(errorMsg))
					Expect(resp).To(BeNil())
				})
			})

			Context("when the volume id is missing", func() {
				It("returns an error", func() {
					var path string = "/test-path"
					resp, err := nc.NodeUnpublishVolume(context, &NodeUnpublishVolumeRequest{
						TargetPath: path,
					})
					errorMsg := "Volume ID is missing in request"
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
					Expect(grpcStatus.Message()).To(Equal(errorMsg))
					Expect(resp).To(BeNil())
				})
			})

			Context("when the mount path is missing", func() {
				It("returns an error", func() {
					resp, err := nc.NodeUnpublishVolume(context, &NodeUnpublishVolumeRequest{
						VolumeId: "abcd",
					})
					errorMsg := "Mount path is missing in the request"
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.InvalidArgument))
					Expect(grpcStatus.Message()).To(Equal(errorMsg))
					Expect(resp).To(BeNil())
				})
			})

			Context("when the removal fails while unmounting", func() {
				var errorMsg string
				BeforeEach(func() {
					errorMsg = "Error ummounting volume abcd"
					fakeOs.RemoveReturns(errors.New(errorMsg))
				})
				It("returns an error", func() {
					var path string = "/test-path"
					resp, err := nc.NodeUnpublishVolume(context, &NodeUnpublishVolumeRequest{
						VolumeId:   "abcd",
						TargetPath: path,
					})
					Expect(err).To(HaveOccurred())
					grpcStatus, _ := status.FromError(err)
					Expect(grpcStatus).NotTo(BeNil())
					Expect(grpcStatus.Code()).To(Equal(codes.Internal))
					Expect(grpcStatus.Message()).To(Equal(errorMsg))
					Expect(resp).To(BeNil())
				})
			})
		})
	})

	Describe("GetNodeID", func() {
		var (
			request          *NodeGetIdRequest
			expectedResponse *NodeGetIdResponse
		)
		Context("when GetNodeID is called with a GetNodeIDRequest", func() {
			BeforeEach(func() {
				request = &NodeGetIdRequest{}
			})
			JustBeforeEach(func() {
				expectedResponse, err = nc.NodeGetId(context, request)
			})
			It("should return a GetNodeIDResponse that has a result with no node ID", func() {
				Expect(expectedResponse).NotTo(BeNil())
				Expect(expectedResponse.GetNodeId()).To(BeEmpty())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("NodeProbe", func() {
		var (
			request          *ProbeRequest
			expectedResponse *ProbeResponse
		)
		Context("when NodeProbe is called with a NodeProbeRequest", func() {
			BeforeEach(func() {
				request = &ProbeRequest{}
			})
			JustBeforeEach(func() {
				expectedResponse, err = nc.Probe(context, request)
			})
			It("should return a NodeProbeResponse", func() {
				Expect(expectedResponse).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("NodeGetCapabilities", func() {
		var (
			request          *NodeGetCapabilitiesRequest
			expectedResponse *NodeGetCapabilitiesResponse
		)
		Context("when NodeGetCapabilities is called with a NodeGetCapabilitiesRequest", func() {
			BeforeEach(func() {
				request = &NodeGetCapabilitiesRequest{}
			})
			JustBeforeEach(func() {
				expectedResponse, err = nc.NodeGetCapabilities(context, request)
			})

			It("should return an empty NodeGetCapabilitiesResponse", func() {
				Expect(expectedResponse).NotTo(BeNil())
				capabilities := expectedResponse.GetCapabilities()
				Expect(capabilities).To(HaveLen(0))
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("GetPluginInfo", func() {
		var (
			request          *GetPluginInfoRequest
			expectedResponse *GetPluginInfoResponse
		)
		Context("when provided with a GetPluginInfoRequest", func() {
			BeforeEach(func() {
				request = &GetPluginInfoRequest{}
			})

			JustBeforeEach(func() {
				expectedResponse, err = nc.GetPluginInfo(context, request)
			})

			It("returns the plugin info", func() {
				Expect(expectedResponse).NotTo(BeNil())
				Expect(err).ToNot(HaveOccurred())
				Expect(expectedResponse.GetName()).To(Equal(node.NODE_PLUGIN_ID))
				Expect(expectedResponse.GetVendorVersion()).To(Equal("0.1.0"))
			})
		})
	})
})

func nodePublish(ctx context.Context, ns NodeServer, volumeId string, volCapability *VolumeCapability, targetPath string) (*NodePublishVolumeResponse, error) {
	mountResponse, err := ns.NodePublishVolume(ctx, &NodePublishVolumeRequest{
		VolumeId:         volumeId,
		TargetPath:       targetPath,
		VolumeCapability: volCapability,
		Readonly:         false,
	})
	return mountResponse, err
}

func nodeUnpublish(ctx context.Context, ns NodeServer, volumeId string, path string) (*NodeUnpublishVolumeResponse, error) {
	unmountResponse, err := ns.NodeUnpublishVolume(ctx, &NodeUnpublishVolumeRequest{
		VolumeId:   volumeId,
		TargetPath: path,
	})
	return unmountResponse, err
}

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
