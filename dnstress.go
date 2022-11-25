package main

import (
	"fmt"
	"github.com/miekg/dns"
	"math"
	"net"
	"sort"
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
	tp        float64
	debug     bool

	// stats
	failedQueries      []DnsQuery
	fqMu               *sync.Mutex
	succeedQueries     uint64
	startTime, endTime time.Time
	elapsed            []int64
}

func NewStress(srv, port string, timeout time.Duration, queries []DnsQuery, workNum int, tp float64, debug bool) *Stress {
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
	s.tp = tp
	s.fqMu = &sync.Mutex{}
	s.succeedQueries = 0
	s.elapsed = make([]int64, len(s.queries))

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
	elapsed := s.endTime.Sub(s.startTime).Seconds()
	return fmt.Sprintf("Queries total:%d, succeed:%d, failed:%d, success rate:%.2f%%, elapsed:%.2f(s), TP%d:%d(Milliseconds)",
		len(s.queries), s.succeedQueries, len(s.failedQueries),
		float64(s.succeedQueries)/float64(len(s.queries))*100,
		elapsed,
		int(s.tp*100),
		time.Duration(s.topPercentile(s.tp)).Milliseconds(),
	)
}

func (s *Stress) doDnsQuery(workID int) {
	for {
		idx := s.getOneQuery()
		if idx == errIndex {
			s.wg.Done()
			break
		}
		qDomain, qType := s.queries[idx].Domain, s.queries[idx].Type

		if qType == "" {
			qType = "A" // default A
		}

		client := new(dns.Client)
		client.DialTimeout = s.timeout
		client.ReadTimeout = s.timeout
		client.UDPSize = 4096
		reqMsg := new(dns.Msg)
		qDomain = dns.Fqdn(qDomain)
		reqMsg.SetQuestion(qDomain, dns.StringToType[qType])
		_, rtt, err := client.Exchange(reqMsg, net.JoinHostPort(s.srv, s.port))
		if err == nil {
			atomic.AddUint64(&s.succeedQueries, 1)
		} else {
			s.fqMu.Lock()
			s.failedQueries = append(s.failedQueries, DnsQuery{Domain: qDomain, Type: qType})
			s.fqMu.Unlock()

			if s.debug {
				fmt.Printf("worker[%d]: query %s[%s] rtt:%d, err:%v\n", workID, qDomain, qType, rtt, err)
			}
		}
		s.elapsed[idx] = rtt.Nanoseconds()
	}
}

const errIndex = -1

func (s *Stress) getOneQuery() int {
	s.qMu.Lock()
	defer s.qMu.Unlock()
	if s.queryIdx < len(s.queries) {
		idx := s.queryIdx
		s.queryIdx++
		return idx
	}
	return errIndex
}

func (s *Stress) topPercentile(tp float64) int64 {
	sort.Slice(s.elapsed, func(i, j int) bool { return s.elapsed[i] < s.elapsed[j] })
	f := math.Ceil(tp * float64(len(s.elapsed)))
	idx := int(f)
	return s.elapsed[idx-1]
}
