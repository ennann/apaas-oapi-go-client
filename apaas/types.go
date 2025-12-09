package apaas

import (
	"encoding/json"
	"fmt"
)

// APIResponse represents a generic OpenAPI response envelope.
type APIResponse struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// DecodeData unmarshals the Data field into the provided target.
func (r *APIResponse) DecodeData(v any) error {
	if r == nil {
		return fmt.Errorf("api response is nil")
	}
	if len(r.Data) == 0 {
		return nil
	}
	return json.Unmarshal(r.Data, v)
}

// TokenResponse models the access token API response.
type TokenResponse struct {
	Code string            `json:"code"`
	Msg  string            `json:"msg"`
	Data TokenResponseData `json:"data"`
}

// TokenResponseData holds token and expiry information.
type TokenResponseData struct {
	AccessToken string `json:"accessToken"`
	ExpireTime  int64  `json:"expireTime"`
}

// Map is a convenient alias for JSON-like maps.
type Map = map[string]any

// RecordsIteratorResult aggregates paginated records.
type RecordsIteratorResult struct {
	Total int              `json:"total"`
	Items []map[string]any `json:"items"`
}

// BatchOperationResult 批量操作结果（创建/更新/删除）
type BatchOperationResult struct {
	Total        int             `json:"total"`
	Success      []OperationItem `json:"success"`
	Failed       []OperationItem `json:"failed"`
	SuccessCount int             `json:"successCount"`
	FailedCount  int             `json:"failedCount"`
}

// OperationItem 操作项（成功或失败）
type OperationItem struct {
	ID      string `json:"_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// BatchResponses groups multiple API responses.
type BatchResponses []*APIResponse

// FlowOperator identifies the operator triggering an automation flow.
type FlowOperator struct {
	ID    int64  `json:"_id"`
	Email string `json:"email"`
}
