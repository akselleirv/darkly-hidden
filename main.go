package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const readme = "README"

func main() {
	ip := flag.String("ip", "", "the IP address of the Darkly instance")
	flag.Parse()

	if *ip == "" {
		fmt.Println("IP address cannot be empty. Add it with: ./main -ip <ip-address>")
		os.Exit(1)
	}
	if net.ParseIP(*ip) == nil {
		fmt.Printf("IP address '%s' is an invalid IP\n", *ip)
		os.Exit(1)
	}

	log.Println("starting scrape on IP => ", *ip)
	start := time.Now()

	matches, readmeFilesFound := scrape([]string{fmt.Sprintf("http://%s/.hidden/", *ip)})

	fmt.Printf("[ DONE ] found %d readme files: interesting match(es) => %s\n", readmeFilesFound, matches)
	fmt.Println("Time spent: ", time.Since(start))
}

func scrape(urls []string) ([]string, int) {
	var (
		matches          []string
		readmeFilesFound int
		fn               func(u []string)
	)

	fn = func(u []string) {
		for _, url := range u {
			resp, err := doRequest(url)
			if err != nil {
				log.Fatal(err)
			}
			if readmeText := getReadmeFile(url, resp); readmeText != "" {
				readmeFilesFound++
				if containsPossibleMd5String(readmeText) {
					matches = append(matches, strings.TrimSpace(readmeText))
				}
			}
			fn(removeUnwantedHrefs(url, resp))
		}
	}

	fn(urls)
	return matches, readmeFilesFound
}

func getReadmeFile(url string, hrefs []string) string {
	for _, href := range hrefs {
		if href == readme {
			resp, err := http.Get(url + href)
			if err != nil {
				log.Fatal(err)
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			return string(b)
		}
	}
	return ""
}

func containsPossibleMd5String(s string) bool {
	if match, err := regexp.MatchString(`[a-f0-9]{32}`, s); err == nil {
		return match
	} else {
		log.Fatalf("bad regex pattern: %s", err.Error())
		return false
	}
}

// removeUnwantedHrefs removes README and parent folder href
func removeUnwantedHrefs(currentUrl string, hrefs []string) []string {
	var result []string
	for _, href := range hrefs {
		if href != readme && href != "../" {
			result = append(result, currentUrl+href)
		}
	}
	return result
}

func doRequest(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to get '%s': %w", url, err)
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Fatalf("unable to close resp body: %s", err)
		}
	}(resp.Body)

	return traversForHrefs(resp.Body)
}

// traversForHrefs finds all the hrefs in the HTML response
func traversForHrefs(data io.Reader) ([]string, error) {
	var result []string
	doc, err := html.Parse(data)
	if err != nil {
		return nil, err
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					result = append(result, a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	return result, nil
}
