package daraja

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterC2BURLs(t *testing.T) {
	for _, tc := range []struct {
		name, status, body, want string
	}{
		{"success", "200", `{"ResponseCode":"0","ResponseDescription":"Success"}`, ""},
		{"v2 success", "200", `{"ResponseCode":"00000000","ResponseDescription":"Success"}`, ""},
		{"service unavailable", "500", `{"requestId":"abc","errorCode":"500.003.1001","errorMessage":"Service is currently unreachable."}`, "request ID abc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/mpesa/c2b/v2/registerurl" || r.Header.Get("Authorization") != "Bearer token" {
					t.Errorf("unexpected request: %s %s", r.URL.Path, r.Header.Get("Authorization"))
				}
				if tc.status == "500" {
					w.WriteHeader(http.StatusInternalServerError)
				}
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			config := Config{BaseURL: server.URL, ShortCode: "600984", ResponseType: "Completed", ConfirmationURL: server.URL + "/confirmation", ValidationURL: server.URL + "/validation"}
			err := RegisterC2BURLs(context.Background(), server.Client(), config, "token")
			if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("got error %v, want substring %q", err, tc.want)
			}
		})
	}
}
