package centos

const (
	MirrorEdge    = "https://mirrors.edge.kernel.org/centos/"
	MirrorArchive = "https://archive.kernel.org/centos-vault/"
	//VersionRegex  = `^(0|[1-9]\d*)(\.(0|[1-9]\d*)?)?(\.(0|[1-9]\d*)?)?(-[a-zA-Z\d][-a-zA-Z.\d]*)?(\+[a-zA-Z\d][-a-zA-Z.\d]*)?\/?$`
	VersionRegex = `^.+\/?$`
)

var (
	defaultVersions = []string{"8-stream"}
	defaultRepos    = []string{
		"/AppStream/x86_64/os/repodata/repomd.xml",
		"/BaseOS/x86_64/os/repodata/repomd.xml",
		"/Devel/x86_64/os/repodata/repomd.xml",
		"/os/x86_64/repodata/repomd.xml",
		"/os/x86_64/repodata/repomd.xml",
	}
)
