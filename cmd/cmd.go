package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/maxgio92/linux-packages/pkg/distro/centos"
)

const (
	Centos      = "centos"
	ProgramName = "packages"
)

// TODO: use Cobra.
func Run() {
	if len(os.Args) < 3 {
		fmt.Println("Please specify a distro and package name as arguments")
		fmt.Printf("usage: %s distro|--all package-name\n", ProgramName)
		os.Exit(1)
	}

	switch os.Args[1] {
	case Centos:
		for p := range centos.SearchPackages(context.Background(), os.Args[2]) {
			fmt.Printf("Name: %s\tVersion: %s\tArchitecture: %s\tLocation: %s\n",
				p.Describe(), p.Version(), p.Architecture(), p.Locate())
		}
	case "--all":
		for p := range centos.SearchPackages(context.Background(), os.Args[2]) {
			fmt.Printf("Name: %s\tVersion: %s\tArchitecture: %s\tLocation: %s\n",
				p.Describe(), p.Version(), p.Architecture(), p.Locate())
		}
	default:
		fmt.Println("distro not supported")
	}
}
