package sitemap

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/temoto/robotstxt"
)

// Crawl fetches the root URL and then crawls it with the given parallelism. It will
// return the sitemap and the number of warnings encountered during the crawl, such as
// 404s or parse errors.
func Crawl(logger hclog.Logger, rootURL string, parallelism int) (*Sitemap, int, error) {
	return crawl(logger, webFetcher, rootURL, parallelism)
}

// crawl implements the public Crawl function but allows the fetcher to be injected
// for testing.
func crawl(logger hclog.Logger, fetcher fetchFn, rootURL string, parallelism int) (*Sitemap, int, error) {
	if parallelism < 1 {
		return nil, 0, fmt.Errorf("Parallelism must be > 0 (got %d)", parallelism)
	}

	robots, err := getRobots(fetcher, rootURL)
	if err != nil {
		return nil, 0, fmt.Errorf("Could not set up robots.txt filter: %v", err)
	}

	urls := make(urlChan, parallelism*10)
	results := make(resultChan, parallelism)

	var wg sync.WaitGroup
	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func() {
			for url := range urls {
				links, err := processPage(fetcher, rootURL, robots, url)
				results <- &pageResult{
					url:   url,
					links: links,
					err:   err,
				}
			}
			wg.Done()
		}()
	}

	sm := NewSitemap()
	expectedResults := 1
	backlog := []string{rootURL}
	warnings := 0
	for expectedResults > 0 {
		// Stuff the work queue with as much as we can fit in there, but we can't
		// block here otherwise we won't be able to process results.
	BACKLOG:
		for len(backlog) > 0 {
			nextURL := backlog[0]
			select {
			case urls <- nextURL:
				backlog = backlog[1:]
			default:
				break BACKLOG
			}
		}

		// Block for the next result, since we know there's outstanding work.
		result := <-results
		expectedResults--
		if result.err != nil {
			warnings++
			logger.Warn("Crawling problem", "url", result.url, "error", result.err)
			continue
		}

		newURLs := sm.AddLinks(result.url, result.links)
		for _, newURL := range newURLs {
			expectedResults++
			backlog = append(backlog, newURL)
		}
		logger.Debug("Processed page", "url", result.url,
			"unique_links", len(result.links),
			"new_links", len(newURLs))
	}

	close(urls)
	wg.Wait()
	return sm, warnings, nil
}

// urlChan holds the backlog of pages to be processed.
type urlChan chan string

// pageResult comes back from processing a page and has the raw list of URLs linked
// from the page, as well as any processing errors.
type pageResult struct {
	url   string
	links []string
	err   error
}

// resultChan holds the incoming page processing results.
type resultChan chan *pageResult

// getRobots fetches the robots.txt file from the site, if it exists, and
// returns a filter group we can use to test if we are allowed to crawl a
// given URL. If there's no robots.txt present this won't throw an error,
// it will return an all-pass filter.
func getRobots(fetcher fetchFn, rootURL string) (*robotstxt.Group, error) {
	base, err := url.Parse(rootURL)
	if err != nil {
		return nil, err
	}

	robots, err := url.Parse("/robots.txt")
	if err != nil {
		return nil, err
	}

	robotsURL := base.ResolveReference(robots)
	status, body, err := fetcher(robotsURL.String())
	if err != nil {
		return nil, err
	}
	defer body.Close()

	contents, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	r, err := robotstxt.FromStatusAndBytes(status, contents)
	if err != nil {
		return nil, err
	}

	// We don't know our user agent so we just match the empty string to get
	// the default group of rules.
	return r.FindGroup(""), nil
}

// processPage fetches pages and returns the raw list of URLs linked from the page.
func processPage(fetcher fetchFn, rootURL string, robots *robotstxt.Group, url string) ([]string, error) {
	status, body, err := fetcher(url)
	if status != 200 {
		return nil, fmt.Errorf("Bad HTTP status %d", status)
	}
	defer body.Close()

	rawLinks, err := scanPageForLinks(body)
	if err != nil {
		return nil, err
	}

	links, err := filterLinks(rootURL, robots, rawLinks)
	if err != nil {
		return nil, err
	}
	return links, nil
}

// scanPageForLinks reads the HTML body in the reader and returns all the links
// it finds in anchors.
func scanPageForLinks(r io.Reader) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var links []string
	doc.Find("a[href]").Each(func(i int, item *goquery.Selection) {
		link, _ := item.Attr("href")
		links = append(links, link)
	})
	return links, nil
}

// filterLinks returns a clean set of links that follow from a page. It removes
// duplicates, gets rid of external links, and makes sure links comply with the
// robots.txt policy for the site.
func filterLinks(rootURL string, robots *robotstxt.Group, links []string) ([]string, error) {
	base, err := url.Parse(rootURL)
	if err != nil {
		return nil, err
	}
	basePrefix := base.String()

	dedup := make(map[string]struct{})
	followLinks := make([]string, 0, len(links))
	for _, link := range links {
		u, err := url.Parse(link)
		if err != nil {
			return nil, err
		}

		u = base.ResolveReference(u)
		link := u.String()

		// Fragments aren't sent to the web server, so they shouldn't be treated like separate
		// pages for crawling.
		if len(u.Fragment) > 0 {
			link = strings.TrimSuffix(link, fmt.Sprintf("#%s", u.Fragment))
		}

		if _, ok := dedup[link]; ok {
			continue
		}
		dedup[link] = struct{}{}

		if !strings.HasPrefix(link, basePrefix) {
			continue
		}

		if allow := robots.Test(u.Path); !allow {
			continue
		}

		followLinks = append(followLinks, link)
	}
	return followLinks, nil
}
