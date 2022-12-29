package zone_file_cache

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestZoneFileCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zone File Cache Suite")
}
