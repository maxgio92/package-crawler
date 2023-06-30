package rpm

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/maxgio92/linux-packages/internal/network"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"
)

type Data struct {
	Type     string   `xml:"type,attr"`
	Location Location `xml:"location"`
}

type Location struct {
	Href string `xml:"href,attr"`
}

const (
	metadataDataXPath = "//repomd/data"
)

type DBSearch struct{}

type DBSearchOption func(s *DBSearch)

func NewDBSearcher(o ...DBSearchOption) *DBSearch {
	dbs := new(DBSearch)
	for _, f := range o {
		f(dbs)
	}

	return dbs
}

// Run runs a pipeline stage of which the output is a channel of database URL strings.
// The source of the stage is a channel of repository metadata (repomd) URL strings.
func (ds *DBSearch) Run(ctx context.Context, sourceCh chan string) chan string {
	destCh := make(chan string)

	wg := sync.WaitGroup{}

	for source := range sourceCh {
		source := source
		logrus.WithField("repo", source).Warn("receive")
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, err := url.Parse(source)
			if err != nil {
				return
			}

			reporoot := path.Dir(path.Dir(u.Path))
			u.Path = reporoot

			dbs := []Data{}

			d, err := getPrimaryDBMetadatasFromRepoMetadataURL(ctx, source)
			if err != nil {
				return
			}
			dbs = append(dbs, d...)

			for k, _ := range dbs {
				if u, err := url.JoinPath(u.String(), dbs[k].Location.Href); err == nil {
					logrus.WithField("database", u).Warn("send")
					destCh <- u
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

func getPrimaryDBMetadatasFromRepoMetadataURL(ctx context.Context, metadataURL string) ([]Data, error) {
	var dbs []Data

	u, err := url.Parse(metadataURL)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: network.DefaultClientTransport,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		if !errors.Is(err, unix.ECONNRESET) {
			return nil, err
		}

		// Dump retry logic.
		// TODO: implement smart cyclic retry with increasing backoff.
		time.Sleep(1 * time.Second)
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	}
	if resp.Body == nil {
		return nil, ErrDBMetadataResponseEmpty
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK || resp.Body == nil {
		return nil, fmt.Errorf("unexpected response: %d", resp.StatusCode)
	}

	doc, err := xmlquery.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	datasXML, err := xmlquery.QueryAll(doc, metadataDataXPath)
	if err != nil {
		return nil, err
	}

	for _, v := range datasXML {
		data := &Data{}

		err = xml.Unmarshal([]byte(v.OutputXML(true)), data)
		if err != nil {
			return nil, err
		}

		if data.Type == DBTypePrimary {
			dbs = append(dbs, *data)
		}
	}

	return dbs, nil
}

// TODO: get package metadata
// TODO: download package content
