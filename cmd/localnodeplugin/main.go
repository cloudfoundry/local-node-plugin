package main

import (
	"os"

	cf_lager "code.cloudfoundry.org/cflager"
	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"

	"flag"

	"github.com/Kaixiang/csiplugin"
	"github.com/jeffpak/local-node-plugin/node"
	. "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grpc_server"
	"github.com/tedsuo/ifrit/sigmon"
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
	server := grpc_server.NewGRPCServer(listenAddress, nil, node, RegisterNodeServer)

	monitor := ifrit.Invoke(sigmon.New(server))
	logger.Info("started")

	err = <-monitor.Wait()

	if err != nil {
		logger.Fatal("exited-with-failure:", err)
	}
}

func parseCommandLine() {
	cf_lager.AddFlags(flag.CommandLine)
	flag.Parse()
}
