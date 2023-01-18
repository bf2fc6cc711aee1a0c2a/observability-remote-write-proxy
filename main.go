package main

import (
	"log"
	"net/http"
	"observability-remote-write-proxy/pkg"
)

func main() {
	handler := http.HandlerFunc(handleRequestSuccess)
	http.Handle("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

func handleRequestSuccess(w http.ResponseWriter, r *http.Request) {
	log.Println("request received")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	writeRequest, err := pkg.DecodeWriteRequest(r.Body)
	if err != nil {
		log.Printf("error receiving remote write request: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	clusterIDs := pkg.ValidateClusterIDs(writeRequest)
	if len(clusterIDs) > 1 {
		log.Printf("request contains multiple cluster IDs: %v", clusterIDs)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("there were %v time series in the request", len(writeRequest.Timeseries))
	w.WriteHeader(http.StatusNoContent)
	return
}
