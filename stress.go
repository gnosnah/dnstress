package dnstress

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/miekg/dns"
)

const (
	commentPrefix = "#"
	dialRetryWait = 10 * time.Millisecond
)

type queryItem struct {
	dName string
	dType string
}

type Stress struct {
	cfg        *config
	queries    []*queryItem
	st         *tStats
	wg         *sync.WaitGroup
	start, end time.Time
	running    atomic.Bool
	stopped    atomic.Bool
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
	if len(qs) == 0 {
		return fmt.Errorf("no valid query items in %s", s.cfg.datafile)
	}
	s.queries = qs
	return nil
}

func (s *Stress) Start() *Stress {
	s.start = time.Now()
	s.running.Store(true)

	var progressDone chan struct{}
	if s.cfg.showProgress {
		progressDone = make(chan struct{})
		go func() {
			defer close(progressDone)
			s.showProgress()
		}()
	}

	s.wg.Add(s.cfg.conQueryNum)
	for i := 0; i < s.cfg.conQueryNum; i++ {
		go s.doQuery()
	}
	s.wg.Wait()
	s.end = time.Now()
	s.running.Store(false)
	if progressDone != nil {
		<-progressDone
	}
	return s
}

// Stop requests all workers to exit (e.g. on SIGINT).
func (s *Stress) Stop() {
	s.stopped.Store(true)
}

func (s *Stress) doQuery() {
	defer s.wg.Done()

	addr := net.JoinHostPort(s.cfg.addr, strconv.Itoa(s.cfg.port))
	var msg dns.Msg
	client := dns.Client{
		DialTimeout: time.Duration(s.cfg.timeout) * time.Second,
		ReadTimeout: time.Duration(s.cfg.timeout) * time.Second,
	}

	var rtts []time.Duration
	if s.cfg.tpSize > 0 {
		tpSize := int(math.Ceil(float64(s.cfg.tpSize) / float64(s.cfg.conQueryNum)))
		rtts = make([]time.Duration, 0, tpSize)
	}
	maxRTTSamples := cap(rtts)

	var conn *dns.Conn
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()

	qIdx := 0
	for {
		if s.shouldStop() {
			break
		}

		if conn == nil {
			c, err := client.Dial(addr)
			if err != nil {
				s.st.req.total.Add(1)
				s.st.req.failed.Add(1)
				s.debugInfo("dial %s failed: %v", addr, err)
				time.Sleep(dialRetryWait)
				continue
			}
			conn = c
		}

		if qIdx >= len(s.queries) {
			qIdx = 0
		}
		q := s.queries[qIdx]
		qIdx++

		dname := dns.Fqdn(q.dName)
		dType := q.dType
		if dType == "" {
			dType = s.cfg.qType
		}
		dtype := dns.StringToType[dType]
		msg.SetQuestion(dname, dtype)
		if s.cfg.enableDNSSEC {
			msg.SetEdns0(dns.DefaultMsgSize, true)
		}

		resp, rtt, err := client.ExchangeWithConn(&msg, conn)
		s.st.req.total.Add(1)
		ok := err == nil && resp != nil && resp.Rcode == dns.RcodeSuccess
		if ok {
			s.st.req.success.Add(1)
			if len(rtts) < maxRTTSamples {
				rtts = append(rtts, rtt)
			}
			s.debugInfo("query: %s(%s) RTT: %dms", q.dName, dType, rtt.Milliseconds())
		} else {
			s.st.req.failed.Add(1)
			if err != nil {
				_ = conn.Close()
				conn = nil
				s.debugInfo("query: %s(%s) RTT: %dms, err: %v", q.dName, dType, rtt.Milliseconds(), err)
			} else {
				s.debugInfo("query: %s(%s) RTT: %dms, rcode: %s", q.dName, dType, rtt.Milliseconds(), dns.RcodeToString[resp.Rcode])
			}
		}
		resetMsg(&msg)
	}

	if s.st.rtt != nil && len(rtts) > 0 {
		s.st.rtt.append(rtts)
	}
}

func (s *Stress) shouldStop() bool {
	if s.stopped.Load() {
		return true
	}
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
	s.printBanner()
	interval := progressInterval
	if !stdoutIsTTY {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		running := s.running.Load()
		s.renderProgress(!running)
		if !running {
			return
		}
		<-ticker.C
	}
}

func (s *Stress) Stats() {
	if !s.cfg.showProgress {
		clearProgressLine()
	}

	elapsed := s.end.Sub(s.start).Seconds()
	total := s.st.req.total.Load()
	success := s.st.req.success.Load()
	failed := s.st.req.failed.Load()
	successRate := safeRate(success, total)
	qps := safeQPS(total, elapsed)

	var rows []table.Row
	rows = append(rows, table.Row{"Total", total})
	rows = append(rows, table.Row{"Succeeded", success})
	rows = append(rows, table.Row{"Failed", failed})
	rows = append(rows, table.Row{"Success rate", fmt.Sprintf("%.2f%%", successRate)})
	rows = append(rows, table.Row{"QPS", qps})
	if s.st.rtt != nil {
		tp := s.st.rtt.topPercentile().Milliseconds()
		rows = append(rows, table.Row{fmt.Sprintf("RTT TP%d", s.cfg.tp), fmt.Sprintf("%dms", tp)})
	}
	rows = append(rows, table.Row{"Started", s.start.Format(time.DateTime)})
	rows = append(rows, table.Row{"Finished", s.end.Format(time.DateTime)})
	rows = append(rows, table.Row{"Elapsed", formatElapsed(elapsed)})

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"Metric", "Value"})
	for _, row := range rows {
		t.AppendRow(row)
	}
	fmt.Println()
	fmt.Println(colorize(stdoutIsTTY, ansiBold, "Statistics"))
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
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, commentPrefix) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		var q queryItem
		q.dName = fields[0]
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
	if !s.cfg.showDebug {
		return
	}
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", a...)
}

func safeRate(success, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(success) / float64(total) * 100
}

func safeQPS(total uint64, elapsed float64) int {
	if total == 0 {
		return 0
	}
	if elapsed < 0.001 {
		elapsed = 0.001
	}
	return int(float64(total) / elapsed)
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

