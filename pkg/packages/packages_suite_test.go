//go:build all_tests || all_unit_tests || all_integration_tests || unit_tests || integration_tests

package packages_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRpm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Packages Suite")
}
