package dnstress

import (
	"bufio"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/miekg/dns"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	progressInterval = 1 * time.Second
	commentPrefix    = "#"
)

type Stress struct {
	cfg        *config
	qdata      *queryData
	st         *tStats
	wg         *sync.WaitGroup
	start, end time.Time
	running    atomic.Bool
	counter    atomic.Uint64
}

func NewStress(addr string, port int, datafile string, options ...Option) *Stress {
	var s Stress
	s.cfg = defaultConfig()
	s.cfg.addr = addr
	s.cfg.port = port
	s.cfg.datafile = datafile
	for _, opt := range options {
		opt(&s)
	}
	s.st = newTStats(s.cfg)
	s.wg = &sync.WaitGroup{}
	s.running.Store(false)
	return &s
}

func (s *Stress) Init() error {
	err := s.cfg.verify()
	if err != nil {
		return err
	}
	qs, err := s.loadData(s.cfg.datafile)
	if err != nil {
		return err
	}
	s.qdata = newQueryData(qs)
	return nil
}

func (s *Stress) Start() *Stress {
	s.start = time.Now()
	s.running.Store(true)
	if s.cfg.showProgress {
		go s.showProgress()
	}
	s.wg.Add(s.cfg.conQueryNum)
	for i := 0; i < s.cfg.conQueryNum; i++ {
		go s.doQuery()
	}
	s.wg.Wait()
	s.running.Store(false)
	s.end = time.Now()
	return s
}

func (s *Stress) doQuery() {
	addr := net.JoinHostPort(s.cfg.addr, strconv.Itoa(s.cfg.port))
	var msg dns.Msg
	var client dns.Client
	timeout := time.Duration(s.cfg.timeout) * time.Second
	client.DialTimeout = timeout
	client.ReadTimeout = timeout

	for {
		if s.shouldStop() {
			break
		}

		q := s.qdata.get()
		dname := dns.Fqdn(q.dName)
		if q.dType == "" {
			q.dType = s.cfg.qType
		}
		dtype := dns.StringToType[q.dType]
		msg.SetQuestion(dname, dtype)
		if s.cfg.enableDNSSEC {
			msg.SetEdns0(dns.DefaultMsgSize, true)
		}
		_, rtt, err := client.Exchange(&msg, addr)
		s.st.req.total.Add(1)
		if err == nil {
			s.st.req.success.Add(1)
			s.st.rtt.append(rtt)
			s.debugInfo("query: %s(%s) RTT: %dms", q.dName, q.dType, rtt.Milliseconds())
		} else {
			s.st.req.failed.Add(1)
			s.debugInfo("query: %s(%s) RTT: %dms, err: %v", q.dName, q.dType, rtt.Milliseconds(), err)
		}
		resetMsg(&msg)
	}

	s.wg.Done()
}

func (s *Stress) shouldStop() bool {
	if s.cfg.maxTime > 0 {
		if time.Since(s.start) >= time.Duration(s.cfg.maxTime)*time.Second {
			return true
		}
	}

	if s.cfg.maxQueries > 0 {
		if s.counter.Add(1) > uint64(s.cfg.maxQueries) {
			return true
		}
	}
	return false
}

func (s *Stress) showProgress() {
	for {
		if !s.running.Load() {
			return
		}

		elapsed := time.Since(s.start).Seconds()
		total := s.st.req.total.Load()
		success := s.st.req.success.Load()
		failed := s.st.req.failed.Load()
		successRate := float64(success) / float64(total) * 100
		qps := int(float64(total) / elapsed)
		clearCurLine()
		format := "\r[progress] total:%d, succeed:%d, failed:%d, success rate:%.2f%%, elapsed:%ds, qps:%d"
		fmt.Printf(format, total, success, failed, successRate, int(elapsed), qps)

		time.Sleep(progressInterval)
	}
}

func (s *Stress) Stats() {
	clearCurLine()

	elapsed := s.end.Sub(s.start).Seconds()
	total := s.st.req.total.Load()
	success := s.st.req.success.Load()
	failed := s.st.req.failed.Load()
	successRate := float64(success) / float64(total) * 100
	qps := int(float64(total) / elapsed)
	tp := s.st.rtt.topPercentile().Milliseconds()

	var rows []table.Row
	rows = append(rows, table.Row{"Queries total", total})
	rows = append(rows, table.Row{"Queries succeeded", success})
	rows = append(rows, table.Row{"Queries failed", failed})
	rows = append(rows, table.Row{"Queries success rate", fmt.Sprintf("%.2f%%", successRate)})
	rows = append(rows, table.Row{"Queries QPS", qps}) // average
	rows = append(rows, table.Row{fmt.Sprintf("Queries RTT TP%d", s.cfg.tp), fmt.Sprintf("%dms", tp)})
	rows = append(rows, table.Row{"Queries started at", s.start.Format(time.DateTime)})
	rows = append(rows, table.Row{"Queries finished at", s.end.Format(time.DateTime)})
	rows = append(rows, table.Row{"Queries elapsed", fmt.Sprintf("%ds", int(elapsed))})

	t := table.NewWriter()
	for _, row := range rows {
		t.AppendRow(row)
	}
	fmt.Println("Statistics:")
	fmt.Println(t.Render())
}

func (s *Stress) loadData(path string) (qs []*queryItem, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, commentPrefix) {
			continue
		}
		var q queryItem
		fields := strings.Fields(line)
		q.dName = strings.TrimSpace(fields[0])
		if len(fields) > 1 {
			dType := strings.ToUpper(fields[1])
			if !validQType(dType) {
				continue
			}
			q.dType = dType
		}
		qs = append(qs, &q)
	}
	s.debugInfo("load query items: %d", len(qs))
	return qs, scanner.Err()
}

func (s *Stress) debugInfo(format string, a ...any) {
	if s.cfg.showDebug {
		fmt.Printf("\n[debug] "+format+"", a...)
	}
}

func resetMsg(msg *dns.Msg) {
	msg.MsgHdr.Id = 0
	msg.MsgHdr.Response = false
	msg.MsgHdr.Opcode = 0
	msg.MsgHdr.Authoritative = false
	msg.MsgHdr.Truncated = false
	msg.MsgHdr.RecursionDesired = false
	msg.MsgHdr.RecursionAvailable = false
	msg.MsgHdr.Zero = false
	msg.MsgHdr.AuthenticatedData = false
	msg.MsgHdr.CheckingDisabled = false
	msg.MsgHdr.Rcode = 0
	msg.Compress = false
	msg.Question = msg.Question[:0]
	msg.Answer = msg.Answer[:0]
	msg.Ns = msg.Ns[:0]
	msg.Extra = msg.Extra[:0]
}

func clearCurLine() {
	fmt.Print("\n\033[1A\033[K")
}
