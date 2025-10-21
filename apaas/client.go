package apaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultBaseURL = "https://ae-openapi.feishu.cn"

// ClientOptions configures the API client instance.
type ClientOptions struct {
	Namespace         string
	ClientID          string
	ClientSecret      string
	DisableTokenCache bool
	BaseURL           string
	HTTPClient        *http.Client
	Logger            Logger
	LimiterOptions    *LimiterOptions
	RetryConfig       *RetryConfig
}

// Client wraps HTTP access to the aPaaS OpenAPI.
type Client struct {
	clientID          string
	clientSecret      string
	namespace         string
	disableTokenCache bool

	httpClient *http.Client
	baseURL    *url.URL

	logger Logger

	tokenMu         sync.RWMutex
	accessToken     string
	expireTime      time.Time
	tokenRefreshing bool // Flag to prevent concurrent token refreshes

	limiter *RateLimiter

	retryConfig RetryConfig

	// Service groups
	Object     *ObjectService
	Department *DepartmentService
	Function   *FunctionService
	Page       *PageService
	Attachment *AttachmentService
	Global     *GlobalService
	Automation *AutomationService
}

// NewClient constructs a client using the provided options.
func NewClient(opts ClientOptions) (*Client, error) {
	if strings.TrimSpace(opts.Namespace) == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	if strings.TrimSpace(opts.ClientID) == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	if strings.TrimSpace(opts.ClientSecret) == "" {
		return nil, fmt.Errorf("client secret is required")
	}

	base := opts.BaseURL
	if strings.TrimSpace(base) == "" {
		base = defaultBaseURL
	}

	parsedBase, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	logger := opts.Logger
	if logger == nil {
		logger = newDefaultLogger()
	}

	limiterOpts := DefaultLimiterOptions()
	if opts.LimiterOptions != nil {
		limiterOpts = *opts.LimiterOptions
	}

	retryConfig := DefaultRetryConfig()
	if opts.RetryConfig != nil {
		retryConfig = *opts.RetryConfig
	}

	client := &Client{
		clientID:          opts.ClientID,
		clientSecret:      opts.ClientSecret,
		namespace:         opts.Namespace,
		disableTokenCache: opts.DisableTokenCache,
		httpClient:        httpClient,
		baseURL:           parsedBase,
		logger:            logger,
		limiter:           NewRateLimiter(limiterOpts),
		retryConfig:       retryConfig,
	}

	client.Object = newObjectService(client)
	client.Department = &DepartmentService{client: client}
	client.Function = &FunctionService{client: client}
	client.Page = &PageService{client: client}
	client.Attachment = newAttachmentService(client)
	client.Global = newGlobalService(client)
	client.Automation = newAutomationService(client)

	client.log(LoggerLevelInfo, "[client] Client initialized successfully")
	return client, nil
}

// Init primes the client by ensuring a valid token is available.
func (c *Client) Init(ctx context.Context) error {
	if err := c.ensureTokenValid(ctx); err != nil {
		return err
	}
	c.log(LoggerLevelInfo, "[client] Client initialized and ready")
	return nil
}

// SetLoggerLevel adjusts the verbosity level for the client logger.
func (c *Client) SetLoggerLevel(level LoggerLevel) {
	if c.logger == nil {
		return
	}
	c.logger.SetLevel(level)
	c.log(LoggerLevelInfo, "[logger] Log level set to %s", level.String())
}

// Token returns the current access token.
func (c *Client) Token() string {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.accessToken
}

// TokenExpiresIn returns the remaining time until the current token expires.
// The second return value indicates whether a valid token is cached.
func (c *Client) TokenExpiresIn() (time.Duration, bool) {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()

	if c.accessToken == "" || c.expireTime.IsZero() {
		return 0, false
	}

	remaining := time.Until(c.expireTime)
	if remaining < 0 {
		return 0, true
	}
	return remaining, true
}

// Namespace returns the namespace associated with the client.
func (c *Client) Namespace() string {
	return c.namespace
}

