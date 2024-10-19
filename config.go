package dnstress

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
)

var (
	defaultAddr            = "127.0.0.1"
	defaultPort            = 53
	defaultMaxQueries      = 0
	defaultMaxTime         = 0
	defaultConQueryNum     = 1
	defaultRttTpSize       = 0
	defaultRttTpVal        = 95
	defaultQueryTimeoutSec = 5
	defaultQType           = "A"
	defaultEnableDNSSEC    = false
	defaultShowProgress    = false
	defaultShowDebug       = false
)

type config struct {
	datafile     string
	addr         string
	port         int
	maxQueries   int
	maxTime      int
	conQueryNum  int
	tpSize       int
	tp           int
	timeout      int
	qType        string
	enableDNSSEC bool
	showProgress bool
	showDebug    bool
}

func (cfg *config) verify() error {
	if cfg.addr == "" || !validIP(cfg.addr) {
		return fmt.Errorf("invalid address: %s", cfg.addr)
	}
	if cfg.port <= 0 || cfg.port > 65535 {
		return fmt.Errorf("invalid port: %d", cfg.port)
	}
	if cfg.datafile == "" {
		return fmt.Errorf("invalid data file path %s", cfg.datafile)
	}
	if cfg.maxQueries < 0 {
		return fmt.Errorf("invalid max queries: %d", cfg.maxQueries)
	}
	if cfg.maxTime < 0 {
		return fmt.Errorf("invalid max time: %d", cfg.maxTime)
	}
	if cfg.conQueryNum < 1 {
		return fmt.Errorf("invalid concurrent queries num: %d", cfg.conQueryNum)
	}
	if cfg.tpSize < 0 {
		return fmt.Errorf("invalid RTT statistics array size: %d", cfg.tpSize)
	}
	if cfg.tp <= 0 || cfg.tp >= 100 {
		return fmt.Errorf("invalid RTT statistics top percentile value: %d", cfg.tp)
	}
	if cfg.timeout < 0 {
		return fmt.Errorf("invalid query timeout: %d", cfg.timeout)
	}
	if cfg.qType == "" || !validQType(cfg.qType) {
		return fmt.Errorf("invalid qtype: %s", cfg.qType)
	}
	return nil
}

func validQType(s string) bool {
	_, ok := dns.StringToType[s]
	return ok
}

func validIP(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil
}

func defaultConfig() *config {
	var cfg config
	cfg.addr = defaultAddr
	cfg.port = defaultPort
	cfg.maxQueries = defaultMaxQueries
	cfg.maxTime = defaultMaxTime
	cfg.conQueryNum = defaultConQueryNum
	cfg.tpSize = defaultRttTpSize
	cfg.tp = defaultRttTpVal
	cfg.timeout = defaultQueryTimeoutSec
	cfg.qType = defaultQType
	cfg.enableDNSSEC = defaultEnableDNSSEC
	cfg.showProgress = defaultShowProgress
	cfg.showDebug = defaultShowDebug
	return &cfg
}

type Option func(stress *Stress)

func WithMaxQueries(n int) Option {
	return func(s *Stress) {
		s.cfg.maxQueries = n
	}
}

func WithTimeout(n int) Option {
	return func(s *Stress) {
		s.cfg.timeout = n
	}
}

func WithMaxTime(n int) Option {
	return func(s *Stress) {
		s.cfg.maxTime = n
	}
}

func WithConQueryNum(n int) Option {
	return func(s *Stress) {
		s.cfg.conQueryNum = n
	}
}

func WithRttTpSize(n int) Option {
	return func(s *Stress) {
		s.cfg.tpSize = n
	}
}

func WithRttTpVal(n int) Option {
	return func(s *Stress) {
		s.cfg.tp = n
	}
}

func WithEnableDNSSEC(n bool) Option {
	return func(s *Stress) {
		s.cfg.enableDNSSEC = n
	}
}

func WithQType(qt string) Option {
	return func(s *Stress) {
		s.cfg.qType = qt
	}
}

func WithShowProcess(show bool) Option {
	return func(s *Stress) {
		s.cfg.showProgress = show
	}
}

func WithShowDebug(show bool) Option {
	return func(s *Stress) {
		s.cfg.showDebug = show
	}
}
