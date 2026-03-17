package sonzai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type httpClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func newHTTPClient(baseURL, apiKey string, timeout time.Duration) *httpClient {
	return &httpClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *httpClient) request(ctx context.Context, method, path string, body interface{}, params map[string]string) ([]byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if len(params) > 0 {
		q := u.Query()
		for k, v := range params {
			if v != "" {
				q.Set(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "sonzai-go/0.1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := string(respBody)
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			msg = errResp.Error
		}
		return nil, newErrorForStatus(resp.StatusCode, msg)
	}

	return respBody, nil
}

// Get performs an HTTP GET request and unmarshals the response into result.
func (c *httpClient) Get(ctx context.Context, path string, params map[string]string, result interface{}) error {
	data, err := c.request(ctx, http.MethodGet, path, nil, params)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

// Post performs an HTTP POST request and unmarshals the response into result.
func (c *httpClient) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	data, err := c.request(ctx, http.MethodPost, path, body, nil)
	if err != nil {
		return err
	}
	if result != nil {
		return json.Unmarshal(data, result)
	}
	return nil
}

// Put performs an HTTP PUT request and unmarshals the response into result.
func (c *httpClient) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	data, err := c.request(ctx, http.MethodPut, path, body, nil)
	if err != nil {
		return err
	}
	if result != nil {
		return json.Unmarshal(data, result)
	}
	return nil
}

// Patch performs an HTTP PATCH request and unmarshals the response into result.
func (c *httpClient) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	data, err := c.request(ctx, http.MethodPatch, path, body, nil)
	if err != nil {
		return err
	}
	if result != nil {
		return json.Unmarshal(data, result)
	}
	return nil
}

// Delete performs an HTTP DELETE request and unmarshals the response into result.
func (c *httpClient) Delete(ctx context.Context, path string, result interface{}) error {
	data, err := c.request(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	if result != nil {
		return json.Unmarshal(data, result)
	}
	return nil
}

// StreamSSE sends a request and calls the callback for each parsed SSE event.
func (c *httpClient) StreamSSE(ctx context.Context, method, path string, body interface{}, callback func(json.RawMessage) error) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", "sonzai-go/0.1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		msg := string(respBody)
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			msg = errResp.Error
		}
		return newErrorForStatus(resp.StatusCode, msg)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "data: [DONE]" {
			return nil
		}
		if strings.HasPrefix(line, "data: ") {
			data := line[6:]
			if err := callback(json.RawMessage(data)); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}
