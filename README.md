# A node service for CSI-compliant local disk

This is [Cloud Foundry](https://github.com/cloudfoundry)'s implementation of the [Container Storage Interface](https://github.com/container-storage-interface/spec/blob/master/spec.md)'s Node Plugin for local volumes. The [CSI Local Volume Release](https://github.com/jeffpak/csi-local-volume-release) for Cloud Foundry submodules this repository. Functionally, this repository enables the CSI Local Volume Release access to node capabilities and serves to make the release compliant with node RPCs.

This repository is to be operated solely for testing purposes. It should be used as an example for all other Cloud Foundry controller plugins adhering to the Container Storage Interface.  

# Developer Notes

THIS REPOSITORY IS A WORK IN PROGRESS.

| RPC | Function | Expected Response | 
|---|---|---|
| NodePublishVolume | Mounts the share on the specified target path | Empty Result Response | 
| NodeUnpublishVolume | Unmounts the share from the specified target path | Empty Result Response |
| GetNodeID | No Op | Empty Result Response |
| ProbeNode | No Op | Empty Result Response |
| NodeGetCapabilities | No Op | Empty Result Response |

## Running Tests

1. Install [go](https://golang.org/doc/install).
1. Install [ginkgo](https://onsi.github.io/ginkgo/).
1. Clone this repository
1. ```ginkgo -r``` inside local-controller-plugin.
