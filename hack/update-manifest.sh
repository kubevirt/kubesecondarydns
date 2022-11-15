#!/bin/bash -e

sed -i 's@registry:5000/alonapaz/kubesecondarydns:latest@'"$IMAGE"'@' manifests/secondarydns.yaml
