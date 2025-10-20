package apaas

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// FunctionService executes cloud functions.
type FunctionService struct {
	client *Client
}

// FunctionInvokeParams defines the payload to invoke a cloud function.
type FunctionInvokeParams struct {
	Name   string
	Params map[string]any
}

// Invoke executes a cloud function.
func (s *FunctionService) Invoke(ctx context.Context, params FunctionInvokeParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/cloudfunction/v1/namespaces/%s/invoke/%s",
		url.PathEscape(s.client.namespace),
		url.PathEscape(params.Name),
	)

	payload := map[string]any{
		"params": params.Params,
	}

	s.client.log(LoggerLevelInfo, "[function.invoke] Invoking cloud function: %s", params.Name)

	resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[function.invoke] Cloud function invoked: %s, code=%s", params.Name, resp.Code)
	return resp, nil
}
