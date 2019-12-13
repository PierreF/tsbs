package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/timescale/tsbs/internal/utils"
	"github.com/timescale/tsbs/load"

	"bufio"
	"fmt"
	"log"
	"strings"
)

const errMissingURLsFlag = "flag error: missing 'urls' flag"

var (
	remoteURLs []string
)

// Global vars
var (
	loader *load.BenchmarkRunner
)

// allows for testing
var fatal = log.Fatalf

func init() {
	var config load.BenchmarkRunnerConfig

	config.AddToFlagSet(pflag.CommandLine)

	pflag.String("urls", "http://localhost:1234", "Prometheus remote storage URLs")

	pflag.Parse()

	err := utils.SetupConfigFile()

	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("unable to decode config: %s", err))
	}

	remoteURLsStr := viper.GetString("urls")

	if len(remoteURLsStr) == 0 {
		fatal(errMissingURLsFlag)
	}

	remoteURLs = strings.Split(remoteURLsStr, ",")
	loader = load.GetBenchmarkRunner(config)
}

type benchmark struct{}

func (b *benchmark) GetPointDecoder(br *bufio.Reader) load.PointDecoder {
	return &decoder{scanner: bufio.NewScanner(br)}
}

func (b *benchmark) GetBatchFactory() load.BatchFactory {
	return &factory{}
}

func (b *benchmark) GetPointIndexer(_ uint) load.PointIndexer {
	return &load.ConstantIndexer{}
}

func (b *benchmark) GetProcessor() load.Processor {
	return &processor{}
}

func (b *benchmark) GetDBCreator() load.DBCreator {
	return &dbCreator{}
}

func main() {
	loader.RunBenchmark(&benchmark{}, load.SingleQueue)
}
