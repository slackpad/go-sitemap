package sitemap

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// fetchAsString is a test helper that fetches a page and returns the contents
// as a string (or fails). This works with any kind of fetcher.
func fetchAsString(t *testing.T, fetcher fetchFn, targetURL string) string {
	status, body, err := fetcher(targetURL)
	require.NoError(t, err)
	defer body.Close()
	require.True(t, status == 200 || status == 404)

	b, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	return bytes.NewBuffer(b).String()
}

// Tests the web fetcher against a local test web server.
func TestFetcher_webFetcher(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	status, body, err := webFetcher(ts.URL)
	require.Nil(t, err)
	require.Equal(t, status, 200)

	greeting, err := ioutil.ReadAll(body)
	body.Close()
	require.Nil(t, err)
	require.Equal(t, "Hello, client\n", string(greeting))
}

// Tests for basic sanity with the disk fetcher.
func TestFetcher_Disk(t *testing.T) {
	fetcher := (&diskFetcher{"../fixtures"}).fetch

	fakeSites := []struct {
		url      string
		expected string
	}{
		{"https://www.404s.com/nope.txt", ""},
		{"https://www.404s.com/hello.txt", "Hello, world!"},
	}
	for _, site := range fakeSites {
		t.Run(site.url, func(t *testing.T) {
			actual := fetchAsString(t, fetcher, site.url)
			require.Equal(t, site.expected, actual)
		})
	}
}
