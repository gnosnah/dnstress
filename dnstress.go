package main

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type DnsQuery struct {
	Domain string
	Type   string
}

type Stress struct {
	srv, port string
	timeout   time.Duration
	queries   []DnsQuery
	queryIdx  int
	qMu       *sync.Mutex
	wg        *sync.WaitGroup
	workerNum int
	debug     bool

	// stat
	failedQueries      []DnsQuery
	fqMu               *sync.Mutex
	succeedQueries     uint64
	startTime, endTime time.Time
}

func NewStress(srv, port string, timeout time.Duration, queries []DnsQuery, workNum int, debug bool) *Stress {
	var s Stress
	s.srv = srv
	s.port = port
	s.timeout = timeout
	s.queries = queries
	s.queryIdx = 0
	s.qMu = &sync.Mutex{}
	s.wg = &sync.WaitGroup{}
	s.workerNum = workNum
	s.debug = debug
	s.fqMu = &sync.Mutex{}
	s.succeedQueries = 0

	return &s
}

func (s *Stress) Start() *Stress {
	s.startTime = time.Now()
	for i := 0; i < s.workerNum; i++ {
		s.wg.Add(1)
		go s.doDnsQuery(i)
	}
	s.wg.Wait()
	s.endTime = time.Now()
	return s
}

func (s *Stress) Result() string {
	return fmt.Sprintf("Queries total:%d, succeed:%d, failed:%d, success rate:%.2f%%, elapsed:%.2f(s)",
		len(s.queries), s.succeedQueries, len(s.failedQueries),
		float64(s.succeedQueries)/float64(len(s.queries))*100,
		s.endTime.Sub(s.startTime).Seconds())
}

func (s *Stress) doDnsQuery(workID int) {
	for {
		qDomain, qType := s.getOneQuery()
		if qDomain == "" {
			s.wg.Done()
			break
		}

		if qType == "" {
			qType = "A" // default A
		}

		client := new(dns.Client)
		client.DialTimeout = s.timeout
		client.ReadTimeout = s.timeout
		reqMsg := new(dns.Msg)
		qDomain = dns.Fqdn(qDomain)
		reqMsg.SetQuestion(qDomain, dns.StringToType[qType])
		_, _, err := client.Exchange(reqMsg, net.JoinHostPort(s.srv, s.port))
		if err == nil {
			atomic.AddUint64(&s.succeedQueries, 1)
		} else {
			s.fqMu.Lock()
			s.failedQueries = append(s.failedQueries, DnsQuery{Domain: qDomain, Type: qType})
			s.fqMu.Unlock()

			if s.debug {
				fmt.Printf("worker[%d]: query %s[%s] err:%v\n", workID, qDomain, qType, err)
			}
		}
	}
}

func (s *Stress) getOneQuery() (qDomain, qType string) {
	s.qMu.Lock()
	defer s.qMu.Unlock()
	if s.queryIdx < len(s.queries) {
		q := s.queries[s.queryIdx]
		s.queryIdx++
		return q.Domain, q.Type
	}
	return "", ""
}
