# Simple web crawler in Go

Given a starting URL, this will crawl that subdomain and output a simple textual sitemap.

```
$ ./go-sitemap -url https://www.slackpad.com
2019-09-22T00:58:53.227-0700 [INFO]  go-sitemap: Starting crawl of https://www.slackpad.com
2019-09-22T00:58:53.662-0700 [INFO]  go-sitemap: Writing output to sitemap.txt
2019-09-22T00:58:53.663-0700 [INFO]  go-sitemap: Crawl complete

$ cat sitemap.txt
https://www.slackpad.com
 -> https://www.slackpad.com
 -> https://www.slackpad.com/
 -> https://www.slackpad.com/about/
 -> https://www.slackpad.com/feed.xml
 -> https://www.slackpad.com/startups/hashtagtodo/programming/2015/08/14/seriously-a-todo-list.html
 -> https://www.slackpad.com/travel/california/offroad/iheart395/2012/07/13/buttermilk-country.html

https://www.slackpad.com/about/
 -> https://www.slackpad.com
 -> https://www.slackpad.com/
 -> https://www.slackpad.com/about/
 -> https://www.slackpad.com/startups/hashtagtodo/programming/2015/08/14/seriously-a-todo-list.html
```

## Usage

```
Usage of ./go-sitemap:
  -fail-with-warnings
        Fail for any crawler warnings
  -filename string
        Output filename (default "sitemap.txt")
  -log-level string
        Log level (DEBUG, INFO, or ERROR) (default "INFO")
  -parallelism int
        Number of simultaneous requests (default 10)
  -url string
        URL to crawl
```
