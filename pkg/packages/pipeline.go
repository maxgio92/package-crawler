package packages

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
)

// *****************************************************************
// Pipelines
// *****************************************************************

func RunSearchPipeline(ctx context.Context, producer Producer, search SearchStageRunner, stages ...StageRunner) chan *Package {
	data := producer.Produce(ctx)
	for _, stage := range stages {
		data = stage.Run(ctx, data)
	}

	return search.Run(ctx, data)
}

// *****************************************************************
// Producers
// *****************************************************************

type Producer interface {
	Produce(ctx context.Context) chan string
}

type GenericProducer struct {
	seeds []string
}

func NewGenericProducer(seeds ...string) *GenericProducer {
	return &GenericProducer{seeds: seeds}
}

// Produce is a mirror producer that streams mirror URLs.
func (p *GenericProducer) Produce(_ context.Context) chan string {
	data := make(chan string)

	wg := new(sync.WaitGroup)

	for _, v := range p.seeds {
		v := v
		wg.Add(1)
		go func() {
			defer wg.Done()
			logrus.WithField("seed", v).Warn("send")
			data <- v
		}()
	}
	go func() {
		wg.Wait()
		close(data)
	}()

	return data
}

// *****************************************************************
// Stages
// *****************************************************************

type StageRunner interface {
	Run(ctx context.Context, source chan string) chan string
}

// SearchStageRunner is a pipeline stage runner.
type SearchStageRunner interface {
	Run(ctx context.Context, dbURLs chan string) chan *Package
}

// PackageDescriptor describes a package alongside its metadata.
type PackageDescriptor interface {
	Describe() string
	Version() string
	Architecture() string
}

// PackageLocator describes a package alongside its metadata.
type PackageLocator interface {
	Locate() string
}

type Package struct {
	name         string
	version      string
	location     string
	architecture string
}

type PackageOption func(o *Package)

func WithName(name string) PackageOption {
	return func(o *Package) {
		o.name = name
	}
}

func WithVersion(version string) PackageOption {
	return func(o *Package) {
		o.version = version
	}
}

func WithLocation(location string) PackageOption {
	return func(o *Package) {
		o.location = location
	}
}

func WithArchitecture(architecture string) PackageOption {
	return func(o *Package) {
		o.architecture = architecture
	}
}

func NewPackage(options ...PackageOption) *Package {
	pkg := new(Package)
	for _, f := range options {
		f(pkg)
	}

	return pkg
}

func (p *Package) Describe() string     { return p.name }
func (p *Package) Version() string      { return p.version }
func (p *Package) Locate() string       { return p.location }
func (p *Package) Architecture() string { return p.architecture }

type PackageConverter interface {
	Convert(ctx context.Context, r io.Reader) (io.Reader, error)
}

type PackageDownloader interface {
	Download(ctx context.Context, location string) (io.Reader, error)
}

// *****************************************************************
// Sinks
// *****************************************************************

type SinkRunner interface {
	Run(ctx context.Context, source chan string) error
}
