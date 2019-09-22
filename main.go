package main

import (
	"flag"
	"fmt"
	"os"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/slackpad/go-sitemap/sitemap"
)

func crawl(logger hclog.Logger, rootURL string, parallelism int, filename string, failWithWarnings bool) error {
	logger.Info(fmt.Sprintf("Starting crawl of %s", rootURL))
	sm, warnings, err := sitemap.Crawl(logger, rootURL, parallelism)
	if err != nil {
		return err
	}
	if warnings > 0 {
		if failWithWarnings {
			return fmt.Errorf("Crawl had %d warnings, not writing sitemap", warnings)
		}
		logger.Warn("Warnings occurred during crawl, see logs", "count", warnings)
	}

	logger.Info(fmt.Sprintf("Writing output to %s", filename))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	if err := sm.Write(f); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	logger.Info("Crawl complete")
	return nil
}

func main() {
	var url = flag.String("url", "", "URL to crawl")
	var filename = flag.String("filename", "sitemap.txt", "Output filename")
	var parallelism = flag.Int("parallelism", 10, "Number of simultaneous requests")
	var logLevel = flag.String("log-level", "INFO", "Log level (DEBUG, INFO, or ERROR)")
	var failWithWarnings = flag.Bool("fail-with-warnings", false, "Fail for any crawler warnings")
	flag.Parse()

	if *url == "" {
		fmt.Fprintf(os.Stderr, "URL to crawl is required (use -url <URL>)\n")
		os.Exit(2)
	}

	level := hclog.LevelFromString(*logLevel)
	if level == hclog.NoLevel {
		fmt.Fprintf(os.Stderr, "Invlid log level %q\n", *logLevel)
		os.Exit(2)
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "go-sitemap",
		Level: level,
	})
	if err := crawl(logger, *url, *parallelism, *filename, *failWithWarnings); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
