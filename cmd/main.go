package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gnosnah/dnstress"
)

const (
	version = "1.0.2"
)

var (
	datafile     = flag.String("d", "example.txt", "specify the input data file")
	addr         = flag.String("s", "127.0.0.1", "sets the server to query (IP or hostname)")
	port         = flag.Int("p", 53, "set the port on which to query the server")
	maxQueries   = flag.Int("q", 0, "specify the maximum number of queries to send (0 = unlimited)")
	timeout      = flag.Int("t", 5, "specify the timeout for query completion in seconds")
	maxTime      = flag.Int("l", 0, "specify how long to run tests in seconds (0 = unlimited)")
	conQueryNum  = flag.Int("c", 1, "specify the number of concurrent queries")
	rttTpSize    = flag.Int("r", 50000, "set RTT statistics array size")
	rttTpVal     = flag.Int("u", 95, "set RTT statistics top percentile value(1-99)")
	enableDNSSEC = flag.Bool("D", false, "set the DNSSEC OK bit (implies EDNS)")
	qType        = flag.String("T", "A", "specify the default query type")
	showProgress = flag.Bool("g", false, "show real-time progress")
	showDebug    = flag.Bool("v", false, "show debug info")
	showHelp     = flag.Bool("h", false, "show help")
)

func main() {
	flag.Usage = func() {
		fmt.Println("dnstress version " + version)
		fmt.Println("options:")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showHelp {
		flag.Usage()
		return
	}

	s := dnstress.NewStress(*addr, *port, *datafile,
		dnstress.WithMaxQueries(*maxQueries),
		dnstress.WithTimeout(*timeout),
		dnstress.WithMaxTime(*maxTime),
		dnstress.WithConQueryNum(*conQueryNum),
		dnstress.WithRttTpSize(*rttTpSize),
		dnstress.WithRttTpVal(*rttTpVal),
		dnstress.WithEnableDNSSEC(*enableDNSSEC),
		dnstress.WithQType(*qType),
		dnstress.WithShowProgress(*showProgress),
		dnstress.WithShowDebug(*showDebug),
	)
	if err := s.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		s.Stop()
	}()

	s.Start()
	s.Stats()
}
