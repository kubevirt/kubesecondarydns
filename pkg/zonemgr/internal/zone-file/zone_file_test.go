package zone_file_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"os"

	"github.com/kubevirt/kubesecondarydns/pkg/zonemgr/internal/zone-file"
)

var _ = Describe("disk zone file maintenance", func() {

	const (
		zoneFileName           = "zones/db.vm"
		zoneFileContent        = "zone file content"
		zoneFileUpdatedContent = "zone file updated content"
		headerSoaSerial        = "$ORIGIN vm. \n$TTL 3600 \n@ IN SOA ns.vm. email.vm. (12345 3600 3600 1209600 3600)\n"
	)
	var zoneFile zone_file.ZoneFileInterface

	BeforeEach(func() {
		zoneFile = zone_file.NewZoneFile(zoneFileName)
		Expect(os.Mkdir("zones", 0777)).To(Succeed())
	})
	AfterEach(func() {
		Expect(os.RemoveAll("zones")).To(Succeed())
	})

	testFileFunc := func(expectedFileContent string) {
		Expect(zoneFile.WriteFile(expectedFileContent)).To(Succeed())
		content, err := os.ReadFile(zoneFileName)
		Expect(err).To(BeNil())
		Expect(string(content)).To(Equal(expectedFileContent))
	}

	When("zone file does not exist", func() {
		AfterEach(func() {
			Expect(os.RemoveAll(zoneFileName)).To(Succeed())
		})

		It("should create a zone file on disk", func() {
			testFileFunc(zoneFileContent)
		})

		It("should return nil when try to read SOA serial", func() {
			soaSerial, err := zoneFile.ReadSoaSerial()
			Expect(err).ToNot(HaveOccurred())
			Expect(soaSerial).To(BeNil())
		})
	})

	When("zone file already exist", func() {
		AfterEach(func() {
			Expect(os.RemoveAll(zoneFileName)).To(Succeed())
		})

		It("should override a zone file on disk", func() {
			Expect(os.WriteFile(zoneFileName, []byte(zoneFileContent), 0644)).To(Succeed())
			testFileFunc(zoneFileUpdatedContent)
		})

		It("should read SOA serial", func() {
			Expect(os.WriteFile(zoneFileName, []byte(headerSoaSerial), 0777)).To(Succeed())
			soaSerial, err := zoneFile.ReadSoaSerial()
			Expect(err).ToNot(HaveOccurred())
			Expect(*soaSerial).To(Equal(12345))
		})
	})
})
