package dnstress

import (
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type tStats struct {
	req *reqStats
	rtt *rttTopPercentileStats
}

func newTStats(cfg *config) *tStats {
	var s tStats
	s.req = &reqStats{}
	if cfg.tpSize > 0 {
		s.rtt = &rttTopPercentileStats{
			max: cfg.tpSize,
			tp:  float64(cfg.tp) / 100,
		}
	}
	return &s
}

type reqStats struct {
	total   atomic.Uint64
	success atomic.Uint64
	failed  atomic.Uint64
}

type rttTopPercentileStats struct {
	mu   sync.Mutex
	rtts []time.Duration
	max  int
	tp   float64
}

func (s *rttTopPercentileStats) append(rtts []time.Duration) {
	if len(rtts) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rtts = append(s.rtts, rtts...)
}

func (s *rttTopPercentileStats) topPercentile() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := len(s.rtts)
	if n == 0 {
		return 0
	}
	if n > s.max {
		n = s.max
	}
	rtts := s.rtts[:n]
	sort.Slice(rtts, func(i, j int) bool { return rtts[i] < rtts[j] })
	idx := int(math.Ceil(s.tp * float64(n)))
	if idx < 1 {
		idx = 1
	}
	if idx > n {
		idx = n
	}
	return rtts[idx-1]
}
