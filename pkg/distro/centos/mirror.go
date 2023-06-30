package centos

import (
	"context"
	"github.com/sirupsen/logrus"
	"sync"

	wfind "github.com/maxgio92/wfind/pkg/find"
)

type MirrorRootsSearcher struct{}

type Option func(o *MirrorRootsSearcher)

func NewMirrorRootSearcher(options ...Option) *MirrorRootsSearcher {
	mrs := new(MirrorRootsSearcher)
	for _, f := range options {
		f(mrs)
	}

	return mrs
}

func (c *MirrorRootsSearcher) Run(_ context.Context, sourceCh chan string) chan string {
	destCh := make(chan string)

	wg := sync.WaitGroup{}

	for source := range sourceCh {
		source := source
		logrus.WithField("mirror", source).Warn("receive")
		wg.Add(1)
		go func() {
			defer wg.Done()

			finder := wfind.NewFind(
				wfind.WithSeedURLs([]string{source}),
				wfind.WithFilenameRegexp(VersionRegex),
				wfind.WithFileType(wfind.FileTypeDir),
				wfind.WithRecursive(false),
				wfind.WithAsync(true),
				wfind.WithContextDeadlineRetryBackOff(wfind.DefaultExponentialBackOffOptions),
				wfind.WithConnResetRetryBackOff(wfind.DefaultExponentialBackOffOptions),
				wfind.WithConnTimeoutRetryBackOff(wfind.DefaultExponentialBackOffOptions),
			)

			found, err := finder.Find()
			if err != nil {
				logrus.Debug(err)
			}
			if found != nil {
				for _, v := range found.URLs {
					v := v
					logrus.WithField("root", v).Warn("send")
					destCh <- v
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
