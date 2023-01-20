package remotewrite

import "go.buf.build/protocolbuffers/go/prometheus/prometheus"

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
