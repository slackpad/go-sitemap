package sitemap

import (
	"bytes"
	"path"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

// Tests the full crawler against fake web sites stored in the fixtures folder.
func TestCrawler_Integration(t *testing.T) {
	logger := hclog.NewNullLogger()
	fetcher := (&diskFetcher{"../fixtures"}).fetch

	fakeSites := []string{
		"https://www.withoutrobots.com", // full site is open
		"https://www.withrobots.com",    // execs.html has forbidden links
		"https://www.filters.com",       // exercises filterLinks
	}
	for _, site := range fakeSites {
		t.Run(site, func(t *testing.T) {
			sm, warnings, err := crawl(logger, fetcher, site, 3)
			require.Zero(t, warnings)
			require.NoError(t, err)

			expected := fetchAsString(t, fetcher, path.Join(site, "sitemap.txt"))
			var b bytes.Buffer
			require.NoError(t, sm.Write(&b))
			require.Equal(t, expected, b.String())
		})
	}
}

// Tests that errors and warnings are returned from the crawl.
func TestCrawler_Errors(t *testing.T) {
	logger := hclog.NewNullLogger()
	fetcher := (&diskFetcher{"../fixtures"}).fetch

	t.Run("bad parallelism", func(t *testing.T) {
		_, warnings, err := crawl(logger, fetcher, "https://www.404s.com", 0)
		require.Zero(t, warnings)
		require.Contains(t, err.Error(), "Parallelism must be > 0 (got 0)")
	})

	t.Run("missing page", func(t *testing.T) {
		_, warnings, err := crawl(logger, fetcher, "https://www.404s.com", 3)
		require.Equal(t, warnings, 1)
		require.Nil(t, err)
	})
}
