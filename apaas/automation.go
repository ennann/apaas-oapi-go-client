package apaas

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AutomationService groups V1 and V2 automation flow APIs.
type AutomationService struct {
	client *Client
	V1     *AutomationV1Service
	V2     *AutomationV2Service
}

// AutomationV1Service handles v1 flow execution.
type AutomationV1Service struct {
	client *Client
}

// AutomationV2Service handles v2 flow execution.
type AutomationV2Service struct {
	client *Client
}

// AutomationV1ExecuteParams executes a v1 flow.
type AutomationV1ExecuteParams struct {
	FlowAPIName string
	Operator    FlowOperator
	Params      map[string]any
}

// AutomationV2ExecuteParams executes a v2 flow.
type AutomationV2ExecuteParams struct {
	FlowAPIName   string
	Operator      FlowOperator
	Params        map[string]any
	IsResubmit    *bool
	PreInstanceID string
}

func newAutomationService(client *Client) *AutomationService {
	return &AutomationService{
		client: client,
		V1:     &AutomationV1Service{client: client},
		V2:     &AutomationV2Service{client: client},
	}
}

// Execute runs a v1 automation flow.
func (s *AutomationV1Service) Execute(ctx context.Context, params AutomationV1ExecuteParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/api/flow/v1/namespaces/%s/flows/%s/execute",
		url.PathEscape(s.client.namespace),
		url.PathEscape(params.FlowAPIName),
	)

	payload := map[string]any{
		"operator": params.Operator,
		"params":   params.Params,
	}

	s.client.log(LoggerLevelInfo, "[automation.v1.execute] Executing flow: %s", params.FlowAPIName)

	resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[automation.v1.execute] Flow executed: %s, code=%s", params.FlowAPIName, resp.Code)
	return resp, nil
}

// Execute runs a v2 automation flow.
func (s *AutomationV2Service) Execute(ctx context.Context, params AutomationV2ExecuteParams) (*APIResponse, error) {
	if err := s.client.ensureTokenValid(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"/v2/namespaces/%s/flows/%s/execute",
		url.PathEscape(s.client.namespace),
		url.PathEscape(params.FlowAPIName),
	)

	payload := map[string]any{
		"operator": params.Operator,
		"params":   params.Params,
	}

	if params.IsResubmit != nil {
		payload["is_resubmit"] = *params.IsResubmit
	}
	if params.PreInstanceID != "" {
		payload["pre_instance_id"] = params.PreInstanceID
	}

	s.client.log(LoggerLevelInfo, "[automation.v2.execute] Executing flow: %s", params.FlowAPIName)

	resp, err := s.client.doJSON(ctx, http.MethodPost, endpoint, payload, true, nil)
	if err != nil {
		return nil, err
	}

	s.client.log(LoggerLevelDebug, "[automation.v2.execute] Flow executed: %s, code=%s", params.FlowAPIName, resp.Code)
	return resp, nil
}
