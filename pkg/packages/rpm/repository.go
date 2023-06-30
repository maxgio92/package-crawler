package rpm

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"

	wfind "github.com/maxgio92/wfind/pkg/find"
)

type RepoSearcher struct{}

type RepoSearchOption func(s *RepoSearcher)

const (
	Repomd = "repomd.xml$"
)

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
		logrus.WithField("mirror", source).Warn("receive")
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
				fmt.Println(errors.Wrap(err, "repoSearch"))
			}
			if found != nil {
				for _, v := range found.URLs {
					v := v
					logrus.WithField("repository", v).Warn("send")
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
