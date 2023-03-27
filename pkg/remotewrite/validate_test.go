package remotewrite

import (
	"testing"

	"github.com/onsi/gomega"
	"go.buf.build/protocolbuffers/go/prometheus/prometheus"
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

func TestValidateRequest(t *testing.T) {
	type args struct {
		remoteWriteRequest *prometheus.WriteRequest
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should return an error if there are multiple cluster IDs in the request",
			args: args{
				remoteWriteRequest: reqWithMultipleClusterID,
			},
			wantErr: true,
		},
		{
			name: "should return an error if there are no cluster IDs in the request",
			args: args{
				remoteWriteRequest: req,
			},
			wantErr: true,
		},
		{
			name: "should succeed if there is one cluster ID in the request",
			args: args{
				remoteWriteRequest: reqWithOneClusterID,
			},
			want:    "cluster-1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateRequest(tt.args.remoteWriteRequest)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMetadataRequest(t *testing.T) {
	metadataReq := &prometheus.WriteRequest{
		Metadata: []*prometheus.MetricMetadata{
			{
				MetricFamilyName: "test",
			},
		},
	}
	metadataAndTimeseriesReq := &prometheus.WriteRequest{
		Metadata: []*prometheus.MetricMetadata{
			{
				MetricFamilyName: "test",
			},
		},
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
	type args struct {
		remoteWriteRequest *prometheus.WriteRequest
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should return false if the request is not a metadata request",
			args: args{
				remoteWriteRequest: reqWithOneClusterID,
			},
			want: false,
		},
		{
			name: "should return true if the request is a metadata request",
			args: args{
				remoteWriteRequest: metadataReq,
			},
			want: true,
		},
		{
			name: "should return false if the request is a metadata request and a timeseries request",
			args: args{
				remoteWriteRequest: metadataAndTimeseriesReq,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMetadataRequest(tt.args.remoteWriteRequest); got != tt.want {
				t.Errorf("IsMetadataRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
