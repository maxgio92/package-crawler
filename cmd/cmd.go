package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/maxgio92/linux-packages/internal/output/log"
	"github.com/maxgio92/linux-packages/pkg/distro/centos"
)

const (
	ProgramName = "packages"
	flagCentos  = "centos"
	flagAll     = "--all"
)

var (
	LogLevel = logrus.DebugLevel
)

// TODO: use Cobra.
func Run() {
	if len(os.Args) < 3 {
		fmt.Println("Please specify a distro and package name as arguments")
		fmt.Printf("usage: %s distro|--all package-name\n", ProgramName)
		os.Exit(1)
	}

	switch os.Args[1] {
	case flagCentos:
		runCentos(os.Args[2])
	case flagAll:
		runCentos(os.Args[2])
	default:
		fmt.Println("distro not supported")
		os.Exit(1)
	}
}

func runCentos(packageName string) {
	logger := log.NewLogger(LogLevel)

	for p := range centos.NewPackageSearch(
		centos.WithPackageNames(packageName),
		centos.WithSearchLogger(logger),
	).Search(context.Background()) {
		logger.
			WithField("name", p.Describe()).
			WithField("version", p.Version()).
			WithField("architecture", p.Architecture()).
			WithField("location", p.Locate()).
			Info()
	}
}
