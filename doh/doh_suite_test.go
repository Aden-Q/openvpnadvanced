package doh_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDoh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Doh Suite")
}
