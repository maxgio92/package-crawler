package centos

const (
	MirrorEdge    = "https://mirrors.edge.kernel.org/centos/"
	MirrorArchive = "https://archive.kernel.org/centos-vault/"
	//VersionRegex  = `^(0|[1-9]\d*)(\.(0|[1-9]\d*)?)?(\.(0|[1-9]\d*)?)?(-[a-zA-Z\d][-a-zA-Z.\d]*)?(\+[a-zA-Z\d][-a-zA-Z.\d]*)?\/?$`
	VersionRegex = `^.+\/?$`
	keyArch      = "arch"
)

var (
	defaultVersions = []string{"8-stream"}
	defaultReposT   = []string{
		"/AppStream/{{ .arch }}/os/repodata/repomd.xml",
		"/BaseOS/{{ .arch }}/os/repodata/repomd.xml",
		"/Devel/{{ .arch }}/os/repodata/repomd.xml",
		"/os/{{ .arch }}/repodata/repomd.xml",
		"/os/{{ .arch }}/repodata/repomd.xml",
	}
	defaultArchs = []string{"x86_64", "aarch64", "i686", "ppc64le"}
)
