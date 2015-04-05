#!/bin/bash -e

go get -u -v github.com/onsi/ginkgo/ginkgo
GOPATH=$PWD/Godeps/_workspace:$GOPATH ginkgo -r
