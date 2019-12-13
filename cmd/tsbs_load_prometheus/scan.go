package main

import (
	"github.com/prometheus/prometheus/prompb"
	"github.com/timescale/tsbs/load"

	"bufio"
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
		fatal("scan error: %v", d.scanner.Err())
		return nil
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
	b.rows++

	args := strings.Split(thatStr, " ")
	if len(args) != 3 {
		fatal(errNotThreeTuplesFmt, len(args))
		return
	}

	timestamp, err := strconv.ParseInt(args[2], 10, 64)

	if err != nil {
		fatal(errNotInteger, err)
	}

	commonLabels := strings.Split(args[0], ",")
	commonLabelPrefix, commonLabels := commonLabels[0], commonLabels[1:]
	commonPromLabels := make([]*prompb.Label, 0, len(commonLabels))

	for _, commonLabelStr := range commonLabels {
		commonLabel := strings.Split(commonLabelStr, "=")
		commonLabelName := commonLabel[0]
		commonLabelValue := commonLabel[1]

		commonPromLabel := &prompb.Label{
			Name:  commonLabelName,
			Value: commonLabelValue,
		}

		commonPromLabels = append(commonPromLabels, commonPromLabel)
	}

	labelValues := strings.Split(args[1], ",")

	for _, metricStr := range labelValues {
		promLabels := make([]*prompb.Label, len(commonPromLabels))

		copy(promLabels, commonPromLabels)

		labelValue := strings.Split(metricStr, "=")
		label := labelValue[0]

		promLabel := &prompb.Label{
			Name:  "__name__",
			Value: commonLabelPrefix + "_" + label,
		}

		promLabels = append(promLabels, promLabel)

		valueClean := strings.ReplaceAll(labelValue[1], "i", "")
		value, err := strconv.ParseFloat(valueClean, 64)

		if err != nil {
			fatal(errNotFloat, err)
		}

		promSamples := []prompb.Sample{
			{
				Value:     value,
				Timestamp: timestamp / 1000000,
			},
		}

		series := &prompb.TimeSeries{
			Labels:  promLabels,
			Samples: promSamples,
		}

		b.writeRequest.Timeseries = append(b.writeRequest.Timeseries, series)

		b.metrics++
	}
}

type factory struct{}

func (f *factory) New() load.Batch {
	return &batch{}
}
