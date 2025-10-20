package apaas

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ObjectService provides access to object-related APIs.
type ObjectService struct {
	client   *Client
	Metadata *ObjectMetadataService
	Search   *ObjectSearchService
	Create   *ObjectCreateService
	Update   *ObjectUpdateService
	Delete   *ObjectDeleteService
}

func newObjectService(client *Client) *ObjectService {
	service := &ObjectService{client: client}
	service.Metadata = &ObjectMetadataService{client: client}
	service.Search = &ObjectSearchService{client: client}
	service.Create = &ObjectCreateService{client: client}
	service.Update = &ObjectUpdateService{client: client}
	service.Delete = &ObjectDeleteService{client: client}
	return service
}

// ObjectListParams controls pagination for listing objects.
type ObjectListParams struct {
	Offset int               `json:"offset"`
	Filter *ObjectListFilter `json:"filter,omitempty"`
	Limit  int               `json:"limit"`
}

// ObjectListFilter narrows the object listing result.
type ObjectListFilter struct {
	Type       string `json:"type,omitempty"`
	QuickQuery string `json:"quickQuery,omitempty"`
}

// ObjectMetadataService fetches metadata resources.
type ObjectMetadataService struct {
	client *Client
}

// ObjectMetadataFieldParams identifies a single field.
type ObjectMetadataFieldParams struct {
	ObjectName string
	FieldName  string
}

// ObjectMetadataFieldsParams identifies an object.
type ObjectMetadataFieldsParams struct {
	ObjectName string
}

// ObjectSearchService performs record queries.
type ObjectSearchService struct {
	client *Client
}

// ObjectSearchRecordParams requests a single record.
type ObjectSearchRecordParams struct {
	ObjectName string
	RecordID   string
	Select     []string
}

// ObjectSearchRecordsParams requests multiple records.
type ObjectSearchRecordsParams struct {
	ObjectName string
	Data       map[string]any
}

// ObjectRecordsIteratorParams fetches all records via pagination.
type ObjectRecordsIteratorParams struct {
	ObjectName string
	Data       map[string]any
}

// ObjectCreateService inserts records.
type ObjectCreateService struct {
	client *Client
}

// ObjectCreateRecordParams creates a single record.
type ObjectCreateRecordParams struct {
	ObjectName string
	Record     map[string]any
}

// ObjectCreateRecordsParams creates multiple records in a single request.
type ObjectCreateRecordsParams struct {
	ObjectName string
	Records    []map[string]any
}

// ObjectCreateRecordsIteratorParams creates records in batches.
type ObjectCreateRecordsIteratorParams struct {
	ObjectName string
	Records    []map[string]any
}

// ObjectUpdateService updates records.
type ObjectUpdateService struct {
	client *Client
}

// ObjectUpdateRecordParams updates a single record.
type ObjectUpdateRecordParams struct {
	ObjectName string
	RecordID   string
	Record     map[string]any
}

// ObjectUpdateRecordsParams updates up to 100 records per request.
type ObjectUpdateRecordsParams struct {
	ObjectName string
	Records    []map[string]any
}

// ObjectUpdateRecordsIteratorParams updates records in batches.
type ObjectUpdateRecordsIteratorParams struct {
	ObjectName string
	Records    []map[string]any
}

// ObjectDeleteService deletes records.
type ObjectDeleteService struct {
	client *Client
}

// ObjectDeleteRecordParams removes a single record.
type ObjectDeleteRecordParams struct {
	ObjectName string
	RecordID   string
}

// ObjectDeleteRecordsParams removes up to 100 records per request.
type ObjectDeleteRecordsParams struct {
	ObjectName string
	IDs        []string
}

// ObjectDeleteRecordsIteratorParams removes records in batches.
type ObjectDeleteRecordsIteratorParams struct {
	ObjectName string
	IDs        []string
}

// List returns available objects (data tables).
func (s *ObjectService) List(ctx context.Context, params ObjectListParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/api/data/v1/namespaces/%s/meta/objects/list", url.PathEscape(s.client.namespace))

	payload := map[string]any{
		"offset": params.Offset,
		"limit":  params.Limit,
	}
	if params.Filter != nil {
		payload["filter"] = params.Filter
	}

	s.client.log(LoggerLevelDebug, "[object.list] Fetching objects list: offset=%d, limit=%d", params.Offset, params.Limit)

	resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[object.list] Objects list fetched: code=%s", resp.Code)
	return resp, nil
}

