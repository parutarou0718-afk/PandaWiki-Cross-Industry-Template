package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL = "http://localhost:8080/api/v1"
	defaultTimeout = 30 * time.Second
)

// Client 是所有API的统一客户端
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
}

type ClientOption func(*Client)

// New 创建一个新的API客户端
func New(apiBase string, apiKey string, opts ...ClientOption) *Client {
	baseURL, _ := url.Parse(apiBase)
	c := &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithHTTPClient 自定义http.Client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// newRequest 构造http请求
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	u := c.baseURL.JoinPath(path)
	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-API-Version", "1.0.0")
	req.Header.Set("X-App-Name", "Panda-Wiki")
	return req, nil
}

// do 发送请求并解析响应
func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查业务code
	var common CommonResponse
	_ = json.Unmarshal(body, &common)
	if common.Code != 0 {
		return fmt.Errorf("业务错误 code=%d, message=%s", common.Code, common.Message)
	}

	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

// parseErrorResponse 解析错误响应
func parseErrorResponse(resp *http.Response) error {
	var errResp CommonResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("failed to decode error response: %w", err)
	}
	return errors.New(errResp.Message)
}
