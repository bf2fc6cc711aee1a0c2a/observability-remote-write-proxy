FROM registry.access.redhat.com/ubi9/go-toolset:1.18 as builder

USER root

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY main.go main.go
COPY pkg/ pkg/
COPY api/ api/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o proxy main.go

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.1

WORKDIR /
COPY --from=builder /workspace/proxy .

ENTRYPOINT ["/proxy"]