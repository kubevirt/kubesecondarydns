package zone_file_cache

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"fmt"
	"sort"
	"strings"

	k8stypes "k8s.io/apimachinery/pkg/types"
	v1 "kubevirt.io/api/core/v1"
)

var _ = Describe("cached zone file content maintenance", func() {

	const (
		domain       = "domain.com"
		nameServerIP = "185.251.75.10"
	)

	var zoneFileCache *ZoneFileCache

	Describe("cached zone file initialization", func() {
		const (
			headerDefault   = "$ORIGIN vm. \n$TTL 3600 \n@ IN SOA ns.vm. email.vm. (0 3600 3600 1209600 3600)\n"
			headerCustomFmt = "$ORIGIN vm.%s. \n$TTL 3600 \n@ IN SOA ns.vm.%s. email.vm.%s. (0 3600 3600 1209600 3600)\n@ IN NS ns.vm.%s.\nns IN A %s\n"
			headerSoaSerial = "$ORIGIN vm. \n$TTL 3600 \n@ IN SOA ns.vm. email.vm. (12345 3600 3600 1209600 3600)\n"
		)

		var (
			headerCustom = fmt.Sprintf(headerCustomFmt, domain, domain, domain, domain, nameServerIP)
		)

		DescribeTable("generate zone file header", func(nameServerIP, domain, expectedHeader string) {
			zoneFileCache = NewZoneFileCache(nameServerIP, domain, nil)
			Expect(zoneFileCache.header).To(Equal(expectedHeader))
		},
			Entry("header should contain default values", "", "vm", headerDefault),
			Entry("header should contain custom values", nameServerIP, "vm."+domain, headerCustom),
		)

		It("should init header with existing SOA serial", func() {
			soaSerial := 12345
			zoneFileCache = NewZoneFileCache("", "vm", &soaSerial)
			Expect(zoneFileCache.header).To(Equal(headerSoaSerial))
		})
	})

	Describe("cached zone file records update", func() {
		const (
			vmi1Name   = "vmi1"
			vmi2Name   = "vmi2"
			namespace1 = "ns1"
			namespace2 = "ns2"
			nic1Name   = "nic1"
			nic1IP     = "1.2.3.4"
			nic2Name   = "nic2"
			nic2IP     = "5.6.7.8"
			nic3Name   = "nic3"
			nic3IP     = "9.10.11.12"
			nic4Name   = "nic4"
			nic4IP     = "13.14.15.16"
			IPv6       = "fe80::74c8:f2ff:fe5f:ff2b"

			aRecordFmt = "%s.%s.%s IN A %s\n"

			updated    = true
			notUpdated = false
		)

		var (
			aRecord_nic1_vm1_ns1 = fmt.Sprintf(aRecordFmt, nic1Name, vmi1Name, namespace1, nic1IP)
			aRecord_nic2_vm1_ns1 = fmt.Sprintf(aRecordFmt, nic2Name, vmi1Name, namespace1, nic2IP)
			aRecord_nic4_vm1_ns1 = fmt.Sprintf(aRecordFmt, nic4Name, vmi1Name, namespace1, nic4IP)

			aRecord_nic1_vm2_ns1 = fmt.Sprintf(aRecordFmt, nic1Name, vmi2Name, namespace1, nic1IP)
			aRecord_nic2_vm2_ns1 = fmt.Sprintf(aRecordFmt, nic2Name, vmi2Name, namespace1, nic2IP)
			aRecord_nic3_vm2_ns1 = fmt.Sprintf(aRecordFmt, nic3Name, vmi2Name, namespace1, nic3IP)
			aRecord_nic4_vm2_ns1 = fmt.Sprintf(aRecordFmt, nic4Name, vmi2Name, namespace1, nic4IP)

			aRecord_nic3_vm1_ns2 = fmt.Sprintf(aRecordFmt, nic3Name, vmi1Name, namespace2, nic3IP)
			aRecord_nic4_vm1_ns2 = fmt.Sprintf(aRecordFmt, nic4Name, vmi1Name, namespace2, nic4IP)
		)

		validateUpdateFunc := func(vmiName, vmiNamespace string, newInterfaces []v1.VirtualMachineInstanceNetworkInterface,
			expectedIsUpdated bool, expectedRecords string, expectedSoaSerial int) {
			isUpdated := zoneFileCache.UpdateVMIRecords(k8stypes.NamespacedName{Namespace: vmiNamespace, Name: vmiName}, newInterfaces)
			Expect(isUpdated).To(Equal(expectedIsUpdated))
			Expect(sortRecords(zoneFileCache.aRecords)).To(Equal(sortRecords(expectedRecords)))
			Expect(zoneFileCache.soaSerial).To(Equal(expectedSoaSerial))
		}

		When("interfaces records list is empty", func() {
			BeforeEach(func() {
				zoneFileCache = NewZoneFileCache(nameServerIP, domain, nil)
			})

			DescribeTable("Updating interfaces records", validateUpdateFunc,
				Entry("when new vmi with interfaces list is added",
					vmi1Name,
					namespace1,
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic1IP}, Name: nic1Name}},
					updated,
					aRecord_nic1_vm1_ns1,
					1,
				),
				Entry("when non existing vmi is deleted",
					vmi1Name,
					namespace1,
					nil,
					notUpdated,
					"",
					0,
				),
			)
		})

		When("SOA serial already exist", func() {
			It("should init SOA serial with the existing value", func() {
				soaSerial := 5
				zoneFileCache = NewZoneFileCache("", "", &soaSerial)
				zoneFileCache.UpdateVMIRecords(k8stypes.NamespacedName{Namespace: namespace1, Name: vmi1Name},
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic1IP}, Name: nic1Name}})
				Expect(zoneFileCache.soaSerial).To(Equal(6))
			})
		})

		When("interfaces records list contains single vmi", func() {
			BeforeEach(func() {
				zoneFileCache = NewZoneFileCache(nameServerIP, domain, nil)
				isUpdated := zoneFileCache.UpdateVMIRecords(k8stypes.NamespacedName{Namespace: namespace1, Name: vmi1Name},
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic1IP}, Name: nic1Name}, {IPs: []string{nic2IP}, Name: nic2Name}})
				Expect(isUpdated).To(BeTrue())
			})

			DescribeTable("Updating interfaces records list", validateUpdateFunc,
				Entry("when new vmi with interfaces list is added",
					vmi2Name,
					namespace1,
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic3IP}, Name: nic3Name}, {IPs: []string{nic4IP}, Name: nic4Name}},
					updated,
					aRecord_nic1_vm1_ns1+
						aRecord_nic2_vm1_ns1+
						aRecord_nic3_vm2_ns1+
						aRecord_nic4_vm2_ns1,
					2,
				),
				Entry("when existing vmi is deleted",
					vmi1Name,
					namespace1,
					nil,
					updated,
					"",
					2,
				),
				Entry("when existing vmi interfaces list is changed",
					vmi1Name,
					namespace1,
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic1IP}, Name: nic1Name}, {IPs: []string{nic4IP}, Name: nic4Name}},
					updated,
					aRecord_nic1_vm1_ns1+
						aRecord_nic4_vm1_ns1,
					2,
				),
				Entry("when existing vmi is not changed but its interfaces order is changed",
					vmi1Name,
					namespace1,
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic2IP}, Name: nic2Name}, {IPs: []string{nic1IP}, Name: nic1Name}},
					notUpdated,
					aRecord_nic1_vm1_ns1+
						aRecord_nic2_vm1_ns1,
					1,
				),
			)
		})

		When("interfaces records list contains multiple vmis", func() {
			BeforeEach(func() {
				zoneFileCache = NewZoneFileCache(nameServerIP, domain, nil)
				isUpdated := zoneFileCache.UpdateVMIRecords(k8stypes.NamespacedName{Namespace: namespace1, Name: vmi1Name},
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic1IP}, Name: nic1Name}, {IPs: []string{nic2IP}, Name: nic2Name}})
				Expect(isUpdated).To(BeTrue())
				isUpdated = zoneFileCache.UpdateVMIRecords(k8stypes.NamespacedName{Namespace: namespace1, Name: vmi2Name},
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic1IP}, Name: nic1Name}, {IPs: []string{nic2IP}, Name: nic2Name}})
				Expect(isUpdated).To(BeTrue())
			})

			DescribeTable("update interfaces records list", validateUpdateFunc,
				Entry("when new vmi with interfaces list is added",
					vmi1Name,
					namespace2,
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic3IP}, Name: nic3Name}, {IPs: []string{nic4IP}, Name: nic4Name}},
					updated,
					aRecord_nic1_vm1_ns1+
						aRecord_nic2_vm1_ns1+
						aRecord_nic1_vm2_ns1+
						aRecord_nic2_vm2_ns1+
						aRecord_nic3_vm1_ns2+
						aRecord_nic4_vm1_ns2,
					3,
				),
				Entry("when existing vmi is deleted",
					vmi1Name,
					namespace1,
					nil,
					updated,
					aRecord_nic1_vm2_ns1+
						aRecord_nic2_vm2_ns1,
					3,
				),
				Entry("when existing vmi interfaces list is changed",
					vmi1Name,
					namespace1,
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic1IP}, Name: nic1Name}, {IPs: []string{nic4IP}, Name: nic4Name}},
					updated,
					aRecord_nic1_vm1_ns1+
						aRecord_nic4_vm1_ns1+
						aRecord_nic1_vm2_ns1+
						aRecord_nic2_vm2_ns1,
					3,
				),
				Entry("when existing vmi is not changed but its interfaces order is changed",
					vmi1Name,
					namespace1,
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic2IP}, Name: nic2Name}, {IPs: []string{nic1IP}, Name: nic1Name}},
					notUpdated,
					aRecord_nic2_vm1_ns1+
						aRecord_nic1_vm1_ns1+
						aRecord_nic1_vm2_ns1+
						aRecord_nic2_vm2_ns1,
					2,
				),
			)
		})

		When("interfaces records list contains vmi with multiple IPs", func() {
			BeforeEach(func() {
				zoneFileCache = NewZoneFileCache(nameServerIP, domain, nil)
			})

			DescribeTable("Updating interfaces records list", validateUpdateFunc,
				Entry("vmi interfaces contain IPv4 and IPv6",
					vmi1Name,
					namespace1,
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{nic1IP, IPv6}, Name: nic1Name}, {IPs: []string{nic2IP, IPv6}, Name: nic2Name}},
					updated,
					aRecord_nic1_vm1_ns1+
						aRecord_nic2_vm1_ns1,
					1,
				),
				Entry("vmi interfaces contain IPv6 only",
					vmi1Name,
					namespace1,
					[]v1.VirtualMachineInstanceNetworkInterface{{IPs: []string{IPv6}, Name: nic1Name}, {IPs: []string{IPv6}, Name: nic2Name}},
					notUpdated,
					"",
					0,
				),
			)
		})
	})
})

func sortRecords(recordsStr string) (sortedRecordsStr string) {
	strArr := strings.Split(recordsStr, "\n")
	sort.Strings(strArr)
	return strings.Join(strArr, "\n")
}
