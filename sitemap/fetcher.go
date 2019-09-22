package sitemap

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

// fetchFn is a lightweight type we can use to plug different fetchers into the
// crawler.
type fetchFn func(targetURL string) (int, io.ReadCloser, error)

// webFetcher pulls pages over the Internet.
func webFetcher(targetURL string) (int, io.ReadCloser, error) {
	resp, err := http.Get(targetURL)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, resp.Body, nil
}

// diskFetcher pulls pages from the local filesystem using the "fixtures" path
// as the root, with a folder inside for the hostname. When fetching the root
// URL, this will look for a file called "index.html" in the folder.
type diskFetcher struct {
	fixtures string
}

// fetch pulls a fake page from the fixtures.
func (df *diskFetcher) fetch(targetURL string) (int, io.ReadCloser, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return 0, nil, err
	}

	if u.Path == "" {
		u.Path = "index.html"
	}
	diskPath := path.Join(df.fixtures, u.Host, u.Path)

	r, err := os.Open(diskPath)
	if os.IsNotExist(err) {
		body := ioutil.NopCloser(strings.NewReader(""))
		return 404, body, nil
	}
	if err != nil {
		return 0, nil, err
	}
	return 200, r, nil
}
