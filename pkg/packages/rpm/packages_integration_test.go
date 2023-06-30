//go:build all_tests || all_integration_tests || (integration_tests && packages)

package rpm_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/maxgio92/linux-packages/pkg/packages"
	"github.com/maxgio92/linux-packages/pkg/packages/rpm"
)

var _ = Describe("Packages search integration", func() {
	var (
		search *rpm.PackageSearch
		ctx    = context.Background()
	)

	Context("With database searcher", func() {
		BeforeEach(func() {
			search = rpm.NewPackageSearcher(
				rpm.WithPackageNames("vim-common"),
			)
		})
		Context("with seed URLs", Ordered, func() {
			var (
				dbCh   = make(chan string)
				repoCh = make(chan string)
				destCh = make(chan *packages.Package)
				actual []*packages.Package
			)
			BeforeAll(func() {
				// Stage.
				go func() {
					defer close(repoCh)
					for k, _ := range StreamRepos {
						repoCh <- StreamRepos[k]
					}
				}()

				// Stage.
				dbCh = rpm.NewDBSearcher().Run(context.Background(), repoCh)

				// Stage.
				destCh = search.Run(ctx, dbCh)

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
	})
})
