package daraja

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

func GetAccessToken(ctx context.Context, client *http.Client, config Config) (string, error) {
	if config.BaseURL == "" || config.ConsumerKey == "" || config.ConsumerSecret == "" {
		return "", fmt.Errorf("missing Daraja OAuth configuration")
	}

	endpoint := config.BaseURL + "/oauth/v1/generate"

	params := url.Values{}
	params.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint+"?"+params.Encode(),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("create OAuth request: %w", err)
	}

	// Uses HTTP Basic Authentication.
	// Go handles Base64 encoding automatically.
	req.SetBasicAuth(config.ConsumerKey, config.ConsumerSecret)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request OAuth token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if err != nil {
			return "", fmt.Errorf("Daraja OAuth failed with HTTP %d (read error response: %w)", resp.StatusCode, err)
		}
		var apiError struct {
			RequestID    string `json:"requestId"`
			ErrorCode    string `json:"errorCode"`
			ErrorMessage string `json:"errorMessage"`
		}
		if json.Unmarshal(body, &apiError) == nil && (apiError.ErrorCode != "" || apiError.ErrorMessage != "") {
			return "", fmt.Errorf("Daraja OAuth failed with HTTP %d (code %s, request ID %s): %s", resp.StatusCode, apiError.ErrorCode, apiError.RequestID, apiError.ErrorMessage)
		}
		return "", fmt.Errorf(
			"Daraja OAuth failed with HTTP %d (unrecognized error response)",
			resp.StatusCode,
		)
	}

	var result oauthResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode OAuth response: %w", err)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("Daraja returned an empty access token")
	}

	return result.AccessToken, nil
}

func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
	}
}
