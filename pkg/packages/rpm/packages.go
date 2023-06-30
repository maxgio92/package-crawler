package rpm

import (
	"compress/gzip"
	"context"
	"encoding/xml"
	"github.com/antchfx/xmlquery"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/maxgio92/linux-packages/internal/network"
	"github.com/maxgio92/linux-packages/pkg/packages"
)

type Package struct {
	XMLName     xml.Name        `xml:"package"`
	Name        string          `xml:"name"`
	Arch        string          `xml:"arch"`
	Version     PackageVersion  `xml:"version"`
	Summary     string          `xml:"summary"`
	Description string          `xml:"description"`
	Packager    string          `xml:"packager"`
	Time        PackageTime     `xml:"time"`
	Size        PackageSize     `xml:"size"`
	Location    PackageLocation `xml:"location"`
	Format      PackageFormat   `xml:"format"`
	url         string
	fileReaders []io.Reader
}

type PackageVersion struct {
	XMLName xml.Name `xml:"version"`
	Epoch   string   `xml:"epoch,attr"`
	Ver     string   `xml:"ver,attr"`
	Rel     string   `xml:"rel,attr"`
}

type PackageTime struct {
	File  string `xml:"file,attr"`
	Build string `xml:"build,attr"`
}

type PackageSize struct {
	Package   string `xml:"package,attr"`
	Installed string `xml:"installed,attr"`
	Archive   string `xml:"archive,attr"`
}

type PackageLocation struct {
	XMLName xml.Name `xml:"location"`
	Href    string   `xml:"href,attr"`
}

type PackageFormat struct {
	XMLName     xml.Name           `xml:"format"`
	License     string             `xml:"license"`
	Vendor      string             `xml:"vendor"`
	Group       string             `xml:"group"`
	Buildhost   string             `xml:"buildhost"`
	HeaderRange PackageHeaderRange `xml:"header-range"`
	Requires    PackageRequires    `xml:"requires"`
	Provides    PackageProvides    `xml:"provides"`
}

type PackageHeaderRange struct {
	Start string `xml:"start,attr"`
	End   string `xml:"end,attr"`
}

type PackageProvides struct {
	XMLName xml.Name `xml:"provides"`
	Entries []Entry  `xml:"entry"`
}

type PackageRequires struct {
	XMLName xml.Name `xml:"requires"`
	Entries []Entry  `xml:"entry"`
}

type Entry struct {
	XMLName xml.Name `xml:"entry"`
	Name    string   `xml:"name,attr"`
}

func (p *Package) Describe() string { return p.Description }

type PackageSearch struct {
	names []string
}

type PackageSearchOption func(s *PackageSearch)

const (
	dataPackageXPath = "//package"
)

func WithPackageNames(names ...string) PackageSearchOption {
	return func(ps *PackageSearch) {
		ps.names = names
	}
}

func NewPackageSearcher(o ...PackageSearchOption) *PackageSearch {
	ps := new(PackageSearch)
	for _, f := range o {
		f(ps)
	}

	return ps
}

func (ps *PackageSearch) validate() error {
	if len(ps.names) == 0 {
		return ErrSearchPackagaNameMissing
	}

	return nil
}

// TODO: avoid sending duplicate packages to che channel.
func (ps *PackageSearch) Run(ctx context.Context, sourceCh chan string) chan *packages.Package {
	destCh := make(chan *packages.Package)
	if ps.validate() != nil {
		return destCh
	}

	wg := sync.WaitGroup{}

	for source := range sourceCh {
		source := source
		logrus.WithField("database", source).Warn("receive")
		wg.Add(1)
		go func() {
			defer wg.Done()
			pxml, err := ps.packagesXMLFromDB(ctx, source)
			if err != nil {
				return
			}
			repoURL, err := url.JoinPath(strings.Split(source, DirRepodata)[0], DirRepodata)
			if err != nil {
				return
			}
			if pxml != nil {
				for pkg := range packagesFromXML(ctx, pxml) {
					pkg := pkg
					pkgURL, err := url.JoinPath(repoURL, pkg.Location.Href)
					if err != nil {
						break
					}
					logrus.WithField("package", pkgURL).Warn("send")
					destCh <- packages.NewPackage(
						packages.WithName(pkg.Name),
						packages.WithVersion(pkg.Version.Ver+"+"+pkg.Version.Rel),
						packages.WithLocation(pkgURL),
						packages.WithArchitecture(pkg.Arch),
					)
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

func (ps *PackageSearch) packagesXMLFromDB(ctx context.Context, dbURL string) ([]*xmlquery.Node, error) {
	if !strings.HasSuffix(path.Base(dbURL), ".xml.gz") {
		return nil, ErrDBFormatNotSupported
	}

	if err := ps.validate(); err != nil {
		return nil, err
	}
	name := ps.names[0]

	u, err := url.Parse(dbURL)
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

		// Dumb retry logic.
		// TODO: implement smart cyclic retry with increasing backoff.
		time.Sleep(1 * time.Second)
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	var packages []*xmlquery.Node
	if resp.StatusCode == http.StatusOK && resp.Body != nil {
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()

		sp, err := xmlquery.CreateStreamParser(
			gr,
			dataPackageXPath,
			dataPackageXPath+"[name='"+name+"']")
		if err != nil {
			return nil, err
		}

		for {
			n, e := sp.Read()
			if e != nil {
				break
			}

			packages = append(packages, n)
		}
	}

	return packages, nil
}

func packagesFromXML(_ context.Context, nodes []*xmlquery.Node) chan *Package {
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))

	outCh := make(chan *Package)

	for k, _ := range nodes {
		go func() {
			defer wg.Done()

			pkg := &Package{}

			err := xml.Unmarshal([]byte(nodes[k].OutputXML(true)), pkg)
			if err == nil {
				outCh <- pkg
			} else {
				logrus.Debug(err)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
