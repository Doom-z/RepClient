package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Doom-z/RepClient/client/model"
	"github.com/Doom-z/RepClient/pkg/logger"
)

type Client struct {
	apiURL   *url.URL
	pageSize int
	client   *http.Client
	apiKey   string
}

type Option func(*Client)

func WithPageSize(size int) Option {
	return func(c *Client) {
		c.pageSize = size
	}
}

func WithApiKey(apiKey string) Option {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.client = httpClient
	}
}

// NewClient creates and configures a new Client instance.
// rawURL is the base URL of the API server, e.g., "https://api.example.com".
func NewClient(rawURL string, opts ...Option) (*Client, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid API URL: %w", err)
	}

	c := &Client{
		apiURL:   parsed,
		pageSize: 100,
		client:   http.DefaultClient,
		apiKey:   "",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// FetchRecordsStream streams DNS records that match a specific query parameter and value.
// It supports paginated responses and continues fetching records until all available data
// has been retrieved or an error occurs.
//
// Parameters:
//   - param: The name of the query parameter to filter by (e.g., "ip", "domain").
//   - value: The value of the parameter to search for.
//
// Returns:
//   - A receive-only channel of model.Record objects streamed from the server.
//   - A receive-only channel of error, which will contain a single error if one occurs during fetching.
//
// Example:
//
//	recordsCh, errCh := client.FetchRecordsStream("ip", "1.1.1.1")
//	for record := range recordsCh {
//	    fmt.Printf("Record: %+v\n", record)
//	}
//	if err, ok := <-errCh; ok {
//	    log.Fatalf("stream error: %v", err)
//	}
//
// Notes:
//   - This method runs the fetch operation in a separate goroutine.
//   - Both channels will be closed when the operation completes or encounters an error.
//   - Errors such as HTTP failures or JSON decoding issues are sent through the error channel.
func (c *Client) FetchRecordsStream(param, value string) (<-chan model.Record, <-chan error) {
	recordsCh := make(chan model.Record, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(recordsCh)
		defer close(errCh)

		pageToken := ""

		for {
			reqURL := c.buildURL("/api/dns/paging", param, value, pageToken)

			req, err := http.NewRequest("GET", reqURL.String(), nil)
			if err != nil {
				errCh <- fmt.Errorf("request creation error: %w", err)
				return
			}
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
			resp, err := c.client.Do(req)
			if err != nil {
				errCh <- fmt.Errorf("request error: %w", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				errCh <- fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
				return
			}

			var result model.RecordsResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				errCh <- fmt.Errorf("decode error: %w", err)
				return
			}

			for _, record := range result.Data {
				recordsCh <- record
			}

			if !result.Pagination.HasMore {
				break
			}
			pageToken = result.Pagination.NextPageToken
		}
	}()

	return recordsCh, errCh
}

// FetchRecords limited fetches DNS records that match a specific query parameter and value.
func (c *Client) FetchRecords(param, value string) ([]model.Record, error) {
	query := url.Values{}
	query.Set(param, value)
	reqURL := c.apiURL.ResolveReference(&url.URL{
		Path:     "/api/dns",
		RawQuery: query.Encode(),
	})
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var result []model.Record
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	return result, nil
}

// FetchDNSRecords retrieves DNS records of a given type (e.g., "a", "aaaa") for the specified IP address.
// It uses a paginated API client to fetch the records and returns a channel of type T and an error channel.
//
// Type Parameters:
//   - T: A type that matches the expected JSON structure (e.g., model.ARecord, model.AAAARecord).
//
// Parameters:
//   - c: A pointer to a Client that handles HTTP requests.
//   - recordType: A string representing the DNS record type ("a", "aaaa", "mx", etc.).
//   - ip: The IP address to query DNS records for.
//
// Returns:
//   - A receive-only channel of type T containing the fetched DNS records.
//   - A receive-only channel of type error for handling any errors that occur during the process.
//
// Example:
//
//	recordsCh, errCh := FetchDNSRecords[model.ARecord](client, "a", "8.8.8.8")
//	for record := range recordsCh {
//	    fmt.Println(record.Domain, record.ASN)
//	}
//	if err, ok := <-errCh; ok {
//	    log.Fatalf("error: %v", err)
//	}
//
// Notes:
//   - This function automatically follows pagination until all records are retrieved.
//   - If an error occurs, the error channel will receive it and then close.
//   - Make sure type T matches the expected structure of the API response's "data" field.
func FetchDNSRecords[T any](c *Client, recordType string, ip string) (<-chan T, <-chan error) {
	recordsCh := make(chan T, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(recordsCh)
		defer close(errCh)

		pageToken := ""
		for {
			param := "ipv4"
			if recordType == "aaaa" {
				param = "ipv6"
			}
			reqURL := c.buildURL(fmt.Sprintf("/api/dns/%s", recordType), param, ip, pageToken)
			req, err := http.NewRequest("GET", reqURL.String(), nil)
			if err != nil {
				errCh <- fmt.Errorf("request creation error: %w", err)
				return
			}
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
			resp, err := c.client.Do(req)

			if err != nil {
				errCh <- fmt.Errorf("request error: %w", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				errCh <- fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
				return
			}

			var result struct {
				Data       []T                      `json:"data"`
				Pagination model.PaginationMetadata `json:"pagination"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				errCh <- fmt.Errorf("decode error: %w", err)
				return
			}

			for _, record := range result.Data {
				recordsCh <- record
			}
			if !result.Pagination.HasMore {
				logger.Infof("masuk$")
				break
			}
			pageToken = result.Pagination.NextPageToken
		}
	}()

	return recordsCh, errCh
}

// buildURL constructs the full URL with query parameters for fetching DNS records.
// - param: the query key (e.g., "ip", "domain_id")
// - value: the corresponding value to filter by
// - pageToken: the token used to fetch the next page of results
func (c *Client) buildURL(pathApi, param, value, pageToken string) *url.URL {
	query := url.Values{}
	query.Set(param, value)
	query.Set("page_size", strconv.Itoa(c.pageSize))
	if pageToken != "" {
		query.Set("page_token", pageToken)
	}

	return c.apiURL.ResolveReference(&url.URL{
		Path:     pathApi,
		RawQuery: query.Encode(),
	})
}
