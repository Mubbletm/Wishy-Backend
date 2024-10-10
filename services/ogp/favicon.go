package ogp

import (
	"errors"
	"slices"
	"sort"
	"strconv"
	"strings"
	arrays "wishlist-backend/utils/array"

	"golang.org/x/net/html"
)

type FaviconAttributes struct {
	Rel   string
	Type  string
	Sizes string
	Href  string
}

// Traverse the node tree to try to get the highest quality favicon.
//
// Prioritizes apple and mask-icons. After that it will start looking for the favicon with the biggest size attribute.
//
// May return low quality results if the given document does not make use of relevant attributes.
//
// May return an error if no favicon is present.
func GetFavicon(node *html.Node) (string, error) {
	links := getHTMLElements(getHead(node), "link")
	faviconData := []FaviconAttributes{}

	for _, link := range *links {
		var faviconAttributes FaviconAttributes
		if err := bindAttributes(link, &faviconAttributes); err != nil {
			return "", err
		}
		faviconData = append(faviconData, faviconAttributes)
	}

	filteredData := arrays.Filter(faviconData, func(fa FaviconAttributes) bool {
		for _, relation := range []string{"icon", "apple-touch-icon", "apple-touch-icon-precomposed", "mask-icon"} {
			if len(fa.Type) > 0 && !strings.Contains(fa.Type, "png") {
				return false
			}

			if strings.Contains(fa.Rel, relation) {
				return true
			}
		}
		return false
	})

	if len(filteredData) <= 0 {
		return "", errors.New("page has no favicon")
	}

	slices.SortFunc(filteredData, func(a, b FaviconAttributes) int {
		if !strings.EqualFold(a.Rel, b.Rel) {
			if strings.EqualFold(a.Rel, "mask-icon") {
				return -1
			}

			if strings.EqualFold(b.Rel, "mask-icon") {
				return 1
			}

			if strings.Contains(a.Rel, "apple") {
				return -1
			}

			if strings.Contains(b.Rel, "apple") {
				return 1
			}
		}

		aSizes := strings.Split(strings.Trim(strings.ReplaceAll(a.Sizes, "x", ""), " "), " ")
		bSizes := strings.Split(strings.Trim(strings.ReplaceAll(a.Sizes, "x", ""), " "), " ")

		stringToInt := func(s string) int {
			if len(s) <= 0 {
				return -1
			} else {
				i, err := strconv.Atoi(s)
				if err != nil {
					return -1
				}
				return i
			}
		}

		convertedA := arrays.Map(aSizes, stringToInt)
		convertedB := arrays.Map(bSizes, stringToInt)

		sort.Sort(sort.Reverse(sort.IntSlice(convertedA)))
		sort.Sort(sort.Reverse(sort.IntSlice(convertedB)))

		if len(convertedA) <= 0 || len(convertedB) <= 0 {
			if len(convertedA) > 0 {
				return -1
			}
			if len(convertedB) > 0 {
				return 1
			}
			return 0
		}
		return convertedB[0] - convertedA[0]
	})

	return filteredData[0].Href, nil
}
