package zonemgr_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"os"

	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/kubevirt/kubesecondarydns/pkg/zonemgr"
	"github.com/kubevirt/kubesecondarydns/pkg/zonemgr/internal/zone-file"
	"github.com/kubevirt/kubesecondarydns/pkg/zonemgr/internal/zone-file-cache"
)

const (
	customDomain = "domain.com"
	customNSIP   = "1.2.3.4"
)

var _ = Describe("Zone Manager functionality", func() {

	BeforeEach(func() {
		os.Setenv("DOMAIN", customDomain)
		os.Setenv("NAME_SERVER_IP", customNSIP)
	})

	Context("Initialization", func() {
		It("should fail updating a VMI with no name", func() {
			zoneMgr, err := zonemgr.NewZoneManager()
			Expect(err).ToNot(HaveOccurred())
			Expect(zoneMgr.UpdateZone(k8stypes.NamespacedName{Namespace: "ns1"}, nil)).NotTo(Succeed())
		})

		It("should fail updating a VMI with no namespace", func() {
			zoneMgr, err := zonemgr.NewZoneManager()
			Expect(err).ToNot(HaveOccurred())
			Expect(zoneMgr.UpdateZone(k8stypes.NamespacedName{Name: "vm1"}, nil)).NotTo(Succeed())
		})

		It("should set custom data", func() {
			_, err := zonemgr.NewZoneManagerWithParams(newZoneFileCacheStub, zone_file.NewZoneFile)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create zone file with correct name", func() {
			_, err := zonemgr.NewZoneManagerWithParams(zone_file_cache.NewZoneFileCache, newZoneFileStub)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func newZoneFileCacheStub(nameServerIP string, domain string, soaSerial *int) *zone_file_cache.ZoneFileCache {
	expectedNameServerIP := customNSIP
	expectedDomain := "vm." + customDomain
	Expect(nameServerIP).To(Equal(expectedNameServerIP))
	Expect(domain).To(Equal(expectedDomain))
	return nil
}

func newZoneFileStub(fileName string) zone_file.ZoneFileInterface {
	expectedZoneFileName := "/zones/db.vm." + customDomain
	Expect(fileName).To(Equal(expectedZoneFileName))
	return &ZoneFileStub{}
}

type ZoneFileStub struct {
}

func (zoneFileStub *ZoneFileStub) WriteFile(content string) (err error) {
	return nil
}

func (zoneFileStub *ZoneFileStub) ReadSoaSerial() (*int, error) {
	return nil, nil
}
