package main

import (
	"net"
	"os"

	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/osshim"

	"code.cloudfoundry.org/lager"
	csi "github.com/jeffpak/csi"
	"github.com/jeffpak/local-node-plugin/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50052"
)

func main() {
	logger := lager.NewLogger("local-node-plugin")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	logger.RegisterSink(sink)

	lis, err := net.Listen("tcp", port)
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
