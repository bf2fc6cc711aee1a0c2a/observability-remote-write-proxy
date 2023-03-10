package remotewrite

import (
	"io"
	"net/http"
	"testing"

	"github.com/golang/snappy"
	"github.com/onsi/gomega"
	"go.buf.build/protocolbuffers/go/prometheus/prometheus"
	"google.golang.org/protobuf/proto"
)

func TestDecodeWriteRequest(t *testing.T) {
	rw := &prometheus.WriteRequest{
		Timeseries: []*prometheus.TimeSeries{
			{
				Labels: []*prometheus.Label{
					{
						Name:  "test",
						Value: "test",
					},
				},
				Samples: []*prometheus.Sample{
					{
						Value:     1,
						Timestamp: 1,
					},
				},
			},
		},
	}
	r := &http.Request{}
	g := gomega.NewGomegaWithT(t)
	err := PopulateRequestBody(rw, r)
	g.Expect(err).To(gomega.BeNil())

	req, err := DecodeWriteRequest(r)
	g.Expect(err).To(gomega.BeNil())

	g.Expect(proto.Equal(rw, req)).To(gomega.BeTrue())

	// Modify the HTTP request to have an invalid body
	r.Body = io.NopCloser(io.MultiReader(r.Body, r.Body))
	writeReq, err := DecodeWriteRequest(r)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(writeReq).To(gomega.BeNil())
}

func TestPopulateRequestBody(t *testing.T) {
	rw := &prometheus.WriteRequest{
		Timeseries: []*prometheus.TimeSeries{
			{
				Labels: []*prometheus.Label{
					{
						Name:  "test",
						Value: "test",
					},
				},
				Samples: []*prometheus.Sample{
					{
						Value:     1,
						Timestamp: 1,
					},
				},
			},
		},
	}
	r := &http.Request{}
	g := gomega.NewGomegaWithT(t)
	err := PopulateRequestBody(rw, r)
	g.Expect(err).To(gomega.BeNil())

	compressed, err := io.ReadAll(r.Body)
	g.Expect(err).To(gomega.BeNil())

	reqBuf, err := snappy.Decode(nil, compressed)
	g.Expect(err).To(gomega.BeNil())

	var req prometheus.WriteRequest
	err = proto.Unmarshal(reqBuf, &req)
	g.Expect(err).To(gomega.BeNil())

	g.Expect(proto.Equal(rw, &req)).To(gomega.BeTrue())
}
