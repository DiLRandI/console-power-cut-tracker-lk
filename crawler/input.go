package crawler

import (
	"errors"

	"golang.org/x/net/html"
)

func GetFirstInput(doc *html.Node) (*html.Node, error) {
	var input *html.Node
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "input" {
			input = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)
	if input != nil {
		return input, nil
	}

	return nil, errors.New("missing <input> in the node tree")
}
