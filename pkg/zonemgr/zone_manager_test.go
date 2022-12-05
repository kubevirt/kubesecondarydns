package zonemgr

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"os"

	k8stypes "k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Zone Manager functionality", func() {

	const (
		customDomain = "domain.com"
		customNSIP   = "1.2.3.4"
	)

	var zoneMgr *ZoneManager

	BeforeEach(func() {
		os.Setenv("DOMAIN", customDomain)
		os.Setenv("NAME_SERVER_IP", customNSIP)
		zoneMgr = NewZoneManager()
	})

	Context("Initialization", func() {
		It("should fail updating a VMI with no name", func() {
			Expect(zoneMgr.UpdateZone(k8stypes.NamespacedName{Namespace: "ns1"}, nil)).NotTo(Succeed())
		})

		It("should fail updating a VMI with no namespace", func() {
			Expect(zoneMgr.UpdateZone(k8stypes.NamespacedName{Name: "vm1"}, nil)).NotTo(Succeed())
		})

		It("should set custom data", func() {
			Expect(zoneMgr.zoneFileCache.domain).To(Equal("vm." + customDomain))
			Expect(zoneMgr.zoneFileCache.nameServerIP).To(Equal(customNSIP))
		})

		It("should create zone file with correct name", func() {
			Expect(zoneMgr.zoneFile.zoneFileFullName).To(Equal("/zones/db.vm." + customDomain))
		})

		When("zone file already exist", func() {
			It("should update cached SOA serial value", func() {
				//TODO
			})
		})
	})
})
