package main

import (
	"container/list"
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
	queries   *list.List
	qMu       *sync.Mutex
	wg        *sync.WaitGroup
	workerNum int
	debug     bool

	// stat
	failedQueries      *list.List
	fqMu               *sync.Mutex
	succeedQueries     uint64
	totalQueries       int
	startTime, endTime time.Time
}

func NewStress(srv, port string, timeout time.Duration, queries []DnsQuery, workNum int, debug bool) *Stress {
	var s Stress
	s.srv = srv
	s.port = port
	s.timeout = timeout
	s.queries = list.New()
	s.qMu = &sync.Mutex{}
	s.wg = &sync.WaitGroup{}
	s.workerNum = workNum
	s.debug = debug

	s.failedQueries = list.New()
	s.fqMu = &sync.Mutex{}
	s.succeedQueries = 0

	for i := 0; i < len(queries); i++ {
		s.queries.PushBack(queries[i])
	}
	s.totalQueries = s.queries.Len()

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
	return fmt.Sprintf("Queries total:%d, succeed:%d, failed:%d,  success rate:%.4f, elapsed:%.3f(s)",
		s.totalQueries, s.succeedQueries, s.failedQueries.Len(),
		float64(s.succeedQueries)/float64(s.totalQueries),
		s.endTime.Sub(s.startTime).Seconds())
}

func (s *Stress) doDnsQuery(workID int) {
	for {
		qDomain, qType := s.getOne()
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
			s.failedQueries.PushBack(DnsQuery{Domain: qDomain, Type: qType})
			s.fqMu.Unlock()

			if s.debug {
				fmt.Printf("worker[%d]: exchange %s[%s] err:%v\n", workID, qDomain, qType, err)
			}
		}
	}
}

func (s *Stress) getOne() (qDomain, qType string) {
	s.qMu.Lock()
	defer s.qMu.Unlock()
	element := s.queries.Front()
	if element == nil {
		return "", ""
	}
	s.queries.Remove(element)
	q := element.Value.(DnsQuery)
	return q.Domain, q.Type
}
