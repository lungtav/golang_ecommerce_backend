package request

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/response"
)

var validate = validator.New()

func Decode(w http.ResponseWriter, r *http.Request, dst any, limit int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)

	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			response.Error(w, http.StatusRequestEntityTooLarge, "request body too large")
			return false
		}
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return false
	}

	if err := validate.Struct(dst); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid input")
		return false
	}

	return true
}
