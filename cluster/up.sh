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

set -ex pipefail

export DEPLOY_CNAO=${DEPLOY_CNAO:-true}
export DEPLOY_KUBEVIRT=${DEPLOY_KUBEVIRT:-true}
export KUBEVIRT_PSA=${KUBEVIRT_PSA:-true}

export KUBEVIRT_FLANNEL=true

source ./cluster/cluster.sh

# use kubevirt latest stable release
KUBEVIRT_VERSION=$(curl -s https://api.github.com/repos/kubevirt/kubevirt/releases | grep tag_name | grep -v -- - | sort -V | tail -1 | awk -F':' '{print $2}' | sed 's/,//' | xargs)
cluster::install

if [[ "$KUBEVIRT_PROVIDER" != external ]]; then
    if [[ "${DEPLOY_CNAO}" = "true" ]]; then
        export KUBEVIRT_WITH_CNAO=true
        export KUBVIRT_WITH_CNAO_SKIP_CONFIG=true
    fi

    $(cluster::path)/cluster-up/up.sh

    if [[ "${DEPLOY_CNAO}" = "true" ]]; then
        # Deploy CNAO CR
        ./cluster/kubectl.sh create -f ./hack/cna/cna-cr.yaml

        # wait for cluster operator
        ./cluster/kubectl.sh wait networkaddonsconfig cluster --for condition=Available --timeout=120s
    fi
fi

if [[ "${DEPLOY_KUBEVIRT}" = "true" ]]; then
    # deploy kubevirt
    ./cluster/kubectl.sh apply -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-operator.yaml

    # Ensure the KubeVirt CRD is created
    count=0
    until ./cluster/kubectl.sh get crd kubevirts.kubevirt.io; do
        ((count++)) && ((count == 30)) && echo "KubeVirt CRD not found" && exit 1
        echo "waiting for KubeVirt CRD"
        sleep 1
    done

    ./cluster/kubectl.sh apply -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml

    # Ensure the KubeVirt CR is created
    count=0
    until ./cluster/kubectl.sh -n kubevirt get kv kubevirt; do
        ((count++)) && ((count == 30)) && echo "KubeVirt CR not found" && exit 1
        echo "waiting for KubeVirt CR"
        sleep 1
    done

    ./cluster/kubectl.sh wait -n kubevirt kv kubevirt --for condition=Available --timeout 360s || (echo "KubeVirt not ready in time" && exit 1)
fi

echo "Done"
