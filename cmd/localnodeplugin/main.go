package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"

	. "github.com/container-storage-interface/spec"
	"github.com/jeffpak/local-node-plugin/node"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grpc_server"
)

const (
	port = 50052
)

func main() {
	logger := lager.NewLogger("local-node-plugin")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	logger.RegisterSink(sink)

	listenAddress := fmt.Sprintf("0.0.0.0:%d", port)

	node := node.NewLocalNode(&osshim.OsShim{}, &filepathshim.FilepathShim{}, logger)
	server := grpc_server.NewGRPCServer(listenAddress, nil, node, RegisterNodeServer)

	monitor := ifrit.Invoke(server)
	logger.Info("Node started")

	err := <-monitor.Wait()

	if err != nil {
		logger.Fatal("exited-with-failure:", err)
	}
}
