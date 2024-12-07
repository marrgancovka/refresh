package responser

import (
	"encoding/json"
	"net/http"
)

type MessageResponse struct {
	Msg string `json:"msg"`
}

func Send200(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(v)
	if err != nil {
		return
	}
	_, _ = w.Write(resp)
}

func Send400(w http.ResponseWriter, msg string) {
	resp, err := json.Marshal(MessageResponse{msg})
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write(resp)
}

func Send401(w http.ResponseWriter, msg string) {
	resp, err := json.Marshal(MessageResponse{msg})
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write(resp)
}

func Send500(w http.ResponseWriter) {
	resp, err := json.Marshal(MessageResponse{"internal server error"})
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(resp)
}
