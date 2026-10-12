package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data any) {

	b, err := json.Marshal(data)

	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func Error(w http.ResponseWriter, status int, message string) {

	errResponse := make(map[string]string)
	errResponse["error"] = message

	JSON(w, status, errResponse)
}
