package controllers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "kubevirt.io/api/core/v1"

	"github.com/kubevirt/kubesecondarydns/pkg/controllers"
)

var _ = Describe("FilterMultusNonDefaultInterfaces", func() {
	const defaultName = "default"
	const nonDefaultName = "non default"
	const nonDefaultName2 = "non default2"

	It("when interfaces and networks lists are nil", func() {
		Expect(controllers.FilterMultusNonDefaultInterfaces(nil, nil)).To(BeEmpty())
	})
	It("when interfaces and networks lists are empty", func() {
		Expect(controllers.FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{}, []v1.Network{})).To(BeEmpty())
	})
	It("when there is only default network", func() {
		Expect(controllers.FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{createVmInterface(defaultName)}, []v1.Network{createDefaultNetwork(defaultName)})).To(BeEmpty())
	})
	It("when there is a non default interface", func() {
		nonDefaultInterface := createVmInterface(nonDefaultName)
		result := controllers.FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{nonDefaultInterface, createVmInterface(defaultName)}, []v1.Network{createDefaultNetwork(defaultName), createMultusNonDefaultNetwork(nonDefaultName)})
		Expect(result).To(ConsistOf(nonDefaultInterface))
	})
	It("when there is default network and multiple non default interface", func() {
		nonDefaultInterface := createVmInterface(nonDefaultName)
		nonDefaultInterface2 := createVmInterface(nonDefaultName2)
		result := controllers.FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{createVmInterface(defaultName), nonDefaultInterface, nonDefaultInterface2}, []v1.Network{createDefaultNetwork(defaultName)})
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

		result := controllers.FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{nonDefaultInterface, nonDefaultInterface2}, []v1.Network{multusDefaultNetwork, createMultusNonDefaultNetwork(nonDefaultName), createMultusNonDefaultNetwork(nonDefaultName2)})
		Expect(result).To(ConsistOf(nonDefaultInterface, nonDefaultInterface2))
	})
	It("when there is no default network", func() {
		nonDefaultInterface := createVmInterface(nonDefaultName)
		nonDefaultInterface2 := createVmInterface(nonDefaultName2)
		result := controllers.FilterMultusNonDefaultInterfaces([]v1.VirtualMachineInstanceNetworkInterface{nonDefaultInterface, nonDefaultInterface2}, []v1.Network{createMultusNonDefaultNetwork(nonDefaultName), createMultusNonDefaultNetwork(nonDefaultName2)})
		Expect(result).To(ConsistOf(nonDefaultInterface, nonDefaultInterface2))
	})
})

var _ = Describe("FilterInterfacesWithNoName", func() {
	It("when interfaces list is nil", func() {
		Expect(controllers.FilterNamedInterfaces(nil)).To(BeEmpty())
	})
	It("when interface list is empty", func() {
		Expect(controllers.FilterNamedInterfaces([]v1.VirtualMachineInstanceNetworkInterface{})).To(BeEmpty())
	})
	It("when there is one interface without a name", func() {
		Expect(controllers.FilterNamedInterfaces([]v1.VirtualMachineInstanceNetworkInterface{createVmInterface("")})).To(BeEmpty())
	})
	It("when there is one inteface with a name", func() {
		iface := createVmInterface("nic1")
		result := controllers.FilterNamedInterfaces([]v1.VirtualMachineInstanceNetworkInterface{iface})
		Expect(result).To(ConsistOf(iface))
	})
	It("when there are multiple interface, some with name some without", func() {
		nic1 := createVmInterface("nic1")
		nic2 := createVmInterface("")
		nic3 := createVmInterface("nic3")
		nic4 := createVmInterface("")
		result := controllers.FilterNamedInterfaces([]v1.VirtualMachineInstanceNetworkInterface{nic1, nic2, nic3, nic4})
		Expect(result).To(ConsistOf(nic1, nic3))
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
