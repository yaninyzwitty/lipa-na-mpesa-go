package daraja

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSimulateC2BUsesV2AndConfiguredShortCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/mpesa/c2b/v2/simulate" {
			t.Errorf("unexpected endpoint: %s %s", r.Method, r.URL.Path)
		}
		var request C2BSimulateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if request.ShortCode != "600996" {
			t.Errorf("shortcode = %q, want 600996", request.ShortCode)
		}
		_, _ = w.Write([]byte(`{"ResponseCode":"0","ResponseDescription":"Accepted"}`))
	}))
	defer server.Close()
	if err := SimulateC2B(context.Background(), server.Client(), Config{BaseURL: server.URL, ShortCode: "600996"}, "token"); err != nil {
		t.Fatal(err)
	}
}
