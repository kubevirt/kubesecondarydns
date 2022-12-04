#!/bin/bash -e

source ./cluster/cluster.sh

if ! nslookup www.google.com > /dev/null 2>&1; then
  echo "nslookup error"
  exit 1
fi

KUBECONFIG=${KUBECONFIG:-$(cluster::kubeconfig)} $GO test -test.timeout=1h -test.v ./tests/... -ginkgo.v
