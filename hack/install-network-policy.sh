#!/bin/bash -ex
#
# Copyright 2025 Red Hat, Inc.
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
#

# This script simulating network restriction by installing deny-all NetworkPolicy in the project namespace.
# Since allowing access for cluster API and DNS is fundamental requirement for the project, an appropriate NetworkPolicies are installed.

readonly ns="$(./cluster/kubectl.sh get pod -l k8s-app=secondary-dns -A -o=custom-columns=NS:.metadata.namespace --no-headers | head -1)"
[[ -z "${ns}" ]] && echo "FATAL: kube-secondary-dns pods not found. Make sure kube-secondary-dns is installed" && exit 1

cat <<EOF | ./cluster/kubectl.sh -n "${ns}" apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress: []
  egress: []
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-egress-to-cluster-dns
spec:
  podSelector:
    matchLabels:
      k8s-app: secondary-dns
  policyTypes:
  - Egress
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
      podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-egress-to-cluster-api
spec:
  podSelector:
    matchLabels:
      k8s-app: secondary-dns
  policyTypes:
  - Egress
  egress:
  - ports:
    - protocol: TCP
      port: 6443
EOF
