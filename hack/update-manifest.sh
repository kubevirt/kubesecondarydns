#!/bin/bash -e

sed -i 's@registry:5000/kubevirt/kubesecondarydns:latest@'"$IMAGE"'@' manifests/secondarydns.yaml
