#!/bin/bash

set -e
echo "installing ifrit"
go get -u "github.com/tedsuo/ifrit"
echo "installing lager"
go get -u "code.cloudfoundry.org/lager"
echo "installing goshims"
go get -u "code.cloudfoundry.org/goshims" >/dev/null 2>&1 || true
echo "installing ginkgo..."
go get -u "github.com/onsi/ginkgo/ginkgo"
echo "installing gomega..."
go get -u "github.com/onsi/gomega"
go get -u "github.com/onsi/gomega/types"
echo "installing grpc..."
go get -u "google.golang.org/grpc"
echo "installing csi spec..."
go get -u "github.com/paulcwarren/spec"

echo "done."
