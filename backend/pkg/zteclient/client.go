package zteclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"billing-backend/internal/domain"
)

// Client represents the HTTP client for interacting with snmp-olt-zte service.
type Client interface {
	CheckHealth(ctx context.Context) (map[string]interface{}, error)
	CheckReadyz(ctx context.Context) (map[string]interface{}, error)
	GetUplinks(ctx context.Context, oltID string) (*domain.ZTEUplinksData, error)
	GetONUs(ctx context.Context, oltID string, board int, pon int) ([]domain.ZTEONUInfo, error)
	GetPaginatedONUs(ctx context.Context, oltID string, board int, pon int, page int, limit int) (*domain.ZTEPaginatedONUs, error)
	GetONUDetail(ctx context.Context, oltID string, board int, pon int, onuID int) (*domain.ZTEONUDetail, error)
	GetEmptyONUIDs(ctx context.Context, oltID string, board int, pon int) ([]int, error)
	GetONUSerials(ctx context.Context, oltID string, board int, pon int, noCache bool) ([]domain.ONUSerialInfo, error)
	ClearCache(ctx context.Context, oltID string, board int, pon int) error
}

type client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient initializes a new ZTE SNMP OLT API client.
func NewClient(baseURL, apiKey string) Client {
	cleanURL := strings.TrimRight(baseURL, "/")
	if cleanURL == "" {
		cleanURL = "http://localhost:8081"
	}
	return &client{
		baseURL: cleanURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// APIResponse represents the standard response envelope from snmp-olt-zte service.
type APIResponse[T any] struct {
	Code      int    `json:"code"`
	Status    string `json:"status"`
	Data      T      `json:"data"`
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func (c *client) doRequest(ctx context.Context, method, path string, queryParams url.Values, target interface{}) error {
	fullURL := fmt.Sprintf("%s%s", c.baseURL, path)
	if len(queryParams) > 0 {
		fullURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request to snmp-olt-zte [%s]: %w", fullURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp APIResponse[interface{}]
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil && errResp.ErrorCode != "" {
			return fmt.Errorf("snmp-olt-zte error [%s]: %s (code: %d)", errResp.ErrorCode, errResp.Message, resp.StatusCode)
		}
		return fmt.Errorf("snmp-olt-zte returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if target != nil {
		if err := json.Unmarshal(bodyBytes, target); err != nil {
			return fmt.Errorf("failed to decode response JSON: %w (body: %s)", err, string(bodyBytes))
		}
	}

	return nil
}

// CheckHealth checks if the snmp-olt-zte service is reachable.
func (c *client) CheckHealth(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, http.MethodGet, "/health", nil, &result)
	return result, err
}

// CheckReadyz queries the readiness probe which includes per-OLT SNMP status.
func (c *client) CheckReadyz(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, http.MethodGet, "/readyz", nil, &result)
	return result, err
}

// GetUplinks retrieves detected line cards and uplink ports for the specified OLT.
func (c *client) GetUplinks(ctx context.Context, oltID string) (*domain.ZTEUplinksData, error) {
	path := "/api/v1/uplinks"
	if oltID != "" {
		path = fmt.Sprintf("/api/v1/olt/%s/uplinks", oltID)
	}

	var resp APIResponse[domain.ZTEUplinksData]
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetONUs retrieves all ONUs on the specified board slot and PON port.
func (c *client) GetONUs(ctx context.Context, oltID string, board int, pon int) ([]domain.ZTEONUInfo, error) {
	path := fmt.Sprintf("/api/v1/board/%d/pon/%d", board, pon)
	if oltID != "" {
		path = fmt.Sprintf("/api/v1/olt/%s/board/%d/pon/%d", oltID, board, pon)
	}

	var resp APIResponse[[]domain.ZTEONUInfo]
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetPaginatedONUs retrieves paginated ONUs on the specified board slot and PON port.
func (c *client) GetPaginatedONUs(ctx context.Context, oltID string, board int, pon int, page int, limit int) (*domain.ZTEPaginatedONUs, error) {
	path := fmt.Sprintf("/api/v1/paginate/board/%d/pon/%d", board, pon)
	if oltID != "" {
		path = fmt.Sprintf("/api/v1/olt/%s/paginate/board/%d/pon/%d", oltID, board, pon)
	}

	query := url.Values{}
	if page > 0 {
		query.Set("page", fmt.Sprintf("%d", page))
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}

	type paginatedResponse struct {
		Code   int                      `json:"code"`
		Status string                   `json:"status"`
		Data   []domain.ZTEONUInfo      `json:"data"`
		Meta   domain.ZTEPaginationMeta `json:"meta"`
	}

	var resp paginatedResponse
	if err := c.doRequest(ctx, http.MethodGet, path, query, &resp); err != nil {
		return nil, err
	}

	return &domain.ZTEPaginatedONUs{
		Data: resp.Data,
		Meta: resp.Meta,
	}, nil
}

// GetONUDetail retrieves deep telemetry data (Rx/Tx power, uptime, distance) for a single ONU.
func (c *client) GetONUDetail(ctx context.Context, oltID string, board int, pon int, onuID int) (*domain.ZTEONUDetail, error) {
	path := fmt.Sprintf("/api/v1/board/%d/pon/%d/onu/%d", board, pon, onuID)
	if oltID != "" {
		path = fmt.Sprintf("/api/v1/olt/%s/board/%d/pon/%d/onu/%d", oltID, board, pon, onuID)
	}

	var resp APIResponse[domain.ZTEONUDetail]
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetEmptyONUIDs retrieves available/unassigned ONU IDs on a PON port.
func (c *client) GetEmptyONUIDs(ctx context.Context, oltID string, board int, pon int) ([]int, error) {
	path := fmt.Sprintf("/api/v1/board/%d/pon/%d/onu_id/empty", board, pon)
	if oltID != "" {
		path = fmt.Sprintf("/api/v1/olt/%s/board/%d/pon/%d/onu_id/empty", oltID, board, pon)
	}

	var resp APIResponse[[]int]
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetONUSerials retrieves serial numbers mapped to ONU IDs.
func (c *client) GetONUSerials(ctx context.Context, oltID string, board int, pon int, noCache bool) ([]domain.ONUSerialInfo, error) {
	path := fmt.Sprintf("/api/v1/board/%d/pon/%d/onu_id_sn", board, pon)
	if oltID != "" {
		path = fmt.Sprintf("/api/v1/olt/%s/board/%d/pon/%d/onu_id_sn", oltID, board, pon)
	}

	query := url.Values{}
	if noCache {
		query.Set("nocache", "true")
	}

	var resp APIResponse[[]domain.ONUSerialInfo]
	if err := c.doRequest(ctx, http.MethodGet, path, query, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ClearCache forces invalidation of the Redis cache for a specific PON.
func (c *client) ClearCache(ctx context.Context, oltID string, board int, pon int) error {
	path := fmt.Sprintf("/api/v1/board/%d/pon/%d/cache/clear", board, pon)
	if oltID != "" {
		path = fmt.Sprintf("/api/v1/olt/%s/board/%d/pon/%d/cache/clear", oltID, board, pon)
	}

	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
