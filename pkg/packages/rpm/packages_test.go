//go:build all_tests || all_unit_tests || (unit_tests && packages && rpm)

package rpm_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/maxgio92/linux-packages/pkg/packages"
	"github.com/maxgio92/linux-packages/pkg/packages/rpm"
)

var _ = Describe("Packages search", func() {
	var (
		search *rpm.PackageSearch
		ctx    = context.Background()
	)

	Context("With names", func() {
		BeforeEach(func() {
			search = rpm.NewPackageSearcher(
				rpm.WithPackageNames("vim-common"),
			)
		})
		Context("with seed URLs", Ordered, func() {
			var (
				sourceCh = make(chan string)
				destCh   = make(chan *packages.Package)
				actual   []*packages.Package
				seeds    = AppStreamDBs
			)
			BeforeAll(func() {
				// Test producer.
				go func() {
					for _, v := range seeds {
						sourceCh <- v
					}
					close(sourceCh)
				}()

				// Stage.
				destCh = search.Run(ctx, sourceCh)

				// Test sink.
				for v := range destCh {
					actual = append(actual, v)
				}
			})
			It("Should not fail", func() {
				Expect(destCh).ToNot(BeNil())
			})
			It("Should stage results", func() {
				Expect(actual).ToNot(BeEmpty())
			})
			It("Should stage correct results", func() {
				Expect(len(actual)).To(Equal(VimCommonPackageCountInPrimaryDB))
			})
		})
		Context("with seed URLs closed channel", Ordered, func() {
			var (
				sourceCh = make(chan string)
				destCh   = make(chan *packages.Package)
				res      []*packages.Package
			)
			BeforeAll(func() {
				// Noop test producer.
				go func() {
					close(sourceCh)
				}()

				// Stage.
				destCh = search.Run(ctx, sourceCh)

				// Test sink.
				for v := range destCh {
					res = append(res, v)
				}
			})
			It("Should not fail", func() {
				Expect(destCh).ToNot(BeNil())
			})
			It("Should not stage results", func() {
				Expect(res).To(BeEmpty())
				Expect(len(res)).To(Equal(0))
			})
		})
	})
})