// Field retrieves metadata for a specific field.
func (s *ObjectMetadataService) Field(ctx context.Context, params ObjectMetadataFieldParams) (*APIResponse, error) {
	var resp *APIResponse

	err := s.client.limiter.Do(ctx, func() error {
		if err := s.client.ensureTokenValid(ctx); err != nil {
			return err
		}

		endpoint := fmt.Sprintf(
			"/api/data/v1/namespaces/%s/meta/objects/%s/fields/%s",
			url.PathEscape(s.client.namespace),
			url.PathEscape(params.ObjectName),
			url.PathEscape(params.FieldName),
		)

		s.client.log(LoggerLevelDebug, "[object.metadata.field] Fetching field metadata: %s.%s", params.ObjectName, params.FieldName)

		var err error
		resp, err = s.client.doJSON(ctx, http.MethodGet, endpoint, nil, true, nil)
		return err
	})

	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[object.metadata.field] Field metadata fetched: %s.%s, code=%s", params.ObjectName, params.FieldName, resp.Code)
	return resp, nil
}

// Fields retrieves metadata for all fields on an object.
func (s *ObjectMetadataService) Fields(ctx context.Context, params ObjectMetadataFieldsParams) (*APIResponse, error) {
	var resp *APIResponse

	err := s.client.limiter.Do(ctx, func() error {
		if err := s.client.ensureTokenValid(ctx); err != nil {
			return err
		}

		endpoint := fmt.Sprintf(
			"/api/data/v1/namespaces/%s/meta/objects/%s",
			url.PathEscape(s.client.namespace),
			url.PathEscape(params.ObjectName),
		)

		s.client.log(LoggerLevelDebug, "[object.metadata.fields] Fetching all fields metadata: %s", params.ObjectName)

		var err error
		resp, err = s.client.doJSON(ctx, http.MethodGet, endpoint, nil, true, nil)
		return err
	})

	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[object.metadata.fields] All fields metadata fetched: %s, code=%s", params.ObjectName, resp.Code)
	return resp, nil
}

// Record retrieves a single record.
func (s *ObjectSearchService) Record(ctx context.Context, params ObjectSearchRecordParams) (*APIResponse, error) {
	s.client.log(LoggerLevelInfo, "[object.search.record] Querying record: %s", params.RecordID)

	var resp *APIResponse
	err := s.client.limiter.Do(ctx, func() error {
		if err := s.client.ensureTokenValid(ctx); err != nil {
			return err
		}

		endpoint := fmt.Sprintf(
			"/v1/data/namespaces/%s/objects/%s/records/%s",
			url.PathEscape(s.client.namespace),
			url.PathEscape(params.ObjectName),
			url.PathEscape(params.RecordID),
		)

		payload := map[string]any{
			"select": params.Select,
		}

		var err error
		resp, err = s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[object.search.record] Record queried: %s.%s, code=%s", params.ObjectName, params.RecordID, resp.Code)
	return resp, nil
}

// Records retrieves up to 100 records.
func (s *ObjectSearchService) Records(ctx context.Context, params ObjectSearchRecordsParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/v1/data/namespaces/%s/objects/%s/records_query",
		url.PathEscape(s.client.namespace),
		url.PathEscape(params.ObjectName),
	)

	resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, params.Data, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[object.search.records] Records queried: %s, code=%s", params.ObjectName, resp.Code)
	return resp, nil
}

// RecordsWithIterator gathers all records using pagination.
func (s *ObjectSearchService) RecordsWithIterator(ctx context.Context, params ObjectRecordsIteratorParams) (*RecordsIteratorResult, error) {
	results := &RecordsIteratorResult{
		Items: make([]map[string]any, 0),
	}

	nextToken := ""

	for {
		requestPayload := cloneMap(params.Data)
		requestPayload["page_token"] = nextToken

		var resp *APIResponse
		err := s.client.limiter.Do(ctx, func() error {
			var err error
			resp, err = s.Records(ctx, ObjectSearchRecordsParams{
				ObjectName: params.ObjectName,
				Data:       requestPayload,
			})
			return err
		})
		if err != nil {
			return nil, err
		}

		var page struct {
			Items         []map[string]any `json:"items"`
			Total         int              `json:"total"`
			NextPageToken string           `json:"next_page_token"`
		}
		if err := resp.DecodeData(&page); err != nil {
			return nil, fmt.Errorf("failed to decode paginated records: %w", err)
		}

		if results.Total == 0 && page.Total > 0 {
			results.Total = page.Total
		}
		if len(page.Items) > 0 {
			results.Items = append(results.Items, page.Items...)
		}

		s.client.log(LoggerLevelInfo, "[object.search.recordsWithIterator] Page completed: items=%d, next=%s", len(page.Items), page.NextPageToken)

		if page.NextPageToken == "" {
			break
		}
		nextToken = page.NextPageToken
	}

	return results, nil
}

// Record creates a single record.
func (s *ObjectCreateService) Record(ctx context.Context, params ObjectCreateRecordParams) (*APIResponse, error) {
	s.client.log(LoggerLevelInfo, "[object.create.record] Creating record in: %s", params.ObjectName)

	var resp *APIResponse
	err := s.client.limiter.Do(ctx, func() error {
		if err := s.client.ensureTokenValid(ctx); err != nil {
			return err
		}

		endpoint := fmt.Sprintf(
			"/v1/data/namespaces/%s/objects/%s/records",
			url.PathEscape(s.client.namespace),
			url.PathEscape(params.ObjectName),
		)

		payload := map[string]any{"record": params.Record}

		var err error
		resp, err = s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelInfo, "[object.create.record] Record created: %s", params.ObjectName)
	return resp, nil
}

// Records creates up to 100 records in a single request.
func (s *ObjectCreateService) Records(ctx context.Context, params ObjectCreateRecordsParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/v1/data/namespaces/%s/objects/%s/records_batch",
		url.PathEscape(s.client.namespace),
		url.PathEscape(params.ObjectName),
	)

	payload := map[string]any{"records": params.Records}
	resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelInfo, "[object.create.records] Creating %d records in: %s", len(params.Records), params.ObjectName)
	return resp, nil
}

