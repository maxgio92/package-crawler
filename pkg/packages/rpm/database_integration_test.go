//go:build all_tests || all_integration_tests || (integration_tests && database)

package rpm_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/maxgio92/linux-packages/pkg/packages/rpm"
)

var _ = Describe("Database search integration", func() {
	var (
		ctx = context.Background()
	)

	Context("With repo searcher", Ordered, func() {
		var (
			search   = rpm.NewDBSearcher()
			sourceCh = make(chan string)
			destCh   = make(chan string)
			actual   []string
			seed     = StreamMirror
		)
		BeforeAll(func() {
			// Test producer.
			go func() {
				sourceCh <- seed
				close(sourceCh)
			}()

			// Stage.
			repoCh := rpm.NewRepoSearcher().Run(ctx, sourceCh)

			// Stage.
			destCh = search.Run(ctx, repoCh)

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
			Expect(len(actual)).To(Equal(980))
		})
	})
})
