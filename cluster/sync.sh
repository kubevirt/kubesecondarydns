#!/bin/bash
#
# Copyright 2018-2022 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -ex

source ./cluster/cluster.sh
cluster::install

registry_port=$(./cluster/cli.sh ports registry | tr -d '\r')
export REGISTRY=localhost:$registry_port

make build
make push
make cluster-clean
make deploy

# Ensure the project network-polices are valid by simulating network restrictions using network-policy
./hack/install-network-policy.sh

pods_ready_wait() {
  if [[ "$KUBEVIRT_PROVIDER" != external ]]; then
    echo "Waiting for non secondary-dns containers to be ready ..."
    ./cluster/kubectl.sh wait pod --all -n kube-system --for=condition=Ready --timeout=5m
  fi

  wait_failed=''
  max_retries=12
  retry_counter=0
  echo "Waiting for secondary dns deployment to be ready ..."
  while [[ "$(./cluster/kubectl.sh wait deployment -n secondary secondary-dns --for=condition=Available --timeout=1m)" = $wait_failed ]] && [[ $retry_counter -lt $max_retries ]]; do
    sleep 5s
    retry_counter=$((retry_counter + 1))
  done

  if [ $retry_counter -eq $max_retries ]; then
    echo "Failed/timed-out waiting for secondary-dns resources"
    exit 1
  else
    echo "secondary-dns deployment is ready"
  fi
}

enable_psa_feature_gate() {
  ./cluster/kubectl.sh apply -f ./hack/psa/kubevirt.yaml
}

pods_ready_wait
make create-nodeport
enable_psa_feature_gate