// RecordsWithIterator creates records in batches of 100.
func (s *ObjectCreateService) RecordsWithIterator(ctx context.Context, params ObjectCreateRecordsIteratorParams) (*RecordsIteratorResult, error) {
	total := len(params.Records)
	result := &RecordsIteratorResult{
		Total: total,
		Items: make([]map[string]any, 0, total),
	}

	if total == 0 {
		return result, nil
	}

	const chunkSize = 100
	for index := 0; index < total; index += chunkSize {
		end := index + chunkSize
		if end > total {
			end = total
		}
		chunk := params.Records[index:end]
		chunkIndex := index/chunkSize + 1

		s.client.log(LoggerLevelInfo, "[object.create.recordsWithIterator] Processing chunk %d/%d: %d records", chunkIndex, (total+chunkSize-1)/chunkSize, len(chunk))

		var resp *APIResponse
		err := s.client.limiter.Do(ctx, func() error {
			var err error
			resp, err = s.Records(ctx, ObjectCreateRecordsParams{
				ObjectName: params.ObjectName,
				Records:    chunk,
			})
			return err
		})
		if err != nil {
			return nil, err
		}

		var page struct {
			Items []map[string]any `json:"items"`
		}
		if err := resp.DecodeData(&page); err != nil {
			return nil, fmt.Errorf("failed to decode batch create response: %w", err)
		}
		if len(page.Items) > 0 {
			result.Items = append(result.Items, page.Items...)
		}

		s.client.log(LoggerLevelDebug, "[object.create.recordsWithIterator] Chunk %d completed: created=%d", chunkIndex, len(page.Items))
	}

	return result, nil
}

