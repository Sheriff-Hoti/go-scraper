package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func parseArgs(args []string) (string, error) {

	guide := "you must specify `--url or -u flag` followed by the site url"

	if len(args) > 3 {
		return "", fmt.Errorf("too many arguents, %v", guide)
	}

	if len(args) < 3 {
		return "", fmt.Errorf("not enough arguents, %v", guide)
	}

	if args[1] == "--url" || args[1] == "-u" {
		return "", fmt.Errorf("invalid argument, %v", guide)
	}

	return args[2], nil
}

func traverse(n *html.Node, operation func(n *html.Node)) {

	operation(n)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverse(c, operation)
	}
}

func visitUrl(url string) ([]string, error) {

	log.Printf("visiting %v\n", url)
	links := make([]string, 0, 10)

	result, http_err := http.Get(url)

	if http_err != nil {
		log.Fatal(http_err)
		return nil, http_err
	}

	defer result.Body.Close()

	if result.StatusCode >= 400 {
		return nil, fmt.Errorf("the status code was %v", result.StatusCode)
	}

	doc, html_err := html.Parse(result.Body)

	if html_err != nil {
		log.Fatal(html_err)
		result.Body.Close()
		return nil, html_err
	}

	traverse(doc, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, val := range n.Attr {
				if val.Key == "href" {
					links = append(links, val.Val)
				}
			}
		}
	})

	return links, nil

}

func main() {

	args := os.Args

	site, err := parseArgs(args)

	if err != nil {
		log.Fatal(err)
		return
	}

	url_site, url_site_err := url.Parse(site)

	if url_site_err != nil {
		log.Fatal(url_site_err)
		return
	}

	visited := make(map[string]int)

	to_be_checked := make([]string, 0, 100)

	to_be_checked = append(to_be_checked, site)

	i := 0

	log.Println("started scraping...")

	for i < len(to_be_checked) {

		//validate if the url to be checked is the same origin if not error
		//also check if the url has been checked before

		//get the element
		current, parse_err := url.Parse(to_be_checked[i])

		log.Println(current)

		i++

		if parse_err != nil {
			log.Printf("Malformed url: %v", parse_err)
			continue
		}

		//then increment i

		if current.Host != url_site.Host {
			log.Printf("the url: %v was not the same as origin", current.String())
			continue
		}

		_, ok := visited[current.String()]

		if ok {
			log.Printf("this url has been checked: %v", current.String())
			continue
		}

		scraped_links, scraped_err := visitUrl(current.String())

		visited[current.String()] = 200

		if scraped_err != nil {
			log.Printf("%s", scraped_err.Error())
			continue
		}

		for _, val := range scraped_links {

			if strings.HasPrefix(val, "http") {
				log.Printf("the url: %v was not the same as origin", val)
				continue
			}

			tested := url_site.JoinPath(val)

			_, ok := visited[tested.String()]

			if !ok {
				to_be_checked = append(to_be_checked, tested.String())
				// visited[tested.String()] = 200

			}

		}

	}

}
