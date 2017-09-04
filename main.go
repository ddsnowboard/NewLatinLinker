package main

import "fmt"
import "bytes"
import "strings"
import "net/url"
import "net/http"
import "golang.org/x/net/html"
import "golang.org/x/net/html/atom"
import "sync"
import "errors"

const Address = "www.thelatinlibrary.com/"

type outFile struct {
	location string
	content  string
}

func main() {
	c := make(chan outFile, 1000)
	wg := sync.WaitGroup{}
	url, err := url.Parse("http://" + Address)
	if err != nil {
		fmt.Println("Something bad happened with parsing the url")
		return
	}
	wg.Add(1)
	processSomething(url, c, &wg)
	aFile := <-c
	fmt.Println("File content was ", aFile.content)
	wg.Wait()
}

func processSomething(url *url.URL, ret chan outFile, wg *sync.WaitGroup) {
	defer wg.Done()
	// Figure out what it is and process it
	if !url.IsAbs() {
		url.Host = Address
		url.Scheme = "http"
	}
	response, err := http.Get(url.String())
	if err != nil {
		fmt.Printf("You passed in a bad URL: %s", url.String())
		return
	}
	headNode, err := html.Parse(response.Body)
	fmt.Printf("Parsed html for %s\n", url.String())

	if err != nil {
		fmt.Printf("Something bad happened with parsing %s", url.String())
		return
	}

	numTd := CountTags(headNode, "td")
	numP := CountTags(headNode, "p")
	// If there are more td's than paragraphs, it's probably a list page
	if numTd > numP {
		wg.Add(1)
		go processList(url.RequestURI(), headNode, ret, wg)
	} else {
		wg.Add(1)
		go processWork(url.RequestURI(), headNode, ret, wg)
	}
}

func processList(path string, headNode *html.Node, ret chan outFile, wg *sync.WaitGroup) {
	defer wg.Done()
	out := outFile{location: path}
	tdAtom := getAtom("td")
	nodes := GetAllChildNodes(headNode)
	for _, n := range nodes {
		if n.DataAtom == tdAtom {
			anchor := n.FirstChild
			href, err := getHref(anchor)
			if err != nil {
				fmt.Printf("We couldn't find the href on node %s on %s", n.Data, path)
				continue
			}
			url, err := url.Parse(href)
			if err != nil {
				fmt.Printf("We couldn't parse the url %s", href)
				continue
			}
			wg.Add(1)
			go processSomething(url, ret, wg)
		}
	}
	buffer := new(bytes.Buffer)
	err := html.Render(buffer, getHTMLTag(nodes[0]))
	if err != nil {
		fmt.Printf("We couldn't render the nodes for %s. God help us", path)
		return
	}
	out.content = buffer.String()
	ret <- out
}

func processWork(path string, headNode *html.Node, ret chan outFile, wg *sync.WaitGroup) {
	defer wg.Done()
	const whitakersWords = "http://www.archives.nd.edu/cgi-bin/wordz.pl?keyword=%s"
	const minLength = 10
	out := outFile{location: path}
	nodes := GetAllChildNodes(headNode)
            if nodes == nil {
                fmt.Println("Something bad happened")
            }
	for _, n := range nodes {
		if n.Type == html.TextNode {
			text := n.Data
			// We have to make sure that this is big enough to actually have real text
			// in it. It might not.
			if len(text) > minLength {
				n.Data = ""
				for _, word := range strings.Split(text, " ") {
					textNode := html.Node{}
					attributes := [...]html.Attribute{html.Attribute{Key: "href",
						Val: fmt.Sprintf(whitakersWords, word)}}
					newNode := html.Node{FirstChild: &textNode,
						LastChild: &textNode,
						Type:      html.ElementNode,
						DataAtom:  atom.A,
						Data:      "a",
						Attr:      attributes[:]}
					textNode.Parent = &newNode
					textNode.Data = word

				}
			}
		}
	}

	buffer := new(bytes.Buffer)
	err := html.Render(buffer, getHTMLTag(headNode))
	if err != nil {
		fmt.Printf("We couldn't render the nodes for %s. God help us", path)
		return
	}
	out.content = buffer.String()
	ret <- out
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

func CountTags(headNode *html.Node, name string) int {
	out := 0
	countTags(headNode, name, &out)
	return out
}

func countTags(headNode *html.Node, name string, count *int) {
	if headNode == nil {
		return
	}
	tagHex := getAtom(name)
	switch headNode.Type {
	case html.DocumentNode:
		countTags(headNode.FirstChild, name, count)
		return
	case html.ElementNode:
		for walker := headNode; walker != nil; walker = walker.NextSibling {
			if walker.DataAtom == tagHex {
				*count += 1
			}
			countTags(walker.FirstChild, name, count)
		}
	default:
		return
	}
}

func getHTMLTag(node *html.Node) *html.Node {
	walker := node
	for walker.Type != html.DocumentNode {
		walker = walker.Parent
	}
	return walker
}

func getAtom(tagName string) atom.Atom {
	nameBytes := make([]byte, len(tagName))
	copy(nameBytes, tagName)
	tagHex := atom.Lookup(nameBytes)
	return tagHex
}

func getHref(n *html.Node) (string, error) {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			return attr.Val, nil
		}
	}
	return "", errors.New("There was no href on this element")
}

func GetAllChildNodes(headNode *html.Node) []*html.Node {
	out := make([]*html.Node, 5)
	getAllChildNodes(headNode, &out)
	return out
}

func getAllChildNodes(headNode *html.Node, l *[]*html.Node) {
	if headNode == nil {
		return
	}
	for walker := headNode; walker != nil; walker = walker.NextSibling {
		*l = append(*l, walker)
		getAllChildNodes(walker.FirstChild, l)
	}
}
