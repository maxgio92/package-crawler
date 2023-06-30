//go:build all_tests || all_unit_tests || (unit_tests && database && rpm)

package rpm_test

import (
	"context"
	"net/http"
	"path"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/maxgio92/linux-packages/pkg/packages/rpm"
)

var _ = Describe("Database search", func() {
	var (
		search *rpm.DBSearch
		ctx    = context.Background()
	)

	Context("With names", func() {
		BeforeEach(func() {
			search = rpm.NewDBSearcher()
		})
		Context("with wrong repository URLs", Ordered, func() {
			var (
				sourceCh = make(chan string)
				destCh   = make(chan string)
				actual   []string
				seed     = BaseOSRepoWrongURL
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
			})
			It("Should not fail", func() {
				Expect(destCh).ToNot(BeNil())
			})
			It("Should not stage results", func() {
				Expect(actual).To(BeEmpty())
				Expect(len(actual)).To(Equal(0))
			})
		})
		Context("with valid repository URLs", Ordered, func() {
			var (
				sourceCh        = make(chan string)
				destCh          = make(chan string)
				actual          []string
				seed            = BaseOSRepo
				areDBFiles      = true
				areExistentURLs = true
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

				for _, v := range actual {
					if !(strings.HasSuffix(path.Base(v), ".xml.gz") ||
						strings.HasSuffix(path.Base(v), ".xml.xz") ||
						strings.HasSuffix(path.Base(v), ".xml") ||
						strings.HasSuffix(path.Base(v), ".sqlite.xz")) {
						areDBFiles = false
					}
					req, _ := http.NewRequestWithContext(ctx, http.MethodGet, v, nil)
					resp, _ := http.DefaultClient.Do(req)
					resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						areExistentURLs = false
					}
				}
			})
			It("Should not fail", func() {
				Expect(destCh).ToNot(BeNil())
			})
			It("Should stage results", func() {
				Expect(actual).ToNot(BeEmpty())
				Expect(len(actual)).To(Equal(8))
			})
			It("Should stage existent URLs", func() {
				Expect(areExistentURLs).To(BeTrue())
			})
			It("Should stage URLS of XML DB files", func() {
				Expect(areDBFiles).To(BeTrue())
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
