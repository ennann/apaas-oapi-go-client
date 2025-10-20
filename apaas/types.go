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

// BatchResponses groups multiple API responses.
type BatchResponses []*APIResponse

// FlowOperator identifies the operator triggering an automation flow.
type FlowOperator struct {
	ID    int64  `json:"_id"`
	Email string `json:"email"`
}
