package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	handler := http.HandlerFunc(handleRequestSuccess)
	http.Handle("/test", handler)
	http.ListenAndServe(":8080", nil)
}

func handleRequestSuccess(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Status OK"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Err: %s", err)
	}
	w.Write(jsonResp)
	return
}
