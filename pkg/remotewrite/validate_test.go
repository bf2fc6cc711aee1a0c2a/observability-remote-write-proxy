package remotewrite

import (
	"github.com/onsi/gomega"
	"go.buf.build/protocolbuffers/go/prometheus/prometheus"
	"testing"
)

var (
	reqWithMultipleClusterIDMultipleRequests = &prometheus.WriteRequest{
		Timeseries: []*prometheus.TimeSeries{
			{
				Labels: []*prometheus.Label{
					{
						Name:  "cluster_id",
						Value: "cluster-1",
					},
					{
						Name:  "cluster_id",
						Value: "cluster-2",
					},
					{
						Name:  "cluster_id",
						Value: "cluster-2",
					},
				},
			},
		},
	}
	reqWithMultipleClusterID = &prometheus.WriteRequest{
		Timeseries: []*prometheus.TimeSeries{
			{
				Labels: []*prometheus.Label{
					{
						Name:  "cluster_id",
						Value: "cluster-1",
					},
					{
						Name:  "cluster_id",
						Value: "cluster-2",
					},
				},
			},
		},
	}
	reqWithOneClusterID = &prometheus.WriteRequest{
		Timeseries: []*prometheus.TimeSeries{
			{
				Labels: []*prometheus.Label{
					{
						Name:  "cluster_id",
						Value: "cluster-1",
					},
				},
			},
		},
	}
	req = &prometheus.WriteRequest{
		Timeseries: []*prometheus.TimeSeries{},
	}
)

func TestFindClusterIDs(t *testing.T) {
	type args struct {
		request *prometheus.WriteRequest
	}

	tests := []struct {
		name string
		args args
		want map[string]int
	}{
		{
			name: "should return a map that contains multiple cluster IDs, where one is sending multiple requests",
			args: args{
				request: reqWithMultipleClusterIDMultipleRequests,
			},
			want: map[string]int{
				"cluster-1": 1,
				"cluster-2": 2,
			},
		},
		{
			name: "should return a map that contains multiple cluster IDs",
			args: args{
				request: reqWithMultipleClusterID,
			},
			want: map[string]int{
				"cluster-1": 1,
				"cluster-2": 1,
			},
		},
		{
			name: "should return a map that contains one cluster ID",
			args: args{
				request: reqWithOneClusterID,
			},
			want: map[string]int{
				"cluster-1": 1,
			},
		},
		{
			name: "should return an empty map if there is no cluster ID in the request",
			args: args{
				request: req,
			},
			want: map[string]int{},
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := gomega.NewWithT(t)
			g.Expect(FindClusterIDs(tt.args.request)).To(gomega.Equal(tt.want))
		})
	}
}
