package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	var (
		srv       = flag.String("srv", "127.0.0.1", "DNS server IP address to stress")
		port      = flag.String("port", "53", "DNS server Port to test")
		dataDir   = flag.String("dataDir", "testdata", "Data file directory")
		tfile     = flag.String("tfile", "", "tfile Specifies the input data file. If not specified, use alexa top 1 million as default")
		workerNum = flag.Int("workerNum", 1, "Number of simultaneous test workers to run")
		domainNum = flag.Int("domainNum", 1, "How many domain names to use in the test")
		timeout   = flag.Int("timeout", 5, "UDP timeout(seconds)")
		tp        = flag.Float64("tp", 0.95, "RTT top percentile")
		debug     = flag.Bool("debug", false, "Show debug info (default false)")
	)

	flag.Usage = func() {
		fmt.Println("Simple DNS stress tool\n" +
			"Options:")
		flag.PrintDefaults()

	}
	flag.Parse()

	var domains []string
	var err error

	err = os.MkdirAll(*dataDir, 0777)
	if err != nil {
		fmt.Printf("create dataDir(%s) err:%v", *dataDir, err)
		return
	}

	if *tfile == "" {
		domains, err = GetTop1mDomains(*dataDir)
	} else {
		domains, err = GetTestDomains(*tfile)
	}

	if err != nil {
		fmt.Printf("get domains err:%v, exit\n", err)
		return
	}
	fmt.Printf("got %d domains\n", len(domains))

	if *domainNum > len(domains) {
		*domainNum = len(domains)
	}

	var queries []DnsQuery
	for i := 0; i < *domainNum; i++ {
		queries = append(queries, DnsQuery{Domain: domains[i], Type: "A"})
	}

	s := NewStress(*srv, *port, (time.Duration)(*timeout)*time.Second, queries, *workerNum, *tp, *debug)
	fmt.Println("stress...")
	result := s.Start().Result()
	fmt.Printf("result: %s\n", result)
}
