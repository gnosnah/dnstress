package dnstress

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	progressInterval = 200 * time.Millisecond
	progressBarWidth = 24

	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiGreen  = "\033[32m"
	ansiRed    = "\033[31m"
	ansiCyan   = "\033[36m"
	ansiYellow = "\033[33m"
	ansiClear  = "\r\033[2K"
)

var (
	stdoutIsTTY = isTerminal(os.Stdout)
	printMu     sync.Mutex
)

func isTerminal(f *os.File) bool {
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return (st.Mode() & os.ModeCharDevice) != 0
}

func colorize(enabled bool, code, s string) string {
	if !enabled {
		return s
	}
	return code + s + ansiReset
}

func (s *Stress) printBanner() {
	addr := fmt.Sprintf("%s:%d", s.cfg.addr, s.cfg.port)
	parts := []string{
		colorize(stdoutIsTTY, ansiBold, "dnstress"),
		colorize(stdoutIsTTY, ansiCyan, addr),
		fmt.Sprintf("workers=%d", s.cfg.conQueryNum),
		fmt.Sprintf("type=%s", s.cfg.qType),
	}
	if s.cfg.maxQueries > 0 {
		parts = append(parts, fmt.Sprintf("queries=%d", s.cfg.maxQueries))
	}
	if s.cfg.maxTime > 0 {
		parts = append(parts, fmt.Sprintf("duration=%ds", s.cfg.maxTime))
	}
	printMu.Lock()
	fmt.Println(strings.Join(parts, "  "))
	printMu.Unlock()
}

func (s *Stress) renderProgress(final bool) {
	elapsed := time.Since(s.start).Seconds()
	if final && !s.end.IsZero() {
		elapsed = s.end.Sub(s.start).Seconds()
	}
	total := s.st.req.total.Load()
	success := s.st.req.success.Load()
	failed := s.st.req.failed.Load()
	rate := safeRate(success, total)
	qps := safeQPS(total, elapsed)
	pct, hasTarget := s.progressPercent(total, elapsed)

	var b strings.Builder
	if hasTarget {
		b.WriteString(renderBar(pct, progressBarWidth, stdoutIsTTY))
		b.WriteString("  ")
		b.WriteString(colorize(stdoutIsTTY, ansiYellow, fmt.Sprintf("%5.1f%%", pct)))
		b.WriteString("  ")
	}

	if s.cfg.maxQueries > 0 {
		b.WriteString(fmt.Sprintf("%d/%d", total, s.cfg.maxQueries))
	} else {
		b.WriteString(fmt.Sprintf("total %d", total))
	}

	okText := colorize(stdoutIsTTY, ansiGreen, fmt.Sprintf("ok %d", success))
	failText := colorize(stdoutIsTTY, ansiRed, fmt.Sprintf("fail %d", failed))
	rateText := fmt.Sprintf("%.1f%%", rate)
	if rate >= 99 {
		rateText = colorize(stdoutIsTTY, ansiGreen, rateText)
	} else if rate < 90 && total > 0 {
		rateText = colorize(stdoutIsTTY, ansiRed, rateText)
	}
	qpsText := colorize(stdoutIsTTY, ansiCyan, fmt.Sprintf("%d qps", qps))

	b.WriteString("  ")
	b.WriteString(okText)
	b.WriteString("  ")
	b.WriteString(failText)
	b.WriteString("  ")
	b.WriteString(rateText)
	b.WriteString("  ")
	b.WriteString(qpsText)
	b.WriteString("  ")
	b.WriteString(colorize(stdoutIsTTY, ansiDim, formatElapsed(elapsed)))

	line := b.String()
	printMu.Lock()
	defer printMu.Unlock()
	if stdoutIsTTY {
		fmt.Print(ansiClear + line)
		if final {
			fmt.Println()
		}
	} else {
		fmt.Println(line)
	}
}

func (s *Stress) progressPercent(total uint64, elapsed float64) (float64, bool) {
	switch {
	case s.cfg.maxQueries > 0:
		pct := float64(total) / float64(s.cfg.maxQueries) * 100
		if pct > 100 {
			pct = 100
		}
		return pct, true
	case s.cfg.maxTime > 0:
		pct := elapsed / float64(s.cfg.maxTime) * 100
		if pct > 100 {
			pct = 100
		}
		return pct, true
	default:
		return 0, false
	}
}

func renderBar(pct float64, width int, color bool) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
	return colorize(color, ansiCyan, bar)
}

func formatElapsed(sec float64) string {
	if sec < 0 {
		sec = 0
	}
	m := int(sec) / 60
	s := sec - float64(m*60)
	if m > 0 {
		return fmt.Sprintf("%d:%04.1f", m, s)
	}
	return fmt.Sprintf("%.1fs", s)
}

func clearProgressLine() {
	printMu.Lock()
	defer printMu.Unlock()
	if stdoutIsTTY {
		fmt.Print(ansiClear)
	}
}
