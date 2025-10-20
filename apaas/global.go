package apaas

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GlobalService provides access to global options and variables.
type GlobalService struct {
	client    *Client
	Options   *GlobalOptionsService
	Variables *GlobalVariablesService
}

// GlobalOptionsService handles global options APIs.
type GlobalOptionsService struct {
	client *Client
}

// GlobalVariablesService handles global variables APIs.
type GlobalVariablesService struct {
	client *Client
}

type globalListParams struct {
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Filter map[string]any `json:"filter,omitempty"`
}

// newGlobalService constructs a GlobalService.
func newGlobalService(client *Client) *GlobalService {
	return &GlobalService{
		client:    client,
		Options:   &GlobalOptionsService{client: client},
		Variables: &GlobalVariablesService{client: client},
	}
}

// Detail retrieves global option details.
func (s *GlobalOptionsService) Detail(ctx context.Context, apiName string) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/data/v1/namespaces/%s/globalOptions/%s",
		url.PathEscape(s.client.namespace),
		url.PathEscape(apiName),
	)

	s.client.log(LoggerLevelInfo, "[global.options.detail] Fetching global option detail: %s", apiName)

	resp, err := s.client.doJSON(ctx, http.MethodGet, endpoint, nil, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[global.options.detail] Global option detail fetched: %s, code=%s", apiName, resp.Code)
	return resp, nil
}

// List retrieves a paginated global options list.
func (s *GlobalOptionsService) List(ctx context.Context, limit, offset int, filter map[string]any) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/data/v1/namespaces/%s/globalOptions/list",
		url.PathEscape(s.client.namespace),
	)

	payload := globalListParams{
		Limit:  limit,
		Offset: offset,
	}
	if filter != nil {
		payload.Filter = filter
	}

	s.client.log(LoggerLevelInfo, "[global.options.list] Fetching global options list: offset=%d, limit=%d", offset, limit)

	resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[global.options.list] Global options list fetched: code=%s", resp.Code)
	return resp, nil
}

// ListWithIterator retrieves all global options automatically.
func (s *GlobalOptionsService) ListWithIterator(ctx context.Context, limit int, filter map[string]any) (*RecordsIteratorResult, error) {
	if limit <= 0 {
		limit = 100
	}

	results := &RecordsIteratorResult{
		Items: make([]map[string]any, 0),
	}

	offset := 0

	for {
		resp, err := s.List(ctx, limit, offset, filter)
		if err != nil {
			return nil, err
		}

		var page struct {
			Items []map[string]any `json:"items"`
			Total int              `json:"total"`
		}
		if err := resp.DecodeData(&page); err != nil {
			return nil, fmt.Errorf("failed to decode global options list: %w", err)
		}

		if results.Total == 0 && page.Total > 0 {
			results.Total = page.Total
		}
		if len(page.Items) > 0 {
			results.Items = append(results.Items, page.Items...)
		}

		s.client.log(LoggerLevelInfo, "[global.options.listWithIterator] Page completed: items=%d, offset=%d", len(page.Items), offset)

		offset += limit
		if len(results.Items) >= results.Total || len(page.Items) == 0 {
			break
		}
	}

	return results, nil
}

// Detail retrieves global variable details.
func (s *GlobalVariablesService) Detail(ctx context.Context, apiName string) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/data/v1/namespaces/%s/globalVariables/%s",
		url.PathEscape(s.client.namespace),
		url.PathEscape(apiName),
	)

	s.client.log(LoggerLevelInfo, "[global.variables.detail] Fetching global variable detail: %s", apiName)

	resp, err := s.client.doJSON(ctx, http.MethodGet, endpoint, nil, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[global.variables.detail] Global variable detail fetched: %s, code=%s", apiName, resp.Code)
	return resp, nil
}

// List retrieves a paginated global variables list.
func (s *GlobalVariablesService) List(ctx context.Context, limit, offset int, filter map[string]any) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/data/v1/namespaces/%s/globalVariables/list",
		url.PathEscape(s.client.namespace),
	)

	payload := globalListParams{
		Limit:  limit,
		Offset: offset,
	}
	if filter != nil {
		payload.Filter = filter
	}

	s.client.log(LoggerLevelInfo, "[global.variables.list] Fetching global variables list: offset=%d, limit=%d", offset, limit)

	resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[global.variables.list] Global variables list fetched: code=%s", resp.Code)
	return resp, nil
}

// ListWithIterator retrieves all global variables automatically.
func (s *GlobalVariablesService) ListWithIterator(ctx context.Context, limit int, filter map[string]any) (*RecordsIteratorResult, error) {
	if limit <= 0 {
		limit = 100
	}

	results := &RecordsIteratorResult{
		Items: make([]map[string]any, 0),
	}

	offset := 0

	for {
		resp, err := s.List(ctx, limit, offset, filter)
		if err != nil {
			return nil, err
		}

		var page struct {
			Items []map[string]any `json:"items"`
			Total int              `json:"total"`
		}
		if err := resp.DecodeData(&page); err != nil {
			return nil, fmt.Errorf("failed to decode global variables list: %w", err)
		}

		if results.Total == 0 && page.Total > 0 {
			results.Total = page.Total
		}
		if len(page.Items) > 0 {
			results.Items = append(results.Items, page.Items...)
		}

		s.client.log(LoggerLevelInfo, "[global.variables.listWithIterator] Page completed: items=%d, offset=%d", len(page.Items), offset)

		offset += limit
		if len(results.Items) >= results.Total || len(page.Items) == 0 {
			break
		}
	}

	return results, nil
}
