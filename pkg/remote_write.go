package pkg

import (
	"github.com/golang/snappy"
	"go.buf.build/protocolbuffers/go/prometheus/prometheus"
	"google.golang.org/protobuf/proto"
	"io"
)

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
