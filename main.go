package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64: x64) AppleWebKit/537.36 (KHTML, like Gecko/61.0.31)"}

func checkRelative(href string, baseURL string) string {
	if strings.HasPrefix(href, "/") {
		return fmt.Sprintf("%s%s", baseURL, href)
	} else {
		return href
	}
}

func resolveRelativeLinks(href string, baseURL string) (bool, string) {
	resultHref := checkRelative(href, baseURL)
	baseParse, _ := url.Parse(baseURL)

	resultParse, _ := url.Parse(resultHref)

	if baseParse != nil && resultParse != nil {
		if baseParse.Host == resultParse.Host {
			return true, resultHref
		} else {
			return false, ""
		}
	}

	return false, ""
}

func randomUserAgent() string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Int() % len(userAgents)
	return userAgents[randNum]
}

func discoverLinks(response *http.Response, baseURL string) []string {
	if response != nil {
		doc, _ := goquery.NewDocumentFromResponse(response)

		foundUrls := []string{}

		if doc != nil {
			doc.Find("a").Each(func(i int, s *goquery.Selection) {
				res, _ := s.Attr("href")
				foundUrls = append(foundUrls, res)
			})
		}
		return foundUrls
	}
	return []string{}
}

func getRequest(targetURL string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", targetURL, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", randomUserAgent())

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	} else {
		return res, nil
	}

}

var tokens = make(chan struct{}, 5)

func Crawl(targetURL string, baseURL string) []string {
	fmt.Println(targetURL)

	tokens <- struct{}{}
	res, err := getRequest(targetURL)

	if err != nil {
		panic(err)
	}

	links := discoverLinks(res, baseURL)
	foundUrls := []string{}

	for _, link := range links {

		ok, correctLink := resolveRelativeLinks(link, baseURL)

		if ok {
			if correctLink != nil {
				foundUrls = append(foundUrls, correctLink)
			}
		}

	}
	//ParseHTML(resp)
}

func main() {
	worklist := make(chan []string)
	baseDomain := "https://www.theguardian.com"
	var n int
	n++

	go func() {
		worklist <- []string{"https:www.theguardian.com"}

	}()

	seen := make(map[string]bool)

	for ; n > 0; n-- {

		list := worklist

		for _, link := range list {
			if !seen[link] {
				seen[link] = true
				n++ //tracks new links
				go func(link string, baseUrl string) {
					foundLinks := Crawl(link, baseDomain)

					if foundLinks != nil {
						worklist <- foundLinks
					}
				}()
			}
		}
	}

}
