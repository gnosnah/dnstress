package dnstress

import (
	"fmt"
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

func (s *reqStats) result(start, end time.Time) string {
	elapsed := end.Sub(start).Seconds()
	total := s.total.Load()
	success := s.success.Load()
	failed := s.failed.Load()
	successRate := float64(success) / float64(total) * 100
	return fmt.Sprintf("[stats-request] total:%d, succeed:%d, failed:%d, success rate:%.2f%%, elapsed:%.2f(s)\n",
		total, success, failed, successRate, elapsed,
	)
}

type rttTopPercentileStats struct {
	rtts []time.Duration
	max  int
	tp   float64
	sync.Mutex
}

func (s *rttTopPercentileStats) append(rtt time.Duration) {
	s.Lock()
	defer s.Unlock()
	if len(s.rtts) >= s.max {
		return
	}
	s.rtts = append(s.rtts, rtt)
}

func (s *rttTopPercentileStats) topPercentile() time.Duration {
	if len(s.rtts) == 0 {
		return 0
	}
	sort.Slice(s.rtts, func(i, j int) bool { return s.rtts[i] < s.rtts[j] })
	f := math.Ceil(s.tp * float64(len(s.rtts)))
	idx := int(f)
	val := s.rtts[idx-1]
	return val
}
