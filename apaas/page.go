package apaas

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// PageService manages page metadata and access URLs.
type PageService struct {
	client *Client
}

// PageListParams controls pagination for listing pages.
type PageListParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// PageListWithIteratorParams retrieves all pages automatically.
type PageListWithIteratorParams struct {
	Limit int
}

// PageDetailParams requests metadata for a page.
type PageDetailParams struct {
	PageID string
}

// PageURLParams requests a sharable page URL.
type PageURLParams struct {
	PageID           string
	PageParams       map[string]any
	ParentPageParams map[string]any
	NavID            string
	TabID            string
}

// List retrieves a page collection.
func (s *PageService) List(ctx context.Context, params PageListParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/builder/v1/namespaces/%s/meta/pages",
		url.PathEscape(s.client.namespace),
	)

	s.client.log(LoggerLevelInfo, "[page.list] Fetching pages list: offset=%d, limit=%d", params.Offset, params.Limit)

	resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, params, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[page.list] Pages list fetched: code=%s", resp.Code)
	return resp, nil
}

// ListWithIterator retrieves all pages by auto-paginating.
func (s *PageService) ListWithIterator(ctx context.Context, params *PageListWithIteratorParams) (*RecordsIteratorResult, error) {
	limit := 100
	if params != nil && params.Limit > 0 {
		limit = params.Limit
	}

	results := &RecordsIteratorResult{
		Items: make([]map[string]any, 0),
	}

	offset := 0

	for {
		resp, err := s.List(ctx, PageListParams{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, err
		}

		var page struct {
			Items []map[string]any `json:"items"`
			Total int              `json:"total"`
		}
		if err := resp.DecodeData(&page); err != nil {
			return nil, fmt.Errorf("failed to decode paginated pages: %w", err)
		}

		if results.Total == 0 && page.Total > 0 {
			results.Total = page.Total
		}

		if len(page.Items) > 0 {
			results.Items = append(results.Items, page.Items...)
		}

		s.client.log(LoggerLevelInfo, "[page.listWithIterator] Page completed: items=%d, offset=%d", len(page.Items), offset)

		offset += limit
		if len(results.Items) >= results.Total || len(page.Items) == 0 {
			break
		}
	}

	return results, nil
}

// Detail fetches metadata for a single page.
func (s *PageService) Detail(ctx context.Context, params PageDetailParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/builder/v1/namespaces/%s/meta/pages/%s",
		url.PathEscape(s.client.namespace),
		url.PathEscape(params.PageID),
	)

	s.client.log(LoggerLevelInfo, "[page.detail] Fetching page detail: %s", params.PageID)

	resp, err := s.client.doJSON(ctx, http.MethodGet, endpoint, nil, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[page.detail] Page detail fetched: %s, code=%s", params.PageID, resp.Code)
	return resp, nil
}

// URL builds an accessible URL for the page.
func (s *PageService) URL(ctx context.Context, params PageURLParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/builder/v1/namespaces/%s/meta/pages/%s/link",
		url.PathEscape(s.client.namespace),
		url.PathEscape(params.PageID),
	)

	payload := map[string]any{}
	if params.PageParams != nil {
		payload["pageParams"] = params.PageParams
	}
	if params.ParentPageParams != nil {
		payload["parentPageParams"] = params.ParentPageParams
	}
	if params.NavID != "" {
		payload["navId"] = params.NavID
	}
	if params.TabID != "" {
		payload["tabId"] = params.TabID
	}

	s.client.log(LoggerLevelInfo, "[page.url] Fetching page URL: %s", params.PageID)

	resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[page.url] Page URL fetched: %s, code=%s", params.PageID, resp.Code)
	return resp, nil
}
