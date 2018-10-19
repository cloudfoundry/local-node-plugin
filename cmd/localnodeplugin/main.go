package main

import (
	"flag"

	"code.cloudfoundry.org/csiplugin"
	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/local-node-plugin/node"
	"code.cloudfoundry.org/local-node-plugin/oshelper"
	. "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grpc_server"
	"github.com/tedsuo/ifrit/sigmon"
	"google.golang.org/grpc"
)

var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:9760",
	"host:port to serve on",
)

var pluginsPath = flag.String(
	"pluginsPath",
	"",
	"Path to directory where plugin specs are installed",
)

var volumesRoot = flag.String(
	"volumesRoot",
	"/tmp/_volumes",
	"Path to directory where plugin mount point start with",
)

var nodeId = flag.String(
	"nodeId",
	"",
	"ID of the current node",
)

func main() {
	parseCommandLine()

	logger := lagerflags.NewFromConfig("local-node-plugin", lagerflags.ConfigFromFlags())

	listenAddress := *atAddress

	err := csiplugin.WriteSpec(logger, *pluginsPath, csiplugin.CsiPluginSpec{Name: node.NODE_PLUGIN_ID, Address: listenAddress})
	if err != nil {
		logger.Fatal("exited-with-failure:", err)
	}

	os := &osshim.OsShim{}
	node := node.NewLocalNode(os, oshelper.NewOsHelper(os), &filepathshim.FilepathShim{}, logger, *volumesRoot, *nodeId)
	server := grpc_server.NewGRPCServer(listenAddress, nil, node, RegisterServices)

	monitor := ifrit.Invoke(sigmon.New(server))
	logger.Info("started")

	err = <-monitor.Wait()

	if err != nil {
		logger.Fatal("exited-with-failure:", err)
	}
}

func parseCommandLine() {
	lagerflags.AddFlags(flag.CommandLine)
	flag.Parse()
}

func RegisterServices(s *grpc.Server, srv interface{}) {
	RegisterNodeServer(s, srv.(NodeServer))
	RegisterIdentityServer(s, srv.(IdentityServer))
}
