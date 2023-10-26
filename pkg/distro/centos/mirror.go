package centos

import (
	"context"
	log "github.com/sirupsen/logrus"
	"sync"

	wfind "github.com/maxgio92/wfind/pkg/find"
)

type VersionSearcher struct {
	logger *log.Logger
}

type MirrorSearchOption func(o *VersionSearcher)

func WithMirrorLogger(logger *log.Logger) MirrorSearchOption {
	return func(search *VersionSearcher) {
		search.logger = logger
	}
}

func NewVersionSearcher(options ...MirrorSearchOption) *VersionSearcher {
	mrs := new(VersionSearcher)
	for _, f := range options {
		f(mrs)
	}

	return mrs
}

func (c *VersionSearcher) Run(_ context.Context, sourceCh chan string) chan string {
	destCh := make(chan string)

	wg := sync.WaitGroup{}

	for source := range sourceCh {
		source := source
		c.logger.WithField("mirror", source).Debug("receive")
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
				c.logger.WithError(err).Debug("error searching centos versions")
			}
			if found != nil {
				for _, v := range found.URLs {
					v := v
					c.logger.WithField("version", v).Debug("send")
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
