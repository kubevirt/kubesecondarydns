#!/bin/bash -e

if [ -z $KUBECONFIG ]; then
  KUBECTL=./cluster/kubectl.sh
else
  KUBECTL=kubectl
fi

NAMESPACE=$($KUBECTL get pod -lk8s-app=secondary-dns -A -o=custom-columns=NS:.metadata.namespace --no-headers)

if [ -z $NAMESPACE ]; then
  echo "ERROR: namespace not found"
  exit 1
fi

${KUBECTL} expose -n ${NAMESPACE} deployment/secondary-dns --name=dns-nodeport --type=NodePort --port=31111 --target-port=53 --protocol='UDP'
${KUBECTL} patch -n ${NAMESPACE} service/dns-nodeport --type='json' --patch='[{"op": "replace", "path": "/spec/ports/0/nodePort", "value":31111}]'
