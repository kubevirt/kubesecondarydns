package zonemgr

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"os"
)

var _ = Describe("disk zone file maintenance", func() {

	const (
		zoneFileName           = "zones/db.vm"
		zoneFileContent        = "zone file content"
		zoneFileUpdatedContent = "zone file updated content"
	)
	var zoneFile *ZoneFile

	BeforeEach(func() {
		zoneFile = NewZoneFile(zoneFileName)
		Expect(os.Mkdir("zones", 0777)).To(Succeed())
	})
	AfterEach(func() {
		Expect(os.RemoveAll("zones")).To(Succeed())
	})

	testFileFunc := func(expectedFileContent string) {
		Expect(zoneFile.writeFile(expectedFileContent)).To(Succeed())
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
	})

	When("zone file already exist", func() {
		BeforeEach(func() {
			Expect(os.WriteFile(zoneFileName, []byte(zoneFileContent), zoneFilePerm)).To(Succeed())
		})
		AfterEach(func() {
			Expect(os.RemoveAll(zoneFileName)).To(Succeed())
		})

		It("should override a zone file on disk", func() {
			testFileFunc(zoneFileUpdatedContent)
		})
	})
})
