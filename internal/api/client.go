package api

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
)

// Client is an API client.
type Client struct {
	BaseURL    *url.URL     // The API base URL.
	HTTPClient *http.Client // An HTTP client which handles authentication.
}

// NewClient instantiates an API client using the passed absolute base URL.
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if parsedBaseURL.Host == "" {
		return nil, fmt.Errorf("invalid base URL (missing host): %s", baseURL)
	}
	return &Client{
		BaseURL:    parsedBaseURL,
		HTTPClient: httpClient,
	}, nil
}

// Error is an error returned from an API call, allowing access to the response.
type Error struct {
	Original error
	URL      string
	Response *http.Response
}

func (e Error) Error() string {
	var msg string
	switch {
	case e.Original != nil:
		msg = fmt.Sprintf("API error: %s: %s", e.Original.Error(), e.URL)
	case e.Response != nil:
		msg = fmt.Sprintf("API error: %s: %s", e.Response.Status, e.URL)
	default:
		msg = fmt.Sprintf("API error: %s", e.URL)
	}
	if e.Response != nil && e.Response.StatusCode != http.StatusNotFound {
		defer e.Response.Body.Close()
		d, _ := httputil.DumpResponse(e.Response, false)
		msg += "\n\nFull response:\n\n" + strings.TrimSpace(string(d)) + "\n"
	}
	return msg
}

// resolveURL adds path segments to the client's configured base URL, escaping each segment.
func (c *Client) baseURLWithSegments(segments ...string) (*url.URL, error) {
	var relativeURL string
	for _, p := range segments {
		relativeURL = path.Join(relativeURL, url.PathEscape(p))
	}
	return c.resolveURL(relativeURL)
}

// resolveURL resolves a relative URL according to the client's configured base URL.
func (c *Client) resolveURL(relativeURL string) (*url.URL, error) {
	parsedRef, err := url.Parse(relativeURL)
	if err != nil {
		return nil, err
	}
	return c.BaseURL.ResolveReference(parsedRef), nil
}
