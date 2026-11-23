package daraja

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type registerURLRequest struct {
	ShortCode       string `json:"ShortCode"`
	ResponseType    string `json:"ResponseType"`
	ConfirmationURL string `json:"ConfirmationURL"`
	ValidationURL   string `json:"ValidationURL"`
}

func RegisterC2BURLs(ctx context.Context, client *http.Client, config Config, token string) error {
	if config.BaseURL == "" || config.ShortCode == "" || config.ConfirmationURL == "" || config.ValidationURL == "" {
		return fmt.Errorf("missing C2B configuration")
	}
	if config.ResponseType != "Completed" && config.ResponseType != "Cancelled" {
		return fmt.Errorf("invalid C2B response type %q: use Completed or Cancelled", config.ResponseType)
	}
	for name, value := range map[string]string{"confirmation": config.ConfirmationURL, "validation": config.ValidationURL} {
		parsed, err := url.Parse(value)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return fmt.Errorf("invalid C2B %s URL %q", name, value)
		}
	}
	payload := registerURLRequest{config.ShortCode, config.ResponseType, config.ConfirmationURL, config.ValidationURL}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode C2B registration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.BaseURL+"/mpesa/c2b/v2/registerurl", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create C2B registration request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("register C2B URLs: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read C2B registration response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiError struct {
			RequestID    string `json:"requestId"`
			ErrorCode    string `json:"errorCode"`
			ErrorMessage string `json:"errorMessage"`
		}
		if json.Unmarshal(responseBody, &apiError) == nil && apiError.ErrorCode != "" {
			return fmt.Errorf("registration failed (HTTP %d, code %s, request ID %s): %s", resp.StatusCode, apiError.ErrorCode, apiError.RequestID, apiError.ErrorMessage)
		}
		return fmt.Errorf("registration failed (HTTP %d): %s", resp.StatusCode, responseBody)
	}
	var result struct {
		ResponseCode        string `json:"ResponseCode"`
		ResponseDescription string `json:"ResponseDescription"`
	}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return fmt.Errorf("decode C2B registration response: %w", err)
	}

	slog.Debug("Daraja C2B registration response", "http_status", resp.StatusCode, "body", string(responseBody))
	if result.ResponseCode != "0" && result.ResponseCode != "00000000" {
		return fmt.Errorf("registration rejected: %s", result.ResponseDescription)
	}
	return nil
}
