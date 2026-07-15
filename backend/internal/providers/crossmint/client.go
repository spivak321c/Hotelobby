package crossmint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	apiKey     string
	projectID  string
	envBaseURL string
	httpClient *http.Client
}

func NewClient(apiKey, projectID string) *Client {
	// For MVP, we can hardcode staging environment. In production, this should be configurable.
	return &Client{
		apiKey:     apiKey,
		projectID:  projectID,
		envBaseURL: "https://staging.crossmint.com/api/2022-06-09",
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type OrderRequest struct {
	Recipient struct {
		Email string `json:"email"`
	} `json:"recipient"`
	Payment struct {
		Method       string `json:"method"`
		Currency     string `json:"currency"`
		ReceiptEmail string `json:"receiptEmail,omitempty"`
	} `json:"payment"`
	LineItems []LineItem `json:"lineItems"`
}

type LineItem struct {
	CollectionLocator string `json:"collectionLocator"`
	CallData          struct {
		TotalPrice string `json:"totalPrice"`
		Quantity   int    `json:"quantity"`
	} `json:"callData"`
}

type OrderResponse struct {
	OrderID   string `json:"orderId"`
	Phase     string `json:"phase"`
	ClientSecret string `json:"clientSecret"`
}

// CreateOrder initiates a Crossmint headless checkout or standard order.
func (c *Client) CreateOrder(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("crossmint api key is missing")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal crossmint request: %w", err)
	}

	url := fmt.Sprintf("%s/orders", c.envBaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create crossmint request: %w", err)
	}

	httpReq.Header.Set("X-API-KEY", c.apiKey)
	if c.projectID != "" {
		httpReq.Header.Set("X-PROJECT-ID", c.projectID)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("crossmint api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("crossmint api returned status %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode crossmint response: %w", err)
	}

	// Crossmint docs state order details are inside `order` root field in some endpoints
	// but direct for headless /orders
	var orderResp OrderResponse
	
	if orderData, ok := data["order"].(map[string]interface{}); ok {
		// nested
		if id, ok := orderData["orderId"].(string); ok { orderResp.OrderID = id }
		if p, ok := orderData["phase"].(string); ok { orderResp.Phase = p }
	} else {
		// flat
		if id, ok := data["orderId"].(string); ok { orderResp.OrderID = id }
		if p, ok := data["phase"].(string); ok { orderResp.Phase = p }
		if cs, ok := data["clientSecret"].(string); ok { orderResp.ClientSecret = cs }
	}

	if orderResp.OrderID == "" {
		return nil, fmt.Errorf("crossmint api returned unexpected format, missing orderId")
	}

	return &orderResp, nil
}
