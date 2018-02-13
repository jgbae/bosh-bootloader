package vsphere

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVSphere(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cloudconfig/vsphere")
}
