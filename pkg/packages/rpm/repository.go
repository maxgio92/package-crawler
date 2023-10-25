package rpm

import (
	"context"
	log "github.com/sirupsen/logrus"
	"sync"

	wfind "github.com/maxgio92/wfind/pkg/find"
)

type RepoSearcher struct {
	logger *log.Logger
}

type RepoSearchOption func(s *RepoSearcher)

const (
	Repomd = "repomd.xml$"
)

func WithRepoLogger(logger *log.Logger) RepoSearchOption {
	return func(search *RepoSearcher) {
		search.logger = logger
	}
}

func NewRepoSearcher(o ...RepoSearchOption) *RepoSearcher {
	rs := new(RepoSearcher)
	for _, f := range o {
		f(rs)
	}

	return rs
}

func (rs *RepoSearcher) Run(_ context.Context, sourceCh chan string) chan string {
	destCh := make(chan string)

	wg := sync.WaitGroup{}

	for source := range sourceCh {
		source := source
		rs.logger.WithField("mirror", source).Debug("receive")
		wg.Add(1)
		go func() {
			defer wg.Done()
			finder := wfind.NewFind(
				wfind.WithSeedURLs([]string{source}),
				wfind.WithFilenameRegexp(Repomd),
				wfind.WithFileType(wfind.FileTypeReg),
				wfind.WithRecursive(true),
				wfind.WithAsync(true),
				wfind.WithContextDeadlineRetryBackOff(wfind.DefaultExponentialBackOffOptions),
				wfind.WithConnResetRetryBackOff(wfind.DefaultExponentialBackOffOptions),
				wfind.WithConnTimeoutRetryBackOff(wfind.DefaultExponentialBackOffOptions),
			)

			found, err := finder.Find()
			if err != nil {
				rs.logger.WithError(err).Warn("error searching repositories")
			}
			if found != nil {
				for _, v := range found.URLs {
					v := v
					rs.logger.WithField("repository", v).Debug("send")
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
