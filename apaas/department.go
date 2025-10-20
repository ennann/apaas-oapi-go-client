package apaas

import (
	"context"
	"fmt"
	"net/http"
)

// DepartmentService handles department ID exchange operations.
type DepartmentService struct {
	client *Client
}

// DepartmentExchangeParams exchanges a single department identifier.
type DepartmentExchangeParams struct {
	DepartmentIDType string
	DepartmentID     string
}

// DepartmentBatchExchangeParams exchanges multiple department identifiers.
type DepartmentBatchExchangeParams struct {
	DepartmentIDType string
	DepartmentIDs    []string
}

// Exchange performs a single department ID exchange.
func (s *DepartmentService) Exchange(ctx context.Context, params DepartmentExchangeParams) (map[string]any, error) {
	var result map[string]any

	err := s.client.limiter.Do(ctx, func() error {
		if err := s.client.ensureTokenValid(ctx); err != nil {
			return err
		}

		endpoint := "/api/integration/v2/feishu/getDepartments"
		payload := map[string]any{
			"department_id_type": params.DepartmentIDType,
			"department_ids":     []string{params.DepartmentID},
		}

		s.client.log(LoggerLevelInfo, "[department.exchange] Exchanging department ID: %s", params.DepartmentID)

		resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
		if err != nil {
			return err
		}

		var data []map[string]any
		if err := resp.DecodeData(&data); err != nil {
			return fmt.Errorf("failed to decode department exchange response: %w", err)
		}
		if len(data) == 0 {
			return fmt.Errorf("department exchange returned no data")
		}

		result = data[0]
		return nil
	})

	return result, err
}

// BatchExchange exchanges department IDs in batches of 100.
func (s *DepartmentService) BatchExchange(ctx context.Context, params DepartmentBatchExchangeParams) ([]map[string]any, error) {
	if len(params.DepartmentIDs) == 0 {
		return nil, nil
	}

	const chunkSize = 100
	results := make([]map[string]any, 0, len(params.DepartmentIDs))

	for index := 0; index < len(params.DepartmentIDs); index += chunkSize {
		end := index + chunkSize
		if end > len(params.DepartmentIDs) {
			end = len(params.DepartmentIDs)
		}

		chunk := params.DepartmentIDs[index:end]
		chunkIndex := index/chunkSize + 1

		s.client.log(LoggerLevelInfo, "[department.batchExchange] Processing chunk %d/%d: %d IDs", chunkIndex, (len(params.DepartmentIDs)+chunkSize-1)/chunkSize, len(chunk))

		err := s.client.limiter.Do(ctx, func() error {
			if err := s.client.ensureTokenValid(ctx); err != nil {
				return err
			}

			endpoint := "/api/integration/v2/feishu/getDepartments"
			payload := map[string]any{
				"department_id_type": params.DepartmentIDType,
				"department_ids":     chunk,
			}

			resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
			if err != nil {
				return err
			}

			var data []map[string]any
			if err := resp.DecodeData(&data); err != nil {
				return fmt.Errorf("failed to decode department batch response: %w", err)
			}

			results = append(results, data...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}
