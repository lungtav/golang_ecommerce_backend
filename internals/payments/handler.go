package payments

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/middleware"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/response"
)

type handler struct {
	service Service
secret string

}

func NewHandler(service Service, secret string) *handler {
	return &handler{service: service, secret: secret}
}

func (h *handler) PayOrder(w http.ResponseWriter, r *http.Request) {

	orderID:= chi.URLParam(r, "orderId")

	userId, ok :=middleware.UserId(r.Context())
	if !ok{
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	url, err:=h.service.PayOrder(r.Context(), orderID, userId )
	if err!=nil{
		switch{
		case errors.Is(err, ErrNoOrder):
			response.Error(w, http.StatusConflict, "order doesnt exists")
			return
		case errors.Is(err, ErrNoUser):
			response.Error(w, http.StatusUnauthorized, "user doesnt exist")
			return
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	response.JSON(w, http.StatusOK, map[string]any{"url": url })

}

func (h *handler) Webhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	if !VerifySignature(h.secret, body, r.Header.Get("x-paystack-signature")) {
		response.Error(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	var ev webhookEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid payload")
		return
	}

	if ev.Event != "charge.success" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = h.service.HandleChargeSuccess(r.Context(), ChargeSuccess{
		ProviderID: ev.Data.ID,
		Reference:  ev.Data.Reference,
		Amount:     ev.Data.Amount,
		Currency:   ev.Data.Currency,
		Channel:    ev.Data.Channel,
	})
	if err != nil {
		slog.Error("handle charge success failed", "error", err)
		response.Error(w, http.StatusInternalServerError, "internal server error") // non-200 so Paystack retries
		return
	}

	w.WriteHeader(http.StatusOK)
}