package zone_file_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestZoneFile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zone File Suite")
}
