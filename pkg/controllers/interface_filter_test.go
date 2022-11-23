package controllers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "kubevirt.io/api/core/v1"
)

var _ = Describe("FilterMultusNonDefaultInterfaces", func() {
	const defaultName = "default"
	const nonDefaultName = "non default"
	const nonDefaultName2 = "non default2"

	It("when interfaces and networks lists are nil", func() {
		Expect(FilterMultusNonDefaultInterfaces(nil, nil)).To(BeEmpty())
	})
	It("when interfaces and networks lists are empty", func() {
		Expect(FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{}, []v1.Network{})).To(BeEmpty())
	})
	It("when there is only default network", func() {
		Expect(FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{createVmInterface(defaultName)}, []v1.Network{createDefaultNetwork(defaultName)})).To(BeEmpty())
	})
	It("when there is a non default interface", func() {
		nonDefaultInterface := createVmInterface(nonDefaultName)
		result := FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{nonDefaultInterface, createVmInterface(defaultName)}, []v1.Network{createDefaultNetwork(defaultName), createMultusNonDefaultNetwork(nonDefaultName)})
		Expect(result).To(ConsistOf(nonDefaultInterface))
	})
	It("when there is default network and multiple non default interface", func() {
		nonDefaultInterface := createVmInterface(nonDefaultName)
		nonDefaultInterface2 := createVmInterface(nonDefaultName2)
		result := FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{createVmInterface(defaultName), nonDefaultInterface, nonDefaultInterface2}, []v1.Network{createDefaultNetwork(defaultName)})
		Expect(result).To(ConsistOf(nonDefaultInterface, nonDefaultInterface2))
	})
	It("when there is a multus default interface", func() {
		nonDefaultInterface := createVmInterface(nonDefaultName)
		nonDefaultInterface2 := createVmInterface(nonDefaultName2)
		multusDefaultNetwork := v1.Network{
			NetworkSource: v1.NetworkSource{
				Multus: &v1.MultusNetwork{Default: true}},
			Name: defaultName,
		}

		result := FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{nonDefaultInterface, nonDefaultInterface2}, []v1.Network{multusDefaultNetwork, createMultusNonDefaultNetwork(nonDefaultName), createMultusNonDefaultNetwork(nonDefaultName2)})
		Expect(result).To(ConsistOf(nonDefaultInterface, nonDefaultInterface2))
	})
	It("when there is no default network", func() {
		nonDefaultInterface := createVmInterface(nonDefaultName)
		nonDefaultInterface2 := createVmInterface(nonDefaultName2)
		result := FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{nonDefaultInterface, nonDefaultInterface2}, []v1.Network{createMultusNonDefaultNetwork(nonDefaultName), createMultusNonDefaultNetwork(nonDefaultName2)})
		Expect(result).To(ConsistOf(nonDefaultInterface, nonDefaultInterface2))
	})
})

func createDefaultNetwork(name string) v1.Network {
	return v1.Network{
		NetworkSource: v1.NetworkSource{
			Pod: &v1.PodNetwork{}},
		Name: name,
	}
}

func createMultusNonDefaultNetwork(name string) v1.Network {
	return v1.Network{
		NetworkSource: v1.NetworkSource{
			Multus: &v1.MultusNetwork{Default: false}},
		Name: name,
	}
}

func createVmInterface(name string) v1.VirtualMachineInstanceNetworkInterface {
	return v1.VirtualMachineInstanceNetworkInterface{Name: name}
}
