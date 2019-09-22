package sitemap

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests that the sitemap properly handles new links and duplicates.
func TestSitemap_AddLinks(t *testing.T) {
	sm := NewSitemap()

	// We always assume the fromUrl isn't new, even if we encounter a link to
	// it.
	{
		newURLs := sm.AddLinks("https://wwww.google.com", []string{
			"https://wwww.google.com",
		})
		var expected []string
		require.Equal(t, expected, newURLs)
	}

	// Map some more pages from this one.
	{
		newURLs := sm.AddLinks("https://wwww.google.com", []string{
			"https://wwww.google.com",
			"https://wwww.google.com/about",
			"https://www.google.com/privacy",
			"https://www.google.com/contact",
		})
		expected := []string{
			"https://wwww.google.com/about",
			"https://www.google.com/privacy",
			"https://www.google.com/contact",
		}
		require.Equal(t, expected, newURLs)
	}

	// Map one of these pages that refers to a known one.
	{
		newURLs := sm.AddLinks("https://wwww.google.com/about", []string{
			"https://www.google.com/humans",
			"https://www.google.com/contact",
		})
		expected := []string{
			"https://www.google.com/humans",
		}
		require.Equal(t, expected, newURLs)
	}
}

// Tests the sitemap writer.
func TestSitemap_Write(t *testing.T) {
	sm := NewSitemap()
	sm.AddLinks("https://www.slackpad.com/xyz", []string{
		"https://www.slackpad.com/bbb",
		"https://www.slackpad.com/aaa",
	})
	sm.AddLinks("https://www.slackpad.com", []string{
		"https://www.slackpad.com/xyz",
		"https://www.slackpad.com/abc",
	})
	sm.AddLinks("https://www.slackpad.com/aaa", []string{
		"https://www.slackpad.com",
	})
	var b bytes.Buffer
	require.NoError(t, sm.Write(&b))
	expected := "" +
		"https://www.slackpad.com\n" +
		" -> https://www.slackpad.com/abc\n" +
		" -> https://www.slackpad.com/xyz\n" +
		"\n" +
		"https://www.slackpad.com/aaa\n" +
		" -> https://www.slackpad.com\n" +
		"\n" +
		"https://www.slackpad.com/abc\n" +
		"\n" +
		"https://www.slackpad.com/bbb\n" +
		"\n" +
		"https://www.slackpad.com/xyz\n" +
		" -> https://www.slackpad.com/aaa\n" +
		" -> https://www.slackpad.com/bbb\n" +
		"\n"
	require.Equal(t, expected, b.String())
}
