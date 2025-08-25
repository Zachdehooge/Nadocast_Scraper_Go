package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func main() {
	now := time.Now()
	month := fmt.Sprintf("%02d", int(now.Month()))
	day := fmt.Sprintf("%02d", now.Day())
	year := fmt.Sprintf("20%02d", now.Year()%100)
	hour := now.Hour()
	var timeNow int
	switch {
	case 0 <= hour && hour <= 13:
		timeNow = 0
	case 14 <= hour && hour <= 18:
		timeNow = 12
	case 19 <= hour && hour <= 23:
		timeNow = 18
	default:
		fmt.Printf("(-) No Nadocast Data To Fetch // Check If Statement Set For Constraints Covering Current Time: %d\n", hour)
		return
	}

	urlStr := fmt.Sprintf("http://data.nadocast.com/%s%s/%s%s%s/t%dz/", year, month, year, month, day, timeNow)
	folderLocation := fmt.Sprintf("Nadocast/%s_%s_%s_%dz", day, month, year, timeNow)
	os.MkdirAll(folderLocation, 0755)

	resp, err := http.Get(urlStr)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.HasSuffix(attr.Val, ".png") {
					linkURL, err := url.Parse(attr.Val)
					if err != nil {
						fmt.Println("Error parsing href:", err)
						continue
					}
					fileURL := baseURL.ResolveReference(linkURL).String()
					filename := filepath.Join(folderLocation, filepath.Base(attr.Val))
					fmt.Println("Downloading:", fileURL)
					out, err := os.Create(filename)
					if err != nil {
						fmt.Println("Error creating file:", err)
						return
					}
					resp, err := http.Get(fileURL)
					if err != nil {
						fmt.Println("Error downloading file:", err)
						out.Close()
						return
					}
					_, err = io.Copy(out, resp.Body)
					out.Close()
					resp.Body.Close()
					if err != nil {
						fmt.Println("Error saving file:", err)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}
