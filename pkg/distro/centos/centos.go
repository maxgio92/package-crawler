package centos

import (
	"context"
	"net/url"
	"sync"

	"github.com/maxgio92/linux-packages/pkg/packages"
	"github.com/maxgio92/linux-packages/pkg/packages/rpm"
)

// SearchPackages is a data streaming pipeline.
func SearchPackages(ctx context.Context, names ...string) chan *packages.Package {
	data := packages.NewGenericProducer(MirrorEdge, MirrorArchive).Produce(ctx)
	data = NewMirrorRootSearcher().Run(ctx, data)
	data = rpm.NewRepoSearcher().Run(ctx, data)
	data = rpm.NewDBSearcher().Run(ctx, data)

	return rpm.NewPackageSearcher(
		rpm.WithPackageNames(names...),
	).Run(ctx, data)
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
