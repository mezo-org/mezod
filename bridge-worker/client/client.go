package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mezo-org/mezod/bridge-worker/types"
)

var (
	submitSignatureEndpoint = func(baseURL string) string {
		return fmt.Sprintf("%s/submit-signature", baseURL)
	}
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(url string) *Client {
	return &Client{
		baseURL:    url,
		httpClient: &http.Client{},
	}
}

func (c *Client) SubmitSignature(attestation *types.AssetsUnlocked, signature string) error {
	request := &types.SubmitSignatureRequest{
		Entry:     attestation,
		Signature: signature,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(submitSignatureEndpoint(c.baseURL), "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response types.SubmitSignatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return errors.New(response.Error)
	}

	return nil
}
