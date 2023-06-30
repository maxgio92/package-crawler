//go:build all_tests || all_unit_tests || all_integration_tests || unit_tests || integration_tests

package centos_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/vitorsalgado/mocha/v3"
)

var m *mocha.Mocha

func TestRpm(t *testing.T) {
	m = runMockMirror(t)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Centos Suite")
}

var _ = BeforeSuite(func() {
	Expect(m.URL()).ToNot(BeEmpty())
})
