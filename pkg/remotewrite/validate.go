package remotewrite

import (
	"bytes"
	"errors"
	"github.com/golang/snappy"
	"go.buf.build/protocolbuffers/go/prometheus/prometheus"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net/http"
)

const (
	clusterIDLabelName string = "cluster_id"
)

// FindClusterIDs returns a list of cluster IDs found in the request
func FindClusterIDs(req *prometheus.WriteRequest) map[string]int {
	clusterID := make(map[string]int)
	for _, ts := range req.Timeseries {
		for _, ls := range ts.Labels {
			if ls.Name == clusterIDLabelName {
				clusterID[ls.Value] += 1
			}
		}
	}
	return clusterID
}

// ValidateRequest validates a remote write request
func ValidateRequest(r *http.Request) error {
	log.Println("request received")

	if r.Method != http.MethodPost {
		return errors.New("not found")
	}

	writeRequest, err := DecodeWriteRequest(r.Body)
	if err != nil {
		log.Printf("error receiving remote write request: %v", err.Error())
		return errors.New("error receiving remote write request")
	}

	data, _ := proto.Marshal(writeRequest)
	encoded := snappy.Encode(nil, data)
	body := bytes.NewReader(encoded)
	r.Body = io.NopCloser(body)

	log.Printf("there are %v time series in the request", len(writeRequest.Timeseries))

	clusterIDs := FindClusterIDs(writeRequest)
	if len(clusterIDs) > 1 {
		log.Printf("request contains multiple cluster IDs: %v", clusterIDs)
		return errors.New("request contains multiple cluster IDs")
	}

	return nil
}
