package daraja

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetAccessTokenReportsDarajaError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, secret, ok := r.BasicAuth()
		if !ok || key != "key" || secret != "secret" || r.URL.Path != "/oauth/v1/generate" || r.URL.Query().Get("grant_type") != "client_credentials" {
			t.Errorf("unexpected OAuth request")
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"requestId":"abc","errorCode":"403.001","errorMessage":"Forbidden"}`))
	}))
	defer server.Close()
	_, err := GetAccessToken(context.Background(), server.Client(), Config{BaseURL: server.URL, ConsumerKey: "key", ConsumerSecret: "secret"})
	if err == nil || !strings.Contains(err.Error(), "403.001") || !strings.Contains(err.Error(), "request ID abc") {
		t.Fatalf("unexpected error: %v", err)
	}
}
