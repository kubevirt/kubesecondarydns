package filter

import (
	v1 "kubevirt.io/api/core/v1"
)

func FilterMultusNonDefaultInterfaces(ifaces []v1.VirtualMachineInstanceNetworkInterface, networks []v1.Network) []v1.VirtualMachineInstanceNetworkInterface {
	defaultNetwork := getDefaultNetwork(networks)
	if defaultNetwork == nil {
		return ifaces
	}
	var secondaryInterfaces []v1.VirtualMachineInstanceNetworkInterface
	for _, iface := range ifaces {
		if iface.Name != defaultNetwork.Name {
			secondaryInterfaces = append(secondaryInterfaces, iface)
		}
	}
	return secondaryInterfaces
}

func getDefaultNetwork(networks []v1.Network) *v1.Network {
	for i, network := range networks {
		if isDefaultNetwork(network) {
			return &networks[i]
		}
	}
	return nil
}

func isDefaultNetwork(net v1.Network) bool {
	return net.Pod != nil || (net.Multus != nil && net.Multus.Default)
}

func FilterNamedInterfaces(ifaces []v1.VirtualMachineInstanceNetworkInterface) []v1.VirtualMachineInstanceNetworkInterface {
	var secondaryInterfaces []v1.VirtualMachineInstanceNetworkInterface
	for _, iface := range ifaces {
		if iface.Name != "" {
			secondaryInterfaces = append(secondaryInterfaces, iface)
		}
	}
	return secondaryInterfaces
}
