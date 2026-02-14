package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultBaseURL = "https://search.maven.org/solrsearch/select"

type Client struct {
	baseURL    string
	httpClient *http.Client
	cache      *Cache
}

type Option func(*Client)

func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

func WithCache(cache *Cache) Option {
	return func(c *Client) { c.cache = cache }
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Search(query string, rows, start int, bypassCache bool) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("rows", fmt.Sprintf("%d", rows))
	params.Set("start", fmt.Sprintf("%d", start))
	params.Set("wt", "json")
	params.Set("fl", "id,g,a,v,latestVersion,p,timestamp,versionCount")

	return c.doRequest(params, bypassCache)
}

func (c *Client) SearchMultimodal(query string, rows, start int, bypassCache bool) (*SearchResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return &SearchResponse{}, nil
	}

	var queries []string
	if strings.Contains(query, ":") || (strings.Contains(query, ".") && !strings.Contains(query, " ")) {
		// It looks like a GAV or Group search, use the smart builder
		queries = []string{BuildQuery(query)}
	} else {
		// It's a keyword search, use multimodal strategy
		queries = []string{
			fmt.Sprintf(`a:"%s"`, query), // Exact Artifact
			fmt.Sprintf(`g:"%s"`, query), // Exact Group
			query,                        // General keyword
		}
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []Doc
		total   int
		errs    []error
	)

	wg.Add(len(queries))
	for _, q := range queries {
		go func(q string) {
			defer wg.Done()
			resp, err := c.Search(q, rows, start, bypassCache)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			results = append(results, resp.Response.Docs...)
			if resp.Response.NumFound > total {
				total = resp.Response.NumFound
			}
		}(q)
	}
	wg.Wait()

	if len(results) == 0 && len(errs) > 0 {
		return nil, errs[0]
	}

	return &SearchResponse{
		Response: ResponseBody{
			NumFound: total,
			Docs:     results,
		},
	}, nil
}

func (c *Client) Versions(groupID, artifactID string, rows int, bypassCache bool) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("q", fmt.Sprintf(`g:"%s" AND a:"%s"`, groupID, artifactID))
	params.Set("rows", fmt.Sprintf("%d", rows))
	params.Set("core", "gav")
	params.Set("sort", "timestamp desc")
	params.Set("wt", "json")
	params.Set("fl", "id,g,a,v,latestVersion,p,timestamp,versionCount")

	return c.doRequest(params, bypassCache)
}

func (c *Client) doRequest(params url.Values, bypassCache bool) (*SearchResponse, error) {
	reqURL := c.baseURL + "?" + params.Encode()

	// Check cache
	if c.cache != nil && !bypassCache {
		if val, ok := c.cache.Get(reqURL, 24*time.Hour); ok {
			return val, nil
		}
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("User-Agent", "mvns/1.0 (https://github.com/maher90-90/mvns)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	// Save to cache
	if c.cache != nil {
		c.cache.Set(reqURL, &result)
	}

	return &result, nil
}
