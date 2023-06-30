//go:build all_tests || all_unit_tests || (unit_tests && centos && mirror)

package centos_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/url"
	"testing"

	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"

	"github.com/maxgio92/linux-packages/pkg/distro/centos"
)

const (
	homedir  string = "mirror"
	filename string = "File"
	dirname  string = "Dir"
)

var (
	files       = []string{"hello", "world"}
	subdirs     = []string{"1", "2", "3"}
	homedirBody = fmt.Sprintf(`
<html>
<head><title>Index of %s/</title></head>
<body>
<h1>Index of /%s/</h1><hr><pre><a href="../">../</a>
<a href="%s/">%s/</a>
<a href="%s/">%s/</a>
<a href="%s/">%s/</a>
<a href="%s">%s/</a>
<a href="%s">%s/</a>
</pre><hr></body>
</html>
`, homedir,
		homedir,
		subdirs[0], subdirs[0],
		subdirs[1], subdirs[1],
		subdirs[2], subdirs[2],
		files[0], files[0],
		files[1], files[1])
	subdirBodyF = `
<html>
<head><title>Index of %s/%s/</title></head>
<body>
<h1>Index of /%s/%s/</h1><hr><pre><a href="../">../</a>
<a href="%s/">%s/</a>
<a href="%s">%s</a>
</pre><hr></body>
</html>`
	subdirBodyDotSlashF = `
<html>
<head><title>Index of %s/%s/</title></head>
<body>
<h1>Index of /%s/%s/</h1><hr><pre><a href="../">../</a>
<a href="./%s/">%s/</a>
<a href="./%s">%s</a>
</pre><hr></body>
</html>`
	subdirBodies         []string
	subdirBodiesDotSlash []string
)

func initFileHierarchy() {
	for _, v := range subdirs {
		subdirBodies = append(subdirBodies,
			fmt.Sprintf(subdirBodyF, homedir, v, homedir, v, dirname, dirname, filename, filename),
		)

		subdirBodiesDotSlash = append(subdirBodiesDotSlash,
			fmt.Sprintf(subdirBodyDotSlashF, homedir, v, homedir, v, dirname, dirname, filename, filename),
		)
	}
}

func runMockMirror(t testing.TB) *mocha.Mocha {
	initFileHierarchy()

	m := mocha.New(t).CloseOnCleanup(t)

	m.AddMocks(
		// home dir.
		mocha.Get(expect.URLPath(fmt.Sprintf("/%s/", homedir)).
			Or(expect.URLPath(fmt.Sprintf("/%s", homedir)))).
			Reply(reply.OK().BodyString(homedirBody)))

	// Sub directories.
	for i := range subdirs {
		m.AddMocks(
			// File in root.
			mocha.Get(expect.URLPath(fmt.Sprintf("/%s/%s/", homedir, subdirs[i])).
				Or(expect.URLPath(fmt.Sprintf("/%s/%s", homedir, subdirs[i])))).
				Reply(reply.OK().BodyString(subdirBodies[0])),
			// Sub directory.
			mocha.Get(expect.URLPath(fmt.Sprintf("/%s/%s/", homedir, subdirs[i])).
				Or(expect.URLPath(fmt.Sprintf("/%s/%s", homedir, subdirs[i])))).
				Reply(reply.OK().BodyString(subdirBodies[0])),
			// File in sub directory.
			mocha.Get(expect.URLPath(fmt.Sprintf("/%s/%s/%s/", homedir, subdirs[i], dirname)).
				Or(expect.URLPath(fmt.Sprintf("/%s/%s/%s", homedir, subdirs[i], dirname)))).
				Reply(reply.OK().BodyString("")),
			// Directory in sub directory.
			mocha.Get(expect.URLPath(fmt.Sprintf("/%s/%s/%s", homedir, subdirs[i], filename))).
				Reply(reply.OK().BodyString("")))
	}

	m.Start()

	return m
}

var _ = Describe("Mirror root search mock", func() {
	var (
		search *centos.MirrorRootsSearcher
		ctx    = context.Background()
	)

	Context("With versions", func() {
		BeforeEach(func() {
			search = centos.NewMirrorRootSearcher()
		})
		Context("with seed URLs", Ordered, func() {
			var (
				sourceCh = make(chan string)
				destCh   = make(chan string)
				actual   []string
				expected []string
			)
			BeforeAll(func() {

				// Test producer.
				go func() {
					seed, _ := url.JoinPath(m.URL(), homedir)
					sourceCh <- seed
					close(sourceCh)
				}()

				// Stage.
				destCh = search.Run(ctx, sourceCh)

				// Test sink.
				for v := range destCh {
					actual = append(actual, v)
				}

				// Expected data.
				expected = []string{}
				for _, v := range subdirs {
					s, _ := url.JoinPath(m.URL(), homedir, v+"/")
					expected = append(expected, s)
				}
			})
			It("Should not fail", func() {
				Expect(destCh).ToNot(BeNil())
			})
			It("Should stream results", func() {
				Expect(actual).ToNot(BeEmpty())
			})
			It("Should stream expected results", func() {
				Expect(actual).To(Equal(expected))
			})
		})
	})
})
