package main

import (
	"os"

	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"

	"flag"

	"code.cloudfoundry.org/csiplugin"
	"code.cloudfoundry.org/lager/lagerflags"
	. "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/jeffpak/local-node-plugin/node"
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

func main() {
	parseCommandLine()

	logger := lager.NewLogger("local-node-plugin")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	logger.RegisterSink(sink)

	listenAddress := *atAddress

	err := csiplugin.WriteSpec(logger, *pluginsPath, csiplugin.CsiPluginSpec{Name: node.NODE_PLUGIN_ID, Address: listenAddress})
	if err != nil {
		logger.Fatal("exited-with-failure:", err)
	}

	node := node.NewLocalNode(&osshim.OsShim{}, &filepathshim.FilepathShim{}, logger, *volumesRoot)
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
