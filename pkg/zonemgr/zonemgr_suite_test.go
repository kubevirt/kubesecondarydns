package zonemgr_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestZonemgr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zone manager Suite")
}