// Record updates a single record.
func (s *ObjectUpdateService) Record(ctx context.Context, params ObjectUpdateRecordParams) (*APIResponse, error) {
	s.client.log(LoggerLevelInfo, "[object.update.record] Updating record: %s", params.RecordID)

	var resp *APIResponse
	err := s.client.limiter.Do(ctx, func() error {
		if err := s.client.ensureTokenValid(ctx); err != nil {
			return err
		}

		endpoint := fmt.Sprintf(
			"/v1/data/namespaces/%s/objects/%s/records/%s",
			url.PathEscape(s.client.namespace),
			url.PathEscape(params.ObjectName),
			url.PathEscape(params.RecordID),
		)

		payload := map[string]any{"record": params.Record}
		var err error
		resp, err = s.client.doJSON(ctx, http.MethodPatch, endpoint, payload, true, nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[object.update.record] Record updated: %s.%s, code=%s", params.ObjectName, params.RecordID, resp.Code)
	return resp, nil
}

// Records updates up to 100 records.
func (s *ObjectUpdateService) Records(ctx context.Context, params ObjectUpdateRecordsParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/v1/data/namespaces/%s/objects/%s/records/records_batch",
		url.PathEscape(s.client.namespace),
		url.PathEscape(params.ObjectName),
	)

	payload := map[string]any{"records": params.Records}
	resp, err := s.client.doJSON(ctx, http.MethodPatch, endpoint, payload, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelInfo, "[object.update.records] Records updated: %s", params.ObjectName)
	return resp, nil
}

// RecordsWithIterator updates records in batches.
func (s *ObjectUpdateService) RecordsWithIterator(ctx context.Context, params ObjectUpdateRecordsIteratorParams) (BatchResponses, error) {
	total := len(params.Records)
	if total == 0 {
		return nil, nil
	}

	responses := make(BatchResponses, 0, (total+99)/100)

	const chunkSize = 100
	for index := 0; index < total; index += chunkSize {
		end := index + chunkSize
		if end > total {
			end = total
		}
		chunk := params.Records[index:end]
		chunkIndex := index/chunkSize + 1

		s.client.log(LoggerLevelDebug, "[object.update.recordsWithIterator] Processing chunk %d: %d records", chunkIndex, len(chunk))

		var resp *APIResponse
		err := s.client.limiter.Do(ctx, func() error {
			if err := s.client.ensureTokenValid(ctx); err != nil {
				return err
			}

			endpoint := fmt.Sprintf(
				"/v1/data/namespaces/%s/objects/%s/records_batch",
				url.PathEscape(s.client.namespace),
				url.PathEscape(params.ObjectName),
			)

			payload := map[string]any{"records": chunk}
			var err error
			resp, err = s.client.doJSON(ctx, http.MethodPatch, endpoint, payload, true, nil)
			return err
		})
		if err != nil {
			return nil, err
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// Record deletes a single record.
func (s *ObjectDeleteService) Record(ctx context.Context, params ObjectDeleteRecordParams) (*APIResponse, error) {
	s.client.log(LoggerLevelInfo, "[object.delete.record] Deleting record: %s.%s", params.ObjectName, params.RecordID)

	var resp *APIResponse
	err := s.client.limiter.Do(ctx, func() error {
		if err := s.client.ensureTokenValid(ctx); err != nil {
			return err
		}

		endpoint := fmt.Sprintf(
			"/v1/data/namespaces/%s/objects/%s/records/%s",
			url.PathEscape(s.client.namespace),
			url.PathEscape(params.ObjectName),
			url.PathEscape(params.RecordID),
		)

		var err error
		resp, err = s.client.doJSON(ctx, http.MethodDelete, endpoint, nil, true, nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelInfo, "[object.delete.record] Record deleted: %s.%s", params.ObjectName, params.RecordID)
	return resp, nil
}

// Records deletes up to 100 records in a single request.
func (s *ObjectDeleteService) Records(ctx context.Context, params ObjectDeleteRecordsParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/v1/data/namespaces/%s/objects/%s/records_batch",
		url.PathEscape(s.client.namespace),
		url.PathEscape(params.ObjectName),
	)

	payload := map[string]any{"ids": params.IDs}
	resp, err := s.client.doJSON(ctx, http.MethodDelete, endpoint, payload, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelInfo, "[object.delete.records] Records deleted: %s, count=%d", params.ObjectName, len(params.IDs))
	return resp, nil
}

// RecordsWithIterator deletes records in batches of 100.
func (s *ObjectDeleteService) RecordsWithIterator(ctx context.Context, params ObjectDeleteRecordsIteratorParams) (BatchResponses, error) {
	total := len(params.IDs)
	if total == 0 {
		return nil, nil
	}

	responses := make(BatchResponses, 0, (total+99)/100)

	const chunkSize = 100
	for index := 0; index < total; index += chunkSize {
		end := index + chunkSize
		if end > total {
			end = total
		}
		chunk := params.IDs[index:end]
		chunkIndex := index/chunkSize + 1

		s.client.log(LoggerLevelInfo, "[object.delete.recordsWithIterator] Processing chunk %d/%d: %d records", chunkIndex, (total+chunkSize-1)/chunkSize, len(chunk))

		var resp *APIResponse
		err := s.client.limiter.Do(ctx, func() error {
			if err := s.client.ensureTokenValid(ctx); err != nil {
				return err
			}

			endpoint := fmt.Sprintf(
				"/v1/data/namespaces/%s/objects/%s/records_batch",
				url.PathEscape(s.client.namespace),
				url.PathEscape(params.ObjectName),
			)

			payload := map[string]any{"ids": chunk}
			var err error
			resp, err = s.client.doJSON(ctx, http.MethodDelete, endpoint, payload, true, map[string]string{
				"Content-Type": "application/json",
			})
			return err
		})
		if err != nil {
			return nil, err
		}

		responses = append(responses, resp)
	}

	return responses, nil
}
