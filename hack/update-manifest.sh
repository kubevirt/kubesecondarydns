#!/bin/bash -e

sed -i 's@registry:5000/alonakaplan/kubesecondarydns:latest@'"$IMAGE"'@' manifests/secondarydns.yaml
