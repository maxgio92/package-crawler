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
	names        []string
	repos        []string
	reposAll     bool
	reposDefault bool
	archs        []string

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

func WithRepoTemplates(repos ...string) PackageSearchOption {
	return func(search *PackageSearch) {
		search.repos = repos
	}
}

func WithArchs(archs ...string) PackageSearchOption {
	return func(search *PackageSearch) {
		search.archs = archs
	}
}

func WithAllRepos(reposAll bool) PackageSearchOption {
	return func(search *PackageSearch) {
		search.reposAll = reposAll
	}
}

func WithDefaultRepos(reposDefault bool) PackageSearchOption {
	return func(search *PackageSearch) {
		search.reposDefault = reposDefault
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

	switch s.reposAll {
	case true:
		data = rpm.NewRepoSearcher(rpm.WithRepoLogger(s.logger)).Run(ctx, data)
	case false:
		repos := DefaultRepos()
		if !s.reposDefault && len(s.repos) > 0 && len(s.archs) > 0 {
			t := template.NewMultiplexTemplate(
				template.WithTemplates(s.repos...),
				template.WithVariables(map[string][]string{keyArch: s.archs}),
			)

			repos, _ = t.Run()
		}
		data = stubStage(ctx, data, repos)
	default:
		data = stubStage(ctx, data, DefaultRepos())
	}

	data = rpm.NewDBSearcher(rpm.WithDBLogger(s.logger)).Run(ctx, data)

	return rpm.NewPackageSearcher(
		rpm.WithPackageNames(s.names...),
		rpm.WithPackageLogger(s.logger),
	).Run(ctx, data)
}

func DefaultRepos() []string {
	t := template.NewMultiplexTemplate(
		template.WithTemplates(DefaultReposT...),
		template.WithVariables(map[string][]string{keyArch: DefaultArchs}),
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
