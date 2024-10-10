package fetch

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"maps"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"wishlist-backend/utils"

	"golang.org/x/net/html"
)

type FetchClient struct {
	headers map[string]string
	cookies *cookiejar.Jar
}

type FetchResponse struct {
	Response *http.Response
	// A parser for the response body. Can convert the body stream to other formats.
	Parser *HTTPBodyParser
}

type HTTPBodyParser struct {
	response *http.Response
}

var botClient *FetchClient

// Returns a FetchClient singleton instance that is configured to advertise itself
// as a bot. More specifically; a google webscraper bot.
func BotClient() *FetchClient {
	if botClient != nil {
		return botClient
	}

	jar, err := cookiejar.New(nil)

	if err != nil {
		panic("something went wrong while initializing the cookie jar")
	}

	client := &FetchClient{
		headers: make(map[string]string),
		cookies: jar,
	}

	client.AddHeader("User-Agent", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/W.X.Y.Z Safari/537.36")
	botClient = client
	return botClient
}

// Adds a cookie to the client. The cookie will only be applied to the request if the
// URL matches.
func (client *FetchClient) AddCookie(urlString string, name string, content string, maxAge int) error {
	urlObj, err := url.Parse(urlString)

	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     name,
		Value:    content,
		Quoted:   false,
		Path:     urlObj.Path,
		Domain:   urlObj.Host,
		MaxAge:   maxAge,
		Secure:   false,
		HttpOnly: urlObj.Scheme == "http",
	}

	client.cookies.SetCookies(urlObj, append(client.cookies.Cookies(urlObj), cookie))
	return nil
}

// Removes a cookie from the client
func (client *FetchClient) RemoveCookie(urlString string, name string) error {
	urlObj, err := url.Parse(urlString)

	if err != nil {
		return err
	}

	cookies := client.cookies.Cookies(urlObj)

	for i, cookie := range cookies {
		if cookie.Name == name {
			cookies[i] = cookies[len(client.cookies.Cookies(urlObj))-1]
			break
		}
	}

	newCookies := cookies[:len(client.cookies.Cookies(urlObj))-1]

	client.cookies.SetCookies(urlObj, newCookies)
	return nil
}

// Removes a header from the client.
func (client *FetchClient) RemoveHeader(header string) {
	delete(client.headers, header)
}

// Adds a default header to the client. This header will be applied to all requests
// until removed.
func (client *FetchClient) AddHeader(header string, content string) {
	client.headers[header] = content
}

// Makes a request to the given URL with the provided parameters.
// Body can be given an empty string if preferred empty.
//
// Requests may have additional cookies and headers as provided by the client.
func (client *FetchClient) HTTPFetch(method string, url string, body string) (*FetchResponse, error) {
	m := make(map[string]string)

	if len(body) > 0 {
		if utils.IsJSON(body) {
			m["Content-Type"] = "application/json"
		} else if utils.IsXML(body) {
			m["Content-Type"] = "application/xml"
		} else {
			m["Content-Type"] = "text/plain"
		}
	}

	return HTTPFetchWithHeaders(method, url, body, m)
}

// Makes a request to the given URL with the provided parameters.
// Body can be given an empty string if preferred empty.
//
// Requests may have additional cookies and headers as provided by the client.
//
// The headers provided to this function take precedent over the headers set in the client.
func (client *FetchClient) HTTPFetchWithHeaders(method string, urlString string, body string, headers map[string]string) (*FetchResponse, error) {
	req, err := http.NewRequest(method, urlString, bytes.NewReader([]byte(body)))

	if err != nil {
		return nil, err
	}

	urlObj, err := url.Parse(urlString)

	if err != nil {
		return nil, err
	}

	for _, cookie := range client.cookies.Cookies(urlObj) {
		req.AddCookie(cookie)
	}

	for key, value := range maps.All(client.headers) {
		req.Header.Set(key, value)
	}

	for key, value := range maps.All(headers) {
		req.Header.Set(key, value)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	return &FetchResponse{
		Response: res,
		Parser: &HTTPBodyParser{
			response: res,
		},
	}, nil
}

// Makes a request to the given URL with the provided parameters.
// Body can be given an empty string if preferred empty.
func HTTPFetch(method string, url string, body string) (*FetchResponse, error) {
	m := make(map[string]string)

	if len(body) > 0 {
		if utils.IsJSON(body) {
			m["Content-Type"] = "application/json"
		} else if utils.IsXML(body) {
			m["Content-Type"] = "application/xml"
		} else {
			m["Content-Type"] = "text/plain"
		}
	}

	return HTTPFetchWithHeaders(method, url, body, m)
}

// Makes a request to the given URL with the provided parameters.
// Body can be given an empty string if preferred empty.
func HTTPFetchWithHeaders(method string, url string, body string, headers map[string]string) (*FetchResponse, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(body)))

	if err != nil {
		return nil, err
	}

	for key, value := range maps.All(headers) {
		req.Header.Set(key, value)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	return &FetchResponse{
		Response: res,
		Parser: &HTTPBodyParser{
			response: res,
		},
	}, nil
}

// Parses the body to text.
func (res *HTTPBodyParser) Text() (string, error) {
	defer res.response.Body.Close()
	body, err := io.ReadAll(res.response.Body)

	if err != nil {
		return "", errors.New("something went wrong while parsing the body")
	}

	return string(body), nil
}

// Binds the JSON body to the given struct.
//
// May throw and error if the body is not valid JSON. Or if something goes
// wrong while reading the stream.
func (res *HTTPBodyParser) Json(target interface{}) error {
	defer res.response.Body.Close()
	return json.NewDecoder(res.response.Body).Decode(target)
}

// Binds the XML body to the given struct.
//
// May throw and error if the body is not valid XML. Or if something goes
// wrong while reading the stream.
func (res *HTTPBodyParser) XML(target interface{}) error {
	defer res.response.Body.Close()
	return xml.NewDecoder(res.response.Body).Decode(target)
}

// Parses the body and returns the HTML node at the root of the document.
func (res *HTTPBodyParser) HTML() (*html.Node, error) {
	defer res.response.Body.Close()
	return html.Parse(res.response.Body)
}
