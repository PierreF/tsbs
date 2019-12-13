package main

import (
	"github.com/golang/snappy"
	"github.com/timescale/tsbs/load"

	"bytes"
	"net/http"
)

const (
	errMarshal     = "marshal error: can't marshal write request: %v"
	errNewRequest  = "request error: can't create new request: %v"
	errSendRequest = "request error: can't send request: %v"
	errCloseBody   = "request error: can't close response body"
)

type processor struct {
	url string
}

func (p *processor) Init(workerNum int, _ bool) {
	p.url = remoteURLs[workerNum%(len(remoteURLs))]
}

func (p *processor) ProcessBatch(b load.Batch, doLoad bool) (metricCount, rowCount uint64) {
	batch := b.(*batch)

	if !doLoad {
		return batch.metrics, batch.rows
	}

	marshal, err := batch.writeRequest.Marshal()

	if err != nil {
		fatal(errMarshal, err)
	}

	encoded := snappy.Encode(nil, marshal)
	request, err := http.NewRequest("POST", p.url+"/write", bytes.NewBuffer(encoded))

	if err != nil {
		fatal(errNewRequest, err)
	}

	request.Header.Set("Content-Encoding", "snappy")
	request.Header.Set("Content-Type", "application/x-protobuf")

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		fatal(errSendRequest, err)
	}

	if err := response.Body.Close(); err != nil {
		fatal(errCloseBody, err)
	}

	return batch.metrics, batch.rows
}
