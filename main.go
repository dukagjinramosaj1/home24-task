// find_links_in_page.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// This will get called for each HTML element found
func main() {
	// Create HTTP client with timeout
	var resultingLinks int

	// var URL string = "https://www.golangcode.com"

	urlPrompt := os.Args[1:]
	url := strings.Join(urlPrompt, "")
	// fmt.Println(URL2)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create and modify HTTP request before sending
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Make request
	response, err := client.Do(request)
	if response != nil {
		defer response.Body.Close()
	}

	getWebTitle(url)

	//GET ONLY UNIQUE LINKS
	linksInDocument, err := getUniqueLinksFromResponse(response)
	if err != nil {
		log.Fatal(err)
	}
	//Loop through links to Count how many
	for key, _ := range linksInDocument {
		//- if we want to show which links we use the value url
		// fmt.Println(" - ", key, url)
		resultingLinks = key
	}
	fmt.Println("- Total number of internal and external links available:", resultingLinks)

	getHeadingCount(url)

	//I used only once this approach of parsing in order to get the html NODE of the DOCUMENT TYPE
	doc, err := html.Parse(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	//FUTURE WORK better ways to call the functions and refactor the code into directorys
	getHTMLVersion(doc)

	hasLogin(url)

	getBlogTitle(url)

}

//Extracted only UNIQUE links not duplicates
func getUniqueLinksFromResponse(response *http.Response) ([]string, error) {

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	// Get all a hrefs
	var links []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			cleanLinks, found := getLinksFromHref(href)
			if found {
				links = appendIfNotExists(links, cleanLinks)
			}
		}
	})
	return links, nil
}

//Only extract certain types of Links
func getLinksFromHref(href string) (link string, found bool) {
	// If it is too short it can't be a link

	if len(href) < 3 {
		return "", false
	}

	// If it starts right off with a slash, it is a relative URL and no domain
	if href[0] == '/' {
		return "", false
	}

	// Strip # and everything after
	pos := strings.Index(href, "#")
	if pos > -1 {
		href = href[:pos]
	}

	pos = strings.Index(href, "mailto:")
	if pos > -1 {
		href = href[:pos]
	}

	if !validateLink(href) {
		return "", false
	}
	// if no good domain name obtained, return false
	return strings.ToLower(href), true
}

//Extra Checks
func validateLink(link string) bool {
	if len(link) < 3 {
		return false
	}

	splitLink := strings.Split(link, ".")
	if len(splitLink) < 2 { // There is not two parts to the link
		return false
	}

	if len(splitLink[0]) < 1 {
		return false
	}
	if len(splitLink[1]) < 2 {
		return false
	}

	return true
}

//Checks if already exits
func appendIfNotExists(strings []string, newString string) []string {
	exists := false
	for _, existingString := range strings {
		if existingString == newString {
			exists = true
		}
	}
	if !exists {
		strings = append(strings, newString)
	}
	return strings
}

//Blog Titles if there are any
func getBlogTitle(url string) {

	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic("Something is WRONG")
	}
	// use CSS selector found with the browser inspector
	// for each, use index and item
	doc.Find(".post-title").Each(func(index int, item *goquery.Selection) {
		title := item.Text()
		// linkTag := item.Find("a")
		// link, _ := linkTag.Attr("href")
		fmt.Println("POST TITLE:", index, title)
	})
}

//Website Title
func getWebTitle(url string) {
	var pageTitle string
	var description string

	//Get the Web
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	// use CSS selector found with the browser inspector
	// for each, use index and item
	// usually there is only one Title tag in HTML files for Tittle
	pageTitle = doc.Find("title").Contents().Text()

	// Now get meta description field
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("name"); strings.EqualFold(name, "description") {
			description, _ = s.Attr("content")
			// fmt.Println("THE WEBSITE DESCRIPTION IS: ", description)
		}
	})

	fmt.Printf("- Page Title: '%s'\n", pageTitle)
	fmt.Printf("- Page Description: '%s'\n", description)

}

//Counting Headings Levels on occurance
func getHeadingCount(url string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic("Something is WRONG")
	}

	//Checking for all Headings if they exist
	headingsInput := []string{"h1", "h2", "h3", "h4", "h5", "h6"}

	for _, v := range headingsInput {
		doc.Find("body").Each(func(i int, s *goquery.Selection) {
			heading := s.Find(v).Text()
			headingCounts(heading, v)
		})
	}
}

//Helper function for getting the ccountings in Headers
func headingCounts(heading string, headingType string) map[string]int {

	headings := strings.Fields(heading)
	wc := make(map[string]int)
	for _, word := range headings {
		_, matched := wc[word]
		if matched {
			wc[word]++
		} else {
			wc[word] = 1
		}
	}
	fmt.Printf("\tHEADING TYPE: %v\t %v\n ", headingType, len(wc))
	return wc

}

//Since catching doctype is not currently avaiable through goquery selection
//different approach was taken
func getHTMLVersion(doc *html.Node) {

	var htmlVersion string
	//Older HTML versions include a declaration of HTML version which must refer to a DTD (Document Type Definition)
	//therefore we would be able to catch it if its not the default <!Doctype html> which is HTML5
	//definition by default
	if len(doc.FirstChild.Attr) != 0 {
		htmlVersion = doc.FirstChild.Attr[0].Val
	} else {
		htmlVersion = "- WEBSITE VERSION: HTML5  " //Due to default <!Doctype html>
	}

	fmt.Println(htmlVersion)
}

func hasLogin(url string) {

	var loginType string

	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic("Something is WRONG")

	}

	//NEEDS better Solution
	//created two types of potential Login id or class on an potential HTML
	inputs := []string{".login", "./login", ".log in", ".log_in", ".signup", ".sign_up", ".signin", ".sign_in", ".auth", "#login", "#log_in", "#signup", "#sign_up", "#signin", "#sign_in", "#auth"}

	// use CSS selector found with the browser inspector
	// for each, use index and item
	for _, v := range inputs {
		// fmt.Println("THIS IS V", v)
		doc.Find("body").Each(func(index int, item *goquery.Selection) {
			loginType = item.Find(v).Text()
		})
	}
	if loginType != "" {
		fmt.Println("- This page has a login form with id name: ", loginType)
	} else {
		fmt.Println("- This page does not have a login form")
	}

}
