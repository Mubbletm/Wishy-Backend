package ogp

import (
	"errors"
	"reflect"
	"strings"
	"wishlist-backend/utils/fetch"

	"golang.org/x/net/html"
)

// The OGP data of a page.
type ogpData struct {
	Url         string `json:"url"`
	Image       string `json:"imageUrl"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// The relevant attributes of an OGP meta HTML tag.
type ogpAttributes struct {
	Property string
	Content  string
}

// Retrieve OGP data from the given URL.
// May return an error if the URL is invalid or the body of the URL is invalid HTML.
func GetOGPData(url string) (*ogpData, error) {
	res, err := fetch.BotClient().HTTPFetch("GET", url, "")
	if err != nil {
		return nil, err
	}

	node, err := res.Parser.HTML()

	if err != nil {
		return nil, err
	}

	head := getHead(node)
	metatags := getHTMLElements(head, "meta")
	metadata := ogpData{Url: url}

	for _, metatag := range *metatags {
		var attributes ogpAttributes
		if err := bindAttributes(metatag, &attributes); err != nil {
			return nil, err
		}

		if err := bindOGPAttributes(attributes, &metadata); err != nil {
			return nil, err
		}
	}

	if len(metadata.Image) <= 0 {
		metadata.Image, _ = GetFavicon(head)
	}

	return &metadata, nil
}

// Binds the values of attribute to a single field on the provided struct.
// The `Property` field is interpreted as the key and the `Content` field will be the value.
//
// If the value of the `Property` field starts with "og:", then the substring will be ommitted.
//
// May return an error if the provided struct has a relevant field that cannot be set,
// or if a relevant field is not of type string.
func bindOGPAttributes(attribute ogpAttributes, target interface{}) error {
	targetReflect := reflect.ValueOf(target).Elem()

	for i := 0; i < targetReflect.Type().NumField(); i++ {
		fieldName := strings.Replace(attribute.Property, "og:", "", 1)

		fieldType := targetReflect.Type().Field(i)
		if strings.EqualFold(fieldType.Name, fieldName) {
			field := targetReflect.Field(i)

			if !field.CanSet() {
				return errors.New("cannot set field " + fieldType.Name)
			}
			if fieldType.Type.String() != "string" {
				return errors.New("field " + fieldType.Name + " must be of type string")
			}

			field.Set(reflect.ValueOf(attribute.Content))
		}
	}
	return nil
}

// Binds the attributes of the given HTML node to the provided struct.
//
// May return an error if the provided struct has a relevant field that cannot be set,
// or if a relevant field is not of type string.
func bindAttributes(node *html.Node, target interface{}) error {
	targetReflect := reflect.ValueOf(target).Elem()
	fieldNames := reflect.ValueOf(target).Elem().Type()

	if node.Type != html.ElementNode {
		return errors.New("node is of invalid type")
	}

	for _, attribute := range node.Attr {
		// Iterate over all available attributes of HTML Element

		for i := 0; i < fieldNames.NumField(); i++ {
			// Iterate over all available fields of struct

			fieldType := fieldNames.Field(i)

			// If struct field name matches attribute name.
			if strings.EqualFold(fieldType.Name, attribute.Key) {
				field := targetReflect.FieldByName(fieldType.Name)
				if !field.CanSet() {
					return errors.New("cannot set field " + fieldType.Name)
				}
				if fieldType.Type.String() != "string" {
					return errors.New("field " + fieldType.Name + " must be of type string")
				}

				field.Set(reflect.ValueOf(attribute.Val))
			}
		}
	}
	return nil
}

// Traverses the HTML node tree and returns the first instance of a
// head HTML element, if present, otherwise returns nil.
func getHead(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "head" {
		return node
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		headNode := getHead(child)
		if headNode != nil {
			return headNode
		}
	}
	return nil
}

// Traverses the HTML node tree and returns all found HTML meta elements.
func getHTMLElements(node *html.Node, tag string) *[]*html.Node {
	if node == nil {
		return &[]*html.Node{}
	}

	if node.Type == html.ElementNode && node.Data == tag {
		return &[]*html.Node{node}
	}

	nodes := []*html.Node{}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		metaNode := getHTMLElements(child, tag)
		if metaNode != nil {
			nodes = append(nodes, *metaNode...)
		}
	}
	return &nodes
}
