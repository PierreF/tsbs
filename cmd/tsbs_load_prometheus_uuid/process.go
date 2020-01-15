package main

import (
	"github.com/golang/snappy"
	"github.com/timescale/tsbs/load"

	"bytes"
	"log"
	"net/http"
	"time"
)

const (
	errMarshal     = "marshal error: can't marshal write request: %v"
	errNewRequest  = "request error: can't create new request: %v"
	errSendRequest = "request error: can't send request: %v"
	errCloseBody   = "request error: can't close response body: %v"
)

const timeout = 30 * time.Second

type processor struct {
	client *http.Client
	url    string
}

func (p *processor) Init(workerNum int, _ bool) {
	client := &http.Client{
		Timeout: timeout,
	}

	p.client = client
	p.url = remoteURLs[workerNum%(len(remoteURLs))]
}

func (p *processor) ProcessBatch(b load.Batch, doLoad bool) (metricCount, rowCount uint64) {
	batch := b.(*batch)

	if !doLoad {
		return batch.metrics, batch.rows
	}

	marshal, err := batch.writeRequest.Marshal()

	if err != nil {
		log.Fatalf(errMarshal, err)
	}

	encoded := snappy.Encode(nil, marshal)
	request, err := http.NewRequest("POST", p.url, bytes.NewBuffer(encoded))

	if err != nil {
		log.Fatalf(errNewRequest, err)
	}

	request.Header.Set("Content-Encoding", "snappy")
	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("X-Prometheus-Remote-Write-Version", "2.0.0")

	response, err := p.client.Do(request)

	if err != nil {
		log.Fatalf(errSendRequest, err)
	}

	if err := response.Body.Close(); err != nil {
		log.Fatalf(errCloseBody, err)
	}

	return batch.metrics, batch.rows
}
