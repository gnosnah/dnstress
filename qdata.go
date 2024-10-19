package dnstress

import (
	"sync"
)

type queryItem struct {
	dName string
	dType string
}

type queryData struct {
	queries []*queryItem
	idx     int
	mu      sync.Mutex
}

func newQueryData(queries []*queryItem) *queryData {
	return &queryData{
		queries: queries,
		idx:     0,
	}
}

func (qd *queryData) get() *queryItem {
	qd.mu.Lock()
	defer qd.mu.Unlock()
	if qd.idx >= len(qd.queries) {
		qd.idx = 0
	}
	item := qd.queries[qd.idx]
	qd.idx++
	return item
}
