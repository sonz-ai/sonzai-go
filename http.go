package sonzai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SDKVersion is the current version of the sonzai-go SDK.
const SDKVersion = "1.4.1"

type httpClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func newHTTPClient(baseURL, apiKey string, timeout time.Duration, customClient *http.Client) *httpClient {
	var hc *http.Client
	if customClient != nil {
		hc = customClient
	} else {
		hc = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxConnsPerHost:     10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	}
	return &httpClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: hc,
	}
}

const (
	maxRetries  = 3
	baseBackoff = 500 * time.Millisecond
)

// isRetryable returns true for HTTP status codes that indicate a transient failure.
func isRetryable(statusCode int) bool {
	return statusCode == 502 || statusCode == 503 || statusCode == 504 || statusCode == 429
}

// isIdempotent returns true for HTTP methods that are safe to retry.
func isIdempotent(method string) bool {
	return method == http.MethodGet || method == http.MethodDelete
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

	var bodyData []byte
	if body != nil {
		bodyData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
	}

	attempts := 1
	if isIdempotent(method) {
		attempts = maxRetries
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter.
			backoff := baseBackoff * time.Duration(1<<uint(attempt-1))
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			delay := backoff + jitter

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			log.Printf("sonzai: retrying %s %s (attempt %d/%d)", method, path, attempt+1, attempts)
		}

		var bodyReader io.Reader
		if bodyData != nil {
			bodyReader = bytes.NewReader(bodyData)
		}

		req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", fmt.Sprintf("sonzai-go/%s", SDKVersion))

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("do request: %w", err)
			// Retry on network errors for idempotent methods.
			if isIdempotent(method) && attempt < attempts-1 {
				continue
			}
			return nil, lastErr
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
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

			var retryAfter *int
			if resp.StatusCode == 429 {
				if raHeader := resp.Header.Get("Retry-After"); raHeader != "" {
					if ra, parseErr := strconv.Atoi(raHeader); parseErr == nil {
						retryAfter = &ra
					}
				}
			}
			lastErr = newErrorForStatus(resp.StatusCode, msg, retryAfter)

			// Retry on transient failures for idempotent methods.
			if isIdempotent(method) && isRetryable(resp.StatusCode) && attempt < attempts-1 {
				continue
			}
			return nil, lastErr
		}

		return respBody, nil
	}

	return nil, lastErr
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

// DeleteWithParams performs an HTTP DELETE request with query parameters.
func (c *httpClient) DeleteWithParams(ctx context.Context, path string, params map[string]string, result interface{}) error {
	data, err := c.request(ctx, http.MethodDelete, path, nil, params)
	if err != nil {
		return err
	}
	if result != nil {
		return json.Unmarshal(data, result)
	}
	return nil
}

// PostMultipartFile uploads a file (provided as bytes) as a multipart form POST and unmarshals the response into result.
func (c *httpClient) PostMultipartFile(ctx context.Context, path, fieldName, fileName string, fileData []byte, result interface{}) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return fmt.Errorf("write file data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", fmt.Sprintf("sonzai-go/%s", SDKVersion))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := string(respBody)
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			msg = errResp.Error
		}
		return newErrorForStatus(resp.StatusCode, msg, nil)
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

// PostMultipart sends a multipart/form-data POST request using an io.Reader for the file body,
// plus arbitrary string fields.
func (c *httpClient) PostMultipart(ctx context.Context, path string, fields map[string]string, fileName string, fileContent io.Reader, contentType string, result interface{}) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, val := range fields {
		if err := writer.WriteField(key, val); err != nil {
			return fmt.Errorf("write field %s: %w", key, err)
		}
	}

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, fileContent); err != nil {
		return fmt.Errorf("copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", fmt.Sprintf("sonzai-go/%s", SDKVersion))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := string(respBody)
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			msg = errResp.Error
		}
		return newErrorForStatus(resp.StatusCode, msg, nil)
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
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
	req.Header.Set("User-Agent", fmt.Sprintf("sonzai-go/%s", SDKVersion))

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
		return newErrorForStatus(resp.StatusCode, msg, nil)
	}

	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer size to handle large SSE responses (default is 64KB, set to 1MB max)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
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

// UploadFile sends a multipart/form-data POST request and unmarshals the
// JSON response into result. Pass nil for result to discard the response body.
func (c *httpClient) UploadFile(ctx context.Context, path string, fileName string, fileData []byte, contentType string, result interface{}) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return fmt.Errorf("write file data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &buf)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", fmt.Sprintf("sonzai-go/%s", SDKVersion))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := string(respBody)
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			msg = errResp.Error
		}
		return newErrorForStatus(resp.StatusCode, msg, nil)
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
	}
	return nil
}
