package stores

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	mw "github.com/lungtav/golang_ecommerce_backend/internals/platform/middleware"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/request"
	"github.com/lungtav/golang_ecommerce_backend/internals/platform/response"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateStore(w http.ResponseWriter, r *http.Request) {
	var req StoreRequest
	if !request.Decode(w, r, &req, 1<<10){
		return
	}
	userId, ok:= mw.UserId(r.Context())

	if !ok{
		response.Error(w, http.StatusUnauthorized, "user id not found")
		return
	}

	store, err:=h.service.CreateStore(r.Context(), req, userId)

	if err!=nil{
		switch{
		case errors.Is(err, ErrConflictError):
		response.Error(w, http.StatusConflict, "a store with this name exists")
		return
			
		default:
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
		}
	}

	response.JSON(w, http.StatusCreated, store)
}

func  (h *handler) UpdateStore(w http.ResponseWriter, r *http.Request) {
	storeId:=chi.URLParam(r, "storeId")

	if storeId ==""{
		response.Error(w, http.StatusBadRequest, "store doesn't exist")
		return
	}

	var req StoreRequest
	if !request.Decode(w, r, &req, 1<<10 ){
		return
	}

	if err:=h.service.UpdateStore(r.Context(), req, storeId); err!=nil{
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "store updated successfully"})
}