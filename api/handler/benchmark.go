package handler

import (
	"net/http"

	"bearlysocial-backend/util"
)

func Benchmark(w http.ResponseWriter, r *http.Request) {
	_, err := util.GenerateToken("john_doe@example.com")
	if err != nil {
		util.ReturnMessage(w, http.StatusInternalServerError, "Failed to generate token.")
		return
	}

	w.WriteHeader(http.StatusOK)
}
