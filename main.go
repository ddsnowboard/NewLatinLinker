package main

import "runtime/debug"
import "regexp"
import "os/exec"
import "path"
import "strings"
import "io/ioutil"
import "fmt"
import "bytes"
import "net/url"
import "net/http"
import "golang.org/x/net/html"
import "golang.org/x/net/html/atom"
import "sync"
import "errors"

const Address = "www.thelatinlibrary.com"

type outFile struct {
	Location string
	Content  string
}

type mutexSet struct {
	Set   map[string]bool
	Mutex sync.Mutex
}

type synchronyStuff struct {
	Set mutexSet
	Wg  sync.WaitGroup
}

func main() {
	c := make(chan outFile, 1000)
	waiter := make(chan bool)
	synchro := synchronyStuff{mutexSet{make(map[string]bool), sync.Mutex{}}, sync.WaitGroup{}}
	// url, err := url.Parse("http://" + Address)
	url, err := url.Parse("http://www.thelatinlibrary.com/lucan/lucan1.shtml")
	if err != nil {
		fmt.Println("Something bad happened with parsing the url")
		return
	}
	synchro.Wg.Add(1)
	go processSomething(url, c, &synchro)
	go waitForStuff(waiter, &synchro.Wg)
	for {
		select {
		case out := <-c:
			writeFile(&out)
		case <-waiter:
			fmt.Println("Everything's done!")
			return
		}
	}
}

func processSomething(url *url.URL, ret chan outFile, synchro *synchronyStuff) {
	defer synchro.Wg.Done()
	// Figure out what it is and process it
	if !url.IsAbs() {
		url.Host = Address
		url.Scheme = "http"
	}
	response, err := http.Get(url.String())
	if err != nil {
		fmt.Printf("You passed in a bad URL: %s\n", url.String())
		return
	}
	headNode, err := html.Parse(response.Body)

	if err != nil {
		fmt.Printf("Something bad happened with parsing %s\n", url.String())
		return
	}

	numTd := CountTags(headNode, "td")
	numP := CountTags(headNode, "p")
	// If there are more td's than paragraphs, it's probably a list page
	if numTd > numP {
		synchro.Wg.Add(1)
		go processList(url.RequestURI(), headNode, ret, synchro)
	} else {
		synchro.Wg.Add(1)
		go processWork(url.RequestURI(), headNode, ret, synchro)
	}
}

func processList(path string, headNode *html.Node, ret chan outFile, synchro *synchronyStuff) {
	defer synchro.Wg.Done()
	if checkPath(path, &synchro.Set) {
		return
	}
	out := outFile{Location: path}
	tdAtom := getAtom("td")
	nodes := GetAllChildNodes(headNode)
	for _, n := range nodes {
		// Sometimes they don't have the text and there are no child nodes to find
		if n.DataAtom == tdAtom && n.FirstChild != nil {
			anchor := n.FirstChild
			href, err := getHref(anchor)
			if err != nil {
				fmt.Printf("We couldn't find the href on node %s on %s\n", n.Data, path)
				continue
			}
			url, err := url.Parse(href)
			if err != nil {
				fmt.Printf("We couldn't parse the url %s\n", href)
				continue
			}
			synchro.Wg.Add(1)
			go processSomething(url, ret, synchro)
		}
	}
	buffer := new(bytes.Buffer)
	err := html.Render(buffer, getHTMLTag(nodes[0]))
	if err != nil {
		fmt.Printf("We couldn't render the nodes for %s. God help us\n", path)
		return
	}
	out.Content = buffer.String()
	ret <- out
}

func processWork(path string, headNode *html.Node, ret chan outFile, synchro *synchronyStuff) {
	defer synchro.Wg.Done()
	if checkPath(path, &synchro.Set) {
		return
	}
	const whitakersWords = "http://www.archives.nd.edu/cgi-bin/wordz.pl?keyword=%s"
	const minLength = 10
	splitters := regexp.MustCompile("[!-/:-@[-`{-~ \n]")
        whitespace := regexp.MustCompile("^\\s+$")
	out := outFile{Location: path}
	nodes := GetAllChildNodes(headNode)
	for _, n := range nodes {
		fmt.Printf("Type is %s and data is %s\n", n.Type, n.Data)
	}

	for _, n := range nodes {
		if n.Type == html.TextNode && n.Parent.DataAtom == getAtom("p") {
			text := n.Data
			// We have to make sure that this is big enough to actually have real text
			// in it. It might not.
			if len(text) > minLength {
				n.Data = ""
				for _, word := range splitters.Split(text, -1) {
					if whitespace.MatchString(word) {
						continue
					}
					textNode := html.Node{Type: html.TextNode,
						Data: word + " "}
					attributes := [...]html.Attribute{html.Attribute{Key: "href",
						Val: fmt.Sprintf(whitakersWords, word)}}
					newNode := html.Node{
						FirstChild: &textNode,
						LastChild:  &textNode,
						Type:       html.ElementNode,
						DataAtom:   atom.A,
						Data:       "a",
						Attr:       attributes[:]}
					n.Parent.AppendChild(&newNode)
				}
			}
		}
	}

	buffer := new(bytes.Buffer)
        htmlTag := getHTMLTag(headNode)
        fmt.Println("tag is type", htmlTag.Type, "with data", htmlTag.Data)
	err := html.Render(buffer, htmlTag)
	if err != nil {
		fmt.Printf("We couldn't render the nodes for %s. God help us\n", path)
		fmt.Println("Error was", err)
		return
	}
	out.Content = buffer.String()
        fmt.Println(out.Content)
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
	case html.ElementNode:
		// Sometimes we end up in the middle somehow. Don't ask me
		for walker := headNode; walker != nil; walker = walker.PrevSibling {
		}
		for walker := headNode; walker != nil; walker = walker.NextSibling {
			if walker.DataAtom == tagHex {
				*count += 1
			}
			countTags(walker.FirstChild, name, count)
		}
	default:
		// Who knows what this is
		if headNode.FirstChild != nil {
			countTags(headNode.FirstChild, name, count)
		}
		if headNode.NextSibling != nil {
			countTags(headNode.NextSibling, name, count)
		}
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
	if n == nil {
		debug.PrintStack()
	}
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			return attr.Val, nil
		}
	}
	return "", errors.New("There was no href on this element")
}

func GetAllChildNodes(headNode *html.Node) []*html.Node {
	out := make([]*html.Node, 0)
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

func waitForStuff(c chan bool, wg *sync.WaitGroup) {
	wg.Wait()
	c <- true
}

// Returns whether the path was already done
func checkPath(path string, syncSet *mutexSet) bool {
	syncSet.Mutex.Lock()
	defer syncSet.Mutex.Unlock()
	_, present := syncSet.Set[path]
	if present {
		return true
	} else {
		syncSet.Set[path] = true
		return false
	}
}

func writeFile(f *outFile) {
	const folder = "output/"
	re := regexp.MustCompile(`href\s*=\s*"(.*?)\.shtml"`)
	location := folder + strings.Replace(f.Location, "shtml", "html", -1)
	dir := path.Dir(location)
	command := exec.Command("mkdir", "-p", dir)
	command.Run()
	b := []byte(re.ReplaceAllString(f.Content, `href="$1.html"`))
	err := ioutil.WriteFile(location, b, 0644)
	if err != nil {
		fmt.Println("Something bad happened writing", location)
		fmt.Println(err)
	}
}
