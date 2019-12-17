# TSBS Supplemental Guide: Prometheus Remote Storage

**Prometheus Remote Storage support requires [VictoriaMetrics support][victoriametrics-pull-request].**

[Prometheus](https://github.com/prometheus/prometheus) is a monitoring system and time series database.
Prometheus offers a set of interfaces that allow integrating with remote storage systems
(see [Remote Storage integrations](https://prometheus.io/docs/prometheus/latest/storage/#remote-storage-integrations)).
This additional guide explains how the data generated for the TSBS is stored,
additional flags available when using the data importer (`tsbs_load_prometheus`),
and additional flags available when using the query runner (`tsbs_run_run_run_queries_victoriametrics`).

To install all the necessary tools, follow the steps below:
```
$ cd $GOPATH/src/github.com/timescale/tsbs/cmd
$ cd tsbs_generate_data && go install
$ cd ../tsbs_generate_queries && go install
$ cd ../tsbs_load_prometheus && go install
$ cd ../tsbs_run_queries_victoriametrics && go install
```

**Read the main README before you continue.**

## Generating data

Prometheus Remote Storage use the same format as for InfluxDB.
This is "pseudo-CSV" format, each reading is composed of a single line
where the name of the table is the first item, a comma,
followed by several comma-separated items of tags in the format
of `<label>=<value>`, a space, several comma-separated items of fields
in the format of `<label>=<value>`, a space, and finally the timestamp
for the reading.

One of the ways to generate data for insertion is to use `scripts/generate_data.sh`:
```text
FORMATS=influx SCALE=100 TS_START=2019-12-01T00:00:00Z TS_END=2019-12-15T00:00:00Z ./scripts/generate_data.sh
```

## `tsbs_load_prometheus`

One of the ways to load data in Prometheus Remote Storage is to use `scripts/load_prometheus.sh`:
```text
DATA_FILE=/tmp/bulk_data/data_influx_cpu-only_100_2019-12-01T00:00:00Z_2019-12-15T00:00:00Z_10s_123.dat.gz \
REMOTE_URL=http://localhost:9201/write ./scripts/load_prometheus.sh
```
> Assumed that Prometheus Remote Storage is already installed and ready for insertion on the `REMOTE_URL` url.

### Additional Flags

#### `-urls` (type: `string`, default: `http://localhost:9201/write`)

Comma-separated list of URLs to connect to for inserting data. It can be
just a single-version URL or list of VMInsert URLs. Workers will be
distributed in a round robin fashion across the URLs.

## Generating queries

**Requires [VictoriaMetrics support][victoriametrics-pull-request].**

This section uses the queries generation tool for VictoriaMetrics 
(see [Generating queries](https://github.com/timescale/tsbs/blob/46f535f7d36c41ed3091ff050c63c7da84d0ddd8/docs/victoriametrics.md#generating-queries))

One of the ways to generate queries for Prometheus Remote Storage is to use `scripts/generate_queries.sh`:
```text
FORMATS=victoriametrics SCALE=100 TS_START=2019-12-01T00:00:00Z TS_END=2019-12-15T00:00:00Z \
QUERY_TYPES="cpu-max-all-8" ./scripts/generate_queries.sh
```
> Consider to generate queries with same parameters as generated data.

## Running queries

**Requires [VictoriaMetrics support][victoriametrics-pull-request].**

This section uses `tsbs_run_queries_victoriametrics` 
(see [tsbs_run_queries_victoriametrics](https://github.com/timescale/tsbs/blob/46f535f7d36c41ed3091ff050c63c7da84d0ddd8/docs/victoriametrics.md#tsbs_run_queries_victoriametrics))

It is necessary to have a Prometheus installed with the following minimum configuration:
```yaml
remote_read:
  - url: "http://localhost:9201/read" # Remote Storage read URL
    read_recent: true
```

To run generated queries follow examples in documentation:
```text
cat /tmp/bulk_queries/victoriametrics-cpu-max-all-8-queries.gz | gunzip | tsbs_run_queries_victoriametrics -urls http://localhost:9090
```
> Assumed that Prometheus is ready for querying on the `localhost` host and the `9090` port.
> Change `-urls http://localhost:9090` if this is not the case.

[victoriametrics-pull-request]: https://github.com/timescale/tsbs/pull/96
