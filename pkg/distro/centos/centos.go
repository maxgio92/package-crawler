package centos

import (
	"context"
	"net/url"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/maxgio92/linux-packages/pkg/packages"
	"github.com/maxgio92/linux-packages/pkg/packages/rpm"
	"github.com/maxgio92/linux-packages/pkg/template"
)

type PackageSearch struct {
	names  []string
	logger *log.Logger
}

type PackageSearchOption func(s *PackageSearch)

func WithPackageNames(names ...string) PackageSearchOption {
	return func(search *PackageSearch) {
		search.names = names
	}
}

func WithSearchLogger(logger *log.Logger) PackageSearchOption {
	return func(search *PackageSearch) {
		search.logger = logger
	}
}

func NewPackageSearch(o ...PackageSearchOption) *PackageSearch {
	search := new(PackageSearch)
	for _, f := range o {
		f(search)
	}

	return search
}

// Search is a data streaming pipeline.
func (s *PackageSearch) Search(ctx context.Context) chan *packages.Package {
	data := packages.NewGenericProducer(
		packages.WithSeeds(MirrorEdge, MirrorArchive),
		packages.WithLogger(s.logger),
	).Produce(ctx)
	data = NewVersionSearcher(WithMirrorLogger(s.logger)).Run(ctx, data)
	data = stubStage(ctx, data, defaultRepos())
	//data = rpm.NewRepoSearcher(rpm.WithRepoLogger(s.logger)).Run(ctx, data)
	data = rpm.NewDBSearcher(rpm.WithDBLogger(s.logger)).Run(ctx, data)

	return rpm.NewPackageSearcher(
		rpm.WithPackageNames(s.names...),
		rpm.WithPackageLogger(s.logger),
	).Run(ctx, data)
}

func defaultRepos() []string {
	t := template.NewMultiplexTemplate(
		template.WithTemplates(defaultReposT...),
		template.WithVariables(map[string][]string{keyArch: defaultArchs}),
	)

	repos, _ := t.Run()

	return repos
}

func stubStage(_ context.Context, seedsCh chan string, data []string) chan string {
	destCh := make(chan string, len(data))

	wg := sync.WaitGroup{}

	for seed := range seedsCh {
		seed := seed
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k, _ := range data {
				merged, err := url.JoinPath(seed, data[k])
				if err == nil {
					destCh <- merged
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(destCh)
	}()

	return destCh
}
