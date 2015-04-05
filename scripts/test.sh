#!/bin/bash -e

go install github.com/onsi/ginkgo/ginkgo
GOPATH=$PWD/Godeps/_workspace:$GOPATH ginkgo -r
