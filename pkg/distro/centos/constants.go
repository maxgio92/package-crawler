package centos

const (
	MirrorEdge    = "https://mirrors.edge.kernel.org/centos/"
	MirrorArchive = "https://archive.kernel.org/centos-vault/"
	//VersionRegex  = `^(0|[1-9]\d*)(\.(0|[1-9]\d*)?)?(\.(0|[1-9]\d*)?)?(-[a-zA-Z\d][-a-zA-Z.\d]*)?(\+[a-zA-Z\d][-a-zA-Z.\d]*)?\/?$`
	VersionRegex = `^.+\/?$`
	keyArch      = "arch"
	X86_64       = "x86_64"
	Aarch64      = "aarch64"
	I686         = "i686"
	Ppc64le      = "ppc64le"
)

var (
	defaultVersions = []string{"8-stream"}
	DefaultReposT   = []string{
		"/AppStream/{{ .arch }}/os/repodata/repomd.xml",
		"/BaseOS/{{ .arch }}/os/repodata/repomd.xml",
		"/Devel/{{ .arch }}/os/repodata/repomd.xml",
		"/os/{{ .arch }}/repodata/repomd.xml",
		"/os/{{ .arch }}/repodata/repomd.xml",
	}
	DefaultArchs = []string{X86_64, Aarch64, I686, Ppc64le}
)