func (c *Client) ensureTokenValid(ctx context.Context) error {
	if c.disableTokenCache {
		c.log(LoggerLevelDebug, "[auth] Token cache disabled, refreshing token")
		return c.refreshAccessToken(ctx)
	}

	c.tokenMu.RLock()
	tokenValid := c.accessToken != "" && time.Until(c.expireTime) > time.Minute
	tokenRefreshing := c.tokenRefreshing
	c.tokenMu.RUnlock()

	if tokenValid {
		return nil
	}

	// If another goroutine is already refreshing, wait for it
	if tokenRefreshing {
		time.Sleep(50 * time.Millisecond)
		c.tokenMu.RLock()
		tokenValid = c.accessToken != "" && time.Until(c.expireTime) > time.Minute
		c.tokenMu.RUnlock()
		if tokenValid {
			return nil
		}
	}

	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Re-check after acquiring the lock to avoid redundant refreshes.
	if c.accessToken != "" && time.Until(c.expireTime) > time.Minute {
		return nil
	}

	// Set flag to indicate token refresh is in progress
	c.tokenRefreshing = true
	defer func() {
		c.tokenRefreshing = false
	}()

	return c.refreshAccessTokenLocked(ctx)
}

func (c *Client) refreshAccessToken(ctx context.Context) error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	return c.refreshAccessTokenLocked(ctx)
}

func (c *Client) refreshAccessTokenLocked(ctx context.Context) error {
	payload := map[string]string{
		"clientId":     c.clientID,
		"clientSecret": c.clientSecret,
	}

	c.log(LoggerLevelDebug, "[auth] Refreshing access token")
	resp, err := c.doJSON(ctx, http.MethodPost, "/auth/v1/appToken", payload, false, nil)
	if err != nil {
		return err
	}

	if resp.Code != "0" {
		return fmt.Errorf("failed to fetch access token: %s", resp.Msg)
	}

	var data TokenResponseData
	if err := resp.DecodeData(&data); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	if strings.TrimSpace(data.AccessToken) == "" {
		return fmt.Errorf("received empty access token")
	}

	expireTime := time.UnixMilli(data.ExpireTime)
	c.accessToken = data.AccessToken
	c.expireTime = expireTime

	c.log(LoggerLevelInfo, "[auth] Access token refreshed successfully")
	return nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, auth bool, headers map[string]string) (*APIResponse, error) {
	var reader io.Reader

	if body != nil {
		buf := &bytes.Buffer{}
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(body); err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
		reader = buf

		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Content-Type"] = "application/json"
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Accept"]; !ok {
		headers["Accept"] = "application/json"
	}

	var resp *http.Response
	var err error

	// Execute with retry logic
	retryErr := Retry(ctx, c.retryConfig, func() error {
		resp, err = c.doRequestRaw(ctx, method, path, reader, headers, auth)
		if err != nil {
			return &NetworkError{Operation: "http request", Err: err}
		}

		// Check for retryable HTTP status codes
		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode == http.StatusInternalServerError ||
			resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout {
			defer resp.Body.Close()
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			return newAPIError(resp.StatusCode, "", string(bodyBytes), method, path, nil)
		}

		return nil
	})

	if retryErr != nil {
		return nil, retryErr
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, newAPIError(resp.StatusCode, "", string(bodyBytes), method, path, nil)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	// Check for API-level errors
	if apiResp.Code != "0" && apiResp.Code != "" {
		requestID := resp.Header.Get("X-Request-Id")
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Code:       apiResp.Code,
			Message:    apiResp.Msg,
			RequestID:  requestID,
			Method:     method,
			Endpoint:   path,
		}
		c.log(LoggerLevelWarn, "[client] API error: %v", apiErr)
	}

	return &apiResp, nil
}

func (c *Client) doBinary(ctx context.Context, method, path string, body io.Reader, headers map[string]string, auth bool) ([]byte, http.Header, error) {
	resp, err := c.doRequestRaw(ctx, method, path, body, headers, auth)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, nil, fmt.Errorf("api request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read binary response: %w", err)
	}

	return data, resp.Header.Clone(), nil
}

func (c *Client) doRequestRaw(ctx context.Context, method, path string, body io.Reader, headers map[string]string, auth bool) (*http.Response, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid request path: %w", err)
	}

	endpoint := c.baseURL.ResolveReference(rel)

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if auth {
		if token := c.getAccessToken(); token != "" {
			req.Header.Set("Authorization", token)
		}
	}

	for k, v := range headers {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		req.Header.Set(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *Client) getAccessToken() string {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.accessToken
}

func (c *Client) log(level LoggerLevel, format string, args ...any) {
	if c.logger == nil {
		return
	}
	c.logger.Log(level, format, args...)
}
