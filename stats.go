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
	rtts []time.Duration
	max  int
	tp   float64
	sync.Mutex
}

func (s *rttTopPercentileStats) append(rtts []time.Duration) {
	s.Lock()
	defer s.Unlock()
	s.rtts = append(s.rtts, rtts...)
}

func (s *rttTopPercentileStats) topPercentile() time.Duration {
	if len(s.rtts) == 0 {
		return 0
	}
	s.rtts = s.rtts[:s.max]
	sort.Slice(s.rtts, func(i, j int) bool { return s.rtts[i] < s.rtts[j] })
	f := math.Ceil(s.tp * float64(len(s.rtts)))
	idx := int(f)
	val := s.rtts[idx-1]
	return val
}
