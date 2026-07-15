package paystack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	secretKey  string
	publicKey  string
	httpClient *http.Client
}

func NewClient(secretKey, publicKey string) *Client {
	return &Client{
		secretKey: secretKey,
		publicKey: publicKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type InitializeRequest struct {
	Email     string `json:"email"`
	Amount    int    `json:"amount"` // Amount in lowest denomination (e.g. kobo)
	Reference string `json:"reference"`
}

type InitializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    *struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// InitializeTransaction creates a new checkout session on Paystack.
func (c *Client) InitializeTransaction(ctx context.Context, req InitializeRequest) (*InitializeResponse, error) {
	if c.secretKey == "" {
		return nil, fmt.Errorf("paystack secret key is missing")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.paystack.co/transaction/initialize", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("paystack api request failed: %w", err)
	}
	defer resp.Body.Close()

	var psResp InitializeResponse
	if err := json.NewDecoder(resp.Body).Decode(&psResp); err != nil {
		return nil, fmt.Errorf("paystack decode failed: %w", err)
	}

	if !psResp.Status {
		return nil, fmt.Errorf("paystack error: %s", psResp.Message)
	}

	return &psResp, nil
}
