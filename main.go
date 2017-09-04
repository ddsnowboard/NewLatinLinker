package main

import "fmt"
import "net/url"
import "net/http"
import "golang.org/x/net/html"
import "golang.org/x/net/html/atom"
import "sync"

const Address = "http://www.thelatinlibrary.com/"

type outFile struct {
	location string
	content  string
}

func main() {
	fmt.Println("This is a test")
}

func processSomething(url url.URL, ret chan outFile, wg *sync.WaitGroup) {
	defer wg.Done()
	// Figure out what it is and process it
	if url.IsAbs() {
		fmt.Println("You can't pass in an absolute URL; passed in %s", url.String())
		return
	}
	response, err := http.Get(url.String())
	if err != nil {
		fmt.Println("You passed in a bad URL: %s", url.String())
		return
	}
	nodes, err := html.ParseFragment(response.Body, nil)

	if err != nil {
		fmt.Println("Something bad happened with parsing %s", url.String())
		return
	}

	numTd := countTags(nodes, "td")
	numP := countTags(nodes, "p")
	// If there are more td's than paragraphs, it's probably a list page
	if numTd > numP {
		wg.Add(1)
		go processList(nodes, ret, wg)
	} else {
		wg.Add(1)
		go processWork(nodes, ret, wg)
	}
}

func processList(nodes []*html.Node, ret chan outFile, wg *sync.WaitGroup) {

}

func processWork(nodes []*html.Node, ret chan outFile, wg *sync.WaitGroup) {

}

func hasClass(n *html.Node, className string) bool {
	const Class = "class"
	for _, attr := range n.Attr {
		if attr.Key == Class && attr.Val == className {
			return true
		}
	}
	return false
}

func countTags(nodes []*html.Node, name string) int {
	nameBytes := make([]byte, len(name))
	copy(nameBytes, name)
	tagHex := atom.Lookup(nameBytes)
	out := 0
	for _, tag := range nodes {
		if tag.DataAtom == tagHex {
			out += 1
		}
	}
	return out
}
