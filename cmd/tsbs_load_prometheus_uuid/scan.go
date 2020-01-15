package main

import (
	"crypto/sha1"
	"io"
	"sort"

	gouuid "github.com/gofrs/uuid"
	"github.com/prometheus/prometheus/prompb"
	"github.com/timescale/tsbs/load"

	"bufio"
	"log"
	"strconv"
	"strings"
)

const (
	errNotThreeTuplesFmt = "parse error: line does not have 3 tuples, has %d"
	errNotInteger        = "parse error: value is not an integer: %v"
	errNotFloat          = "parse error: value is not a float: %v"
)

type decoder struct {
	scanner *bufio.Scanner
}

func (d *decoder) Decode(_ *bufio.Reader) *load.Point {
	ok := d.scanner.Scan()

	if !ok && d.scanner.Err() == nil { // nothing scanned & no error = EOF
		return nil
	} else if !ok {
		log.Fatalf("scan error: %v", d.scanner.Err())
	}

	return load.NewPoint(d.scanner.Bytes())
}

type batch struct {
	writeRequest prompb.WriteRequest
	rows         uint64
	metrics      uint64
}

func (b *batch) Len() int {
	return int(b.rows)
}

func (b *batch) Append(item *load.Point) {
	that := item.Data.([]byte)
	thatStr := string(that)

	args := strings.Split(thatStr, " ")

	if len(args) != 3 {
		log.Fatalf(errNotThreeTuplesFmt, len(args))
	}

	timestamp, err := strconv.ParseInt(args[2], 10, 64)

	if err != nil {
		log.Fatalf(errNotInteger, err)
	}

	// Convert InfluxDB tags to Prometheus labels
	tags := strings.Split(args[0], ",")
	measurement, tags := tags[0], tags[1:]
	commonPromLabels := make([]*prompb.Label, 0, len(tags))

	sort.Strings(tags)
	for _, tagStr := range tags {
		tag := strings.Split(tagStr, "=")
		tagKey := tag[0]
		tagValue := tag[1]

		commonPromLabel := &prompb.Label{
			Name:  tagKey,
			Value: tagValue,
		}

		commonPromLabels = append(commonPromLabels, commonPromLabel)
	}

	// Convert InfluxDB field-value pairs to Prometheus '__name__' label and associated sample
	fieldValuePairs := strings.Split(args[1], ",")

	for _, fieldValuePairStr := range fieldValuePairs {
		fieldValuePair := strings.Split(fieldValuePairStr, "=")
		field := fieldValuePair[0]
		promLabels := make([]*prompb.Label, 0, len(commonPromLabels))

		promLabel := &prompb.Label{
			Name:  "__name__",
			Value: measurement + "_" + field,
		}

		promLabels = append(promLabels, promLabel)
		promLabels = append(promLabels, commonPromLabels...)

		h := sha1.New()
		for _, l := range promLabels {
			io.WriteString(h, l.GetName())
			io.WriteString(h, l.GetValue())
		}
		hashValue := h.Sum(nil)
		var metricUUID gouuid.UUID
		copy(metricUUID[:], hashValue)
		
		promLabel = &prompb.Label{
			Name:  "__uuid__",
			Value: metricUUID.String(),
		}
		// comment next line for tsbs_load_prometheus_uuid2
		promLabels = append(promLabels, promLabel)
		// uncomment next line for tsbs_load_prometheus_uuid3
		// promLabels = []*prompb.Label{promLabel}
		// uncomment next lines for tsbs_load_prometheus_uuid4
		/*promLabel = &prompb.Label{
			Name:  "__bleemeo_uuid__",
			Value: metricUUID.String(),
		}
		promLabels = []*prompb.Label{promLabel}*/

		valueStr := strings.Replace(fieldValuePair[1], "i", "", 1) // Remove appended 'i' specific to InfluxDB
		value, err := strconv.ParseFloat(valueStr, 64)

		if err != nil {
			log.Fatalf(errNotFloat, err)
		}

		promSamples := []prompb.Sample{
			{
				Value:     value,
				Timestamp: timestamp / 1000000, // Convert nanoseconds to milliseconds
			},
		}

		series := &prompb.TimeSeries{
			Labels:  promLabels,
			Samples: promSamples,
		}

		b.writeRequest.Timeseries = append(b.writeRequest.Timeseries, series)

		b.metrics++
	}

	b.rows++
}

type factory struct{}

func (f *factory) New() load.Batch {
	return &batch{}
}
