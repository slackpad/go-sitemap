package sitemap

import (
	"fmt"
	"io"
	"sort"
)

// Sitemap is a data structure that keeps track of pages and the links between them.
type Sitemap struct {
	outbounds map[string]urlSet
}

// urlSet is a set of URLs.
type urlSet map[string]struct{}

// NewSitemap constructs a new sitemap.
func NewSitemap() *Sitemap {
	return &Sitemap{
		outbounds: make(map[string]urlSet),
	}
}

// AddLinks records a new page into the sitemap and returns a list of URLs that we've
// never encountered before.
func (sm *Sitemap) AddLinks(fromURL string, toURLs []string) []string {
	if _, ok := sm.outbounds[fromURL]; !ok {
		sm.outbounds[fromURL] = make(urlSet)
	}

	var newURLs []string
	for _, toURL := range toURLs {
		sm.outbounds[fromURL][toURL] = struct{}{}
		if _, ok := sm.outbounds[toURL]; !ok {
			sm.outbounds[toURL] = make(urlSet)
			newURLs = append(newURLs, toURL)
		}
	}
	return newURLs
}

// Write sends the sitemap out in text format to the given writer.
func (sm *Sitemap) Write(w io.Writer) error {
	urls := make([]string, 0, len(sm.outbounds))
	for url := range sm.outbounds {
		urls = append(urls, url)
	}
	sort.Strings(urls)

	for _, url := range urls {
		if _, err := fmt.Fprintf(w, "%s\n", url); err != nil {
			return err
		}
		outbound := sm.outbounds[url]
		links := make([]string, 0, len(outbound))
		for link := range outbound {
			links = append(links, link)
		}
		sort.Strings(links)
		for _, link := range links {
			if _, err := fmt.Fprintf(w, " -> %s\n", link); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "\n"); err != nil {
			return err
		}
	}
	return nil
}
