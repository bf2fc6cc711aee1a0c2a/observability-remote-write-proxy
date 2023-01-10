package main

import (
	"github.com/golang/snappy"
	"go.buf.build/protocolbuffers/go/prometheus/prometheus"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net/http"
)

func main() {
	handler := http.HandlerFunc(handleRequestSuccess)
	http.Handle("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handleRequestSuccess(w http.ResponseWriter, r *http.Request) {
	log.Println("request received")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	writeRequest, err := DecodeWriteRequest(r.Body)
	if err != nil {
		log.Printf("error receiving remote write request: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("there were %v time series in the request", len(writeRequest.Timeseries))
	w.WriteHeader(http.StatusNoContent)
	return
}

// DecodeWriteRequest deserialize a compressed remote write request
func DecodeWriteRequest(r io.Reader) (*prometheus.WriteRequest, error) {
	compressed, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		return nil, err
	}

	var req prometheus.WriteRequest
	if err := proto.Unmarshal(reqBuf, &req); err != nil {
		return nil, err
	}

	return &req, nil
}
