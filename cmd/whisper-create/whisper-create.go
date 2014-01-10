package main

import (
	"../../whisper"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	aggregationMethod = flag.String("a", "average", "aggregation method to use")
	xFilesFactor      = flag.Float64("x", 0.5, "x-files factor")
)

const retention_def_help = `
Retention definitions are specified as pairs timePerPoint:timeToStore,
where timePerPoint and timeToStore specify lengths of time, for
example:

60:1440      60 seconds per datapoint, 1440 datapoints = 1 day of retention
15m:8        15 minutes per datapoint, 8 datapoints = 2 hours of retention
1h:7d        1 hour per datapoint, 7 days of retention
12h:2y       12 hours per datapoint, 2 years of retention

`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: %s [options] filename retention_def...\n",
			os.Args[0], os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, retention_def_help)
	}
	flag.Parse()
	method := whisper.AggregationMethodByName(*aggregationMethod)
	log.SetFlags(0)

	if flag.NArg() < 2 {
		flag.Usage()
		log.Fatal("required arguments missing")
	}

	if method == whisper.AggregationUnknown {
		log.Fatal("error: unknown aggregation method: ", *aggregationMethod)
	}

	args := flag.Args()
	path := args[0]
	archiveStrings := args[1:]

	var archives []whisper.ArchiveInfo
	for _, s := range archiveStrings {
		archive, err := whisper.ParseRetentionDef(s)
		if err != nil {
			log.Fatal(fmt.Sprintf("error: %s", err))
		}
		archives = append(archives, archive)
	}

	_, err := whisper.Create(path, archives, whisper.CreateOptions{XFilesFactor: float32(*xFilesFactor), AggregationMethod: method})
	if err != nil {
		log.Fatal(err)
	}
}
