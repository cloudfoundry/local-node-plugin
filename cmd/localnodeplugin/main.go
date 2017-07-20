package main

import (
	"net"

	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/osshim"

	csi "github.com/jeffpak/csi"
	"github.com/jeffpak/local-node-plugin/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"code.cloudfoundry.org/lager"
)

const (
	port = ":50051"
)

func main() {
	lis, err := net.Listen("tcp", port)
	logger := lager.NewLogger("local-node-plugin")
	if err != nil {
		logger.Fatal("failed to listen:", err)
	}
	s := grpc.NewServer()

	node := node.NewLocalNode(&osshim.OsShim{}, &filepathshim.FilepathShim{}, logger)
	csi.RegisterNodeServer(s, node)

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		logger.Fatal("failed to serve:", err)
	}
}
