package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Initializer interface {
	Initialize(ctx context.Context, amountInKobo int64, email, reference string) (string, error)
}

type PaystackClient struct {
	baseURL            string
	PaystackSercretKey string
	http               *http.Client
}

func NewPaystackClient(paystackSecretKey string) Initializer {
	return &PaystackClient{baseURL: `https://api.paystack.co`, PaystackSercretKey: paystackSecretKey, http: &http.Client{Timeout: time.Second * 15}}
}

type InitializeRequest struct {
	Amount    int    `json:"amount"`
	Email     string `json:"email"`
	Currency  string `json:"currency"`
	Reference string `json:"reference"`
}

type InitializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Authorization_url string `json:"authorization_url"`
		Accesscode        string `json:"access_code"`
		Reference         string `json:"reference"`
	} `json:"data"`
}

var ErrResponseError = errors.New("error in paystack response")

func (p *PaystackClient) Initialize(ctx context.Context, amountInKobo int64, email, reference string) (string, error) {

	payload, err := json.Marshal(InitializeRequest{
		Amount:    int(amountInKobo),
		Email:     email,
		Currency:  "NGN",
		Reference: reference,
	})

	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+`/transaction/initialize`, bytes.NewReader(payload))

	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+p.PaystackSercretKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := p.http.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var out InitializeResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		slog.Error("decode paystack response failed", "error", err)
		return "", err
	}

	if res.StatusCode != http.StatusOK || !out.Status {
		slog.Error("paystack response status code not ok", "error", err)
		return "", ErrResponseError
	}

	return out.Data.Authorization_url, nil
}

func VerifySignature(secret string, body []byte, signature string) bool {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
