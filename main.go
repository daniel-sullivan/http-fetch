package main

import (
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	getMetadata := flag.Bool("metadata", false, "Display metadata regarding image count, etc.")
	flag.Parse()

	fmt.Println("Retrieving the following URLs:")
	urlList := flag.Args()
	fmt.Println(urlList)

	// Ensure there's an output folder
	_ = os.Mkdir("output", os.ModePerm)

	for _, urlString := range urlList {
		uri, err := url.Parse(urlString)
		filename := "output/" + uri.Host + uri.Path + ".html"

		if err != nil {
			fmt.Printf("Invalid URL: %s. Skipping...\n", urlString)
			continue
		}

		pageData, err := RetrieveResource(urlString, filename)
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println("----------------------------------------------------")
		fmt.Printf("%s to %s\n", urlString, filename)

		// If metadata has been requested, print the data to the console
		if getMetadata != nil && *getMetadata == true {
			fmt.Printf("site: %s\n", uri.Host)
			fmt.Printf("images: %d\n", len(pageData.Images))
			fmt.Printf("javascript: %d\n", len(pageData.Javascripts))
			fmt.Printf("stylesheet: %d\n", len(pageData.Stylesheets))
			fmt.Printf("num_links: %d\n", len(pageData.Links))
			fmt.Printf("last_fetch: %s\n", time.Now().String())
		}
	}
}

// getExtension takes a filename and pulls the final period-delimited
// component if one exists (eg: "xxx.yyy.html" -> "html")
func getExtension(filename string) (string, bool) {
	fnSplit := strings.Split(filename, ".")
	if len(fnSplit) > 1 {
		return fnSplit[len(fnSplit) - 1], true
	} else {
		return "", false
	}
}

type PageData struct {
	Links []string
	Images []string
	Stylesheets []string
	Javascripts []string
}

// RetrieveResource takes a string format URL and a filename to output to if successful.
// On success, if the retrieved file was an HTML file, the PageData struct will contain
// metadata about the page.
func RetrieveResource(urlString, filename string) (PageData, error) {
	var pageData PageData

	response, err := http.Get(urlString)
	if err != nil {
		return pageData, fmt.Errorf("Error getting URL: %s. (%s) Skipping...\n", urlString, err.Error())

	}
	defer response.Body.Close()

	f, err := os.Create(filename)
	if err != nil {
		return pageData, fmt.Errorf("Unable to create file: %s (%s)\n", filename, err.Error())
	}
	defer f.Close()

	// If the request is for HTML, parse for links, etc.
	if extension, exists := getExtension(filename); exists && extension == "html" {
		if response.StatusCode != 200 {
			return pageData, fmt.Errorf("Status code error: %d %s", response.StatusCode, response.Status)
		}

		// Create a folder for any assets
		_ = os.Mkdir(filename + "-res", os.ModePerm)

		// Load the HTML document
		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Find the links
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			if link, exists := s.Attr("href"); exists {
				pageData.Links = append(pageData.Links, link)
			}
		})

		baseURI, _ := url.Parse(urlString) // This error may be ignored as we've already run this func. previously

		// Find images, stylesheets, and JS
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			if imgSrc, exists := s.Attr("src"); exists {
				// Create a UUID identifier for the resource
				imgFn := fmt.Sprintf("%s-res/%s", filename, uuid.New().String())
				if extension, exists := getExtension(imgSrc); exists {
					imgFn = fmt.Sprintf("%s.%s", imgFn, extension)
				}

				pageData.Images = append(pageData.Images, imgFn)

				imgURI, err := baseURI.Parse(imgSrc)

				if err != nil {
					return
				}

				_, err = RetrieveResource(imgURI.String(), imgFn)

				if err != nil {
					return
				} else {
					s.SetAttr("src", imgFn)
				}
			}
		})
		doc.Find("link").Each(func(i int, s *goquery.Selection) {
			if linkType, exists := s.Attr("rel"); exists && linkType == "stylesheet" {
				if ssSource, exists := s.Attr("href"); exists {
					// Create a UUID identifier for the resource
					ssFn := fmt.Sprintf("%s-res/%s", filename, uuid.New().String())
					if extension, exists := getExtension(ssSource); exists {
						ssFn = fmt.Sprintf("%s.%s", ssFn, extension)
					}

					pageData.Stylesheets = append(pageData.Stylesheets, ssFn)

					ssURI, err := baseURI.Parse(ssSource)

					if err != nil {
						return
					}

					_, err = RetrieveResource(ssURI.String(), ssFn)

					if err != nil {
						return
					} else {
						s.SetAttr("href", ssFn)
					}
				}
			}
		})
		doc.Find("script").Each(func(i int, s *goquery.Selection) {
			if jsSrc, exists := s.Attr("src"); exists {
				// Create a UUID identifier for the resource
				jsFn := fmt.Sprintf("%s-res/%s", filename, uuid.New().String())
				if extension, exists := getExtension(jsSrc); exists {
					jsFn = fmt.Sprintf("%s.%s", jsFn, extension)
				}

				pageData.Javascripts = append(pageData.Javascripts, jsFn)

				jsURI, err := baseURI.Parse(jsSrc)

				if err != nil {
					return
				}

				_, err = RetrieveResource(jsURI.String(), jsFn)

				if err != nil {
					return
				} else {
					s.SetAttr("src", jsFn)
				}
			}
		})

		html, err := doc.Html()
		if err != nil {
			return pageData, fmt.Errorf("Unable to rebuild page with error: %s", err.Error())
		}

		_, err = io.Copy(f, strings.NewReader(html))

		if err != nil {
			return pageData, fmt.Errorf("Unable to save URL: %s to %s. (%s)\n", urlString, filename, err.Error())
		}
	} else {
		_, err = io.Copy(f, response.Body)

		if err != nil {
			return pageData, fmt.Errorf("Unable to save URL: %s to %s. (%s)\n", urlString, filename, err.Error())
		}
	}

	return pageData, nil
}