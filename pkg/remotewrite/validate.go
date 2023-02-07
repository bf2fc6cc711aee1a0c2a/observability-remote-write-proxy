package remotewrite

import (
	"errors"
	"fmt"
	"go.buf.build/protocolbuffers/go/prometheus/prometheus"
	"log"
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
func ValidateRequest(remoteWriteRequest *prometheus.WriteRequest) (string, error) {
	log.Printf("there are %v time series in the request", len(remoteWriteRequest.Timeseries))

	clusterIDs := FindClusterIDs(remoteWriteRequest)
	if len(clusterIDs) > 1 {
		msg := fmt.Sprintf("request contains multiple cluster IDs: %v", clusterIDs)
		log.Printf(msg)
		return "", errors.New(msg)
	}

	for key, _ := range clusterIDs {
		return key, nil
	}

	return "", errors.New("request does not contain any cluster ids")
}
