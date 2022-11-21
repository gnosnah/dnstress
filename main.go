package main

import (
	"flag"
	"fmt"
	"time"
)

// usage: ./dnsbench -srv 8.8.8.8 -port 53 -workerNum 10 -domainNum 200 -timeout 3
func main() {
	var (
		srv       = flag.String("srv", "127.0.0.1", "DNS server IP address to stress")
		port      = flag.String("port", "53", "DNS server Port to test")
		workerNum = flag.Int("workerNum", 1, "Number of simultaneous test workers to run")
		domainNum = flag.Int("domainNum", 1, "How many domain names to use in the test")
		timeout   = flag.Int("timeout", 5, "UDP timeout (seconds, default 5s)")
		debug     = flag.Bool("debug", false, "Show debug info (default false)")
	)

	flag.Usage = func() {
		fmt.Println("Simple DNS stress tool\n" +
			"Options:")
		flag.PrintDefaults()

	}
	flag.Parse()

	domains, err := GetTop1mDomains()
	if err != nil {
		fmt.Printf("get top 1m domains err:%v, exit", err)
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

	s := NewStress(*srv, *port, (time.Duration)(*timeout)*time.Second, queries, *workerNum, *debug)
	fmt.Println("stress...")
	result := s.Start().Result()
	fmt.Printf("result: %s\n", result)
}
