//go:build all_tests || all_unit_tests || (unit_tests && packages)

package packages_test

import (
	"context"
	"github.com/maxgio92/linux-packages/pkg/packages"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/url"
	"sync"
)

var _ = Describe("Pipeline run", func() {
	var (
		ctx = context.Background()
	)

	Context("With stub stages", Ordered, func() {
		var (
			destCh = make(chan *packages.Package)
			actual []*packages.Package
		)
		BeforeAll(func() {
			// Producer.
			producer := packages.NewGenericProducer(
				"https://mirrors.edge.kernel.org/centos/",
			)

			// Stages.
			stages := []packages.StageRunner{
				newStageStub("8-stream"),
				newStageStub("AppStream/x86_64/os/"),
				newStageStub(
					"repodata/6f9196426e6cc8c57c90460168e3d445960019f070a5c539ac7ac1a1f2eab467-modules.yaml.xz",
					"repodata/6b5f696637d9dd2d00fc9940ab33c9f5094092a72f1c1cb4a6e3aec57a8d9a81-primary.xml.gz",
					"repodata/6ad4f8a3a9dc71608b58936e91ce4658ce721db4f310526535db835bca7e1d45-filelists.xml.gz",
					"repodata/b62f4bc64314d5e16c267b9fcd3675ba94584c8e95d5961b2388f0cf3bda42de-other.xml.gz",
					"repodata/2637cf7528bc978c588022488da84bf22f26dcbc5b6fe2a45bbbd7b721969edf-primary.sqlite.xz",
					"repodata/4a7d786d22dd91efeee473f389b5a721f4cd8f8443b3cd5cdac8cd4bf7a970a0-filelists.sqlite.xz",
					"repodata/cbd8a3c3726c9aefbd22c95fb8cf99b05ca21ed04ad3cc0a7f9cc1b77a29fd20-other.sqlite.xz",
					"repodata/0ec907e2460ea55f3a652fcd30e9e73333ebbd206d1f940d7893304d6789f10f-comps-AppStream.x86_64.xml",
					"repodata/8ab3e067e08b2fae103eb31d369f6e3953655b267cf7c9698aac4c028c34240c-comps-AppStream.x86_64.xml.xz",
					"repodata/6f9196426e6cc8c57c90460168e3d445960019f070a5c539ac7ac1a1f2eab467-modules.yaml.xz",
				),
			}

			// Search final stage.
			search := newSearchStageStub(
				"vim-common", "vim-common", "vim-common", "vim-common",
				"vim-common", "vim-common", "vim-common", "vim-common",
			)

			destCh = packages.RunSearchPipeline(ctx, producer, search, stages...)

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
			Expect(len(actual)).To(Equal(8))
		})
	})
})

// stageStub is a pipeline stage stubs that returns channel with static data.
type stageStub struct {
	data []string
}

func newStageStub(data ...string) *stageStub {
	return &stageStub{data: data}
}

func (s *stageStub) Run(_ context.Context, sourceCh chan string) chan string {
	destCh := make(chan string)

	wg := sync.WaitGroup{}

	for source := range sourceCh {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, v := range s.data {
				s, _ := url.JoinPath(source, v)
				destCh <- s
			}
		}()
	}

	go func() {
		wg.Wait()
		close(destCh)
	}()

	return destCh
}

type searchStageStub struct {
	data []string
}

func newSearchStageStub(data ...string) *searchStageStub {
	return &searchStageStub{data: data}
}

func (s *searchStageStub) Run(_ context.Context, _ chan string) chan *packages.Package {
	destCh := make(chan *packages.Package)

	wg := sync.WaitGroup{}

	for _, v := range s.data {
		wg.Add(1)
		go func() {
			defer wg.Done()
			destCh <- packages.NewPackage(v)
		}()
	}

	go func() {
		wg.Wait()
		close(destCh)
	}()

	return destCh
}
