package c2b

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type transaction struct {
	TransactionType   string `json:"TransactionType"`
	TransID           string `json:"TransID"`
	TransTime         string `json:"TransTime"`
	TransAmount       string `json:"TransAmount"`
	BusinessShortCode string `json:"BusinessShortCode"`
	BillRefNumber     string `json:"BillRefNumber"`
	InvoiceNumber     string `json:"InvoiceNumber"`
	OrgAccountBalance string `json:"OrgAccountBalance"`
	ThirdPartyTransID string `json:"ThirdPartyTransID"`
	MSISDN            string `json:"MSISDN"`
	FirstName         string `json:"FirstName"`
	MiddleName        string `json:"MiddleName"`
	LastName          string `json:"LastName"`
}

func HandleValidation(w http.ResponseWriter, r *http.Request) {
	tx, ok := decodeTransaction(w, r)
	if !ok {
		return
	}
	slog.Info("C2B validation", "account", tx.BillRefNumber, "amount", tx.TransAmount)
	accept(w)
}

func HandleConfirmation(w http.ResponseWriter, r *http.Request) {
	tx, ok := decodeTransaction(w, r)
	if !ok {
		return
	}
	slog.Info("C2B payment received", "transID", tx.TransID, "account", tx.BillRefNumber, "amount", tx.TransAmount)
	accept(w)
}

func decodeTransaction(w http.ResponseWriter, r *http.Request) (transaction, bool) {
	var tx transaction
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return tx, false
	}
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return tx, false
	}
	return tx, true
}

func accept(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		ResultCode int    `json:"ResultCode"`
		ResultDesc string `json:"ResultDesc"`
	}{0, "Accepted"})
}
