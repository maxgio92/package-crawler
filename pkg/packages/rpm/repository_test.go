//go:build all_tests || all_unit_tests || (unit_tests && repository && rpm)

package rpm_test

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	wfind "github.com/maxgio92/wfind/pkg/find"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/maxgio92/linux-packages/pkg/packages/rpm"
)

var _ = Describe("Repository search", func() {
	var (
		search *rpm.RepoSearcher
		ctx    = context.Background()
		less   = func(a, b string) bool { return a < b }
	)

	Context("With names", func() {
		BeforeEach(func() {
			search = rpm.NewRepoSearcher()
		})
		Context("with seed URLs", Ordered, func() {
			var (
				sourceCh         = make(chan string)
				destCh           = make(chan string)
				actual           []string
				expected         []string
				seed             = StreamMirror
				equalIgnoreOrder bool
			)
			BeforeAll(func() {
				// Test producer.
				go func() {
					sourceCh <- seed
					close(sourceCh)
				}()

				// Stage.
				destCh = search.Run(ctx, sourceCh)

				// Test sink.
				for v := range destCh {
					actual = append(actual, v)
				}

				// Expected data.
				f := wfind.NewFind(
					wfind.WithSeedURLs([]string{seed}),
					wfind.WithFilenameRegexp(rpm.Repomd),
					wfind.WithFileType(wfind.FileTypeReg),
					wfind.WithRecursive(true),
					wfind.WithAsync(true),
					wfind.WithContextDeadlineRetryBackOff(wfind.DefaultExponentialBackOffOptions),
					wfind.WithConnResetRetryBackOff(wfind.DefaultExponentialBackOffOptions),
					wfind.WithConnTimeoutRetryBackOff(wfind.DefaultExponentialBackOffOptions),
				)
				e, _ := f.Find()
				expected = e.URLs
				equalIgnoreOrder = cmp.Diff(
					actual,
					expected,
					cmpopts.SortSlices(less),
				) == ""
			})
			It("Should not fail", func() {
				Expect(destCh).ToNot(BeNil())
			})
			It("Should stage results", func() {
				Expect(actual).ToNot(BeEmpty())
				Expect(len(actual)).To(Equal(len(expected)))
			})
			It("Should stage valid results", func() {
				Expect(equalIgnoreOrder).To(BeTrue())
			})
		})
		Context("with seed URLs closed channel", Ordered, func() {
			var (
				sourceCh = make(chan string)
				destCh   = make(chan string)
				res      []string
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
