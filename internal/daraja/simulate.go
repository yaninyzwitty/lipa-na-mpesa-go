package daraja

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type C2BSimulateRequest struct {
	ShortCode     string `json:"ShortCode"`
	CommandID     string `json:"CommandID"`
	Amount        int64  `json:"Amount"`
	Msisdn        string `json:"Msisdn"`
	BillRefNumber string `json:"BillRefNumber"`
}

func SimulateC2B(
	ctx context.Context,
	client *http.Client,
	config Config,
	accessToken string,
) error {
	payload := C2BSimulateRequest{
		ShortCode:     config.ShortCode,
		CommandID:     "CustomerPayBillOnline",
		Amount:        100,
		Msisdn:        "254708374149",
		BillRefNumber: "MEM-001",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal C2B simulation: %w", err)
	}

	if config.BaseURL == "" || config.ShortCode == "" {
		return fmt.Errorf("missing C2B simulation configuration")
	}
	endpoint := config.BaseURL + "/mpesa/c2b/v2/simulate"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("simulate C2B payment: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(
		io.LimitReader(resp.Body, 1<<20),
	)
	if err != nil {
		return err
	}

	slog.Debug("Daraja C2B simulation response", "http_status", resp.StatusCode, "body", string(responseBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf(
			"C2B simulation failed (HTTP %d): %s",
			resp.StatusCode,
			responseBody,
		)
	}
	var result struct {
		ResponseCode        string `json:"ResponseCode"`
		ResponseDescription string `json:"ResponseDescription"`
	}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return fmt.Errorf("decode C2B simulation response: %w", err)
	}
	if result.ResponseCode != "0" && result.ResponseCode != "00000000" {
		return fmt.Errorf("C2B simulation rejected: %s", result.ResponseDescription)
	}

	return nil
}
