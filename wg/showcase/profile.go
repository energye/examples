package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// frameProfile accumulates per-phase timings across frames. All values are
// microseconds; Report() prints per-frame averages plus frame-interval
// stability (p50/p95/p99/max, jank count and jitter).
type frameProfile struct {
	clearUs     int64
	setupUs     int64
	blitUs      int64
	rasterUs    int64 // AA software rasterisation (subset of widgetUs)
	loadUs      int64 // Go pixel buffer -> bitmap HDC
	widgetUs    []int64
	frames      int
	fullFrames  int
	dirtyFrames int
	first       time.Time
	last        time.Time
	intervals   []int64 // microseconds between consecutive presented frames
	prepareUs   []int64 // per-frame PREPARE+PRESENT cost (separates render jitter from scheduler jitter)
}

const maxIntervalSamples = 600

func (p *frameProfile) ensure(n int) {
	if len(p.widgetUs) < n {
		p.widgetUs = append(p.widgetUs, make([]int64, n-len(p.widgetUs))...)
	}
	if cap(p.intervals) < maxIntervalSamples {
		p.intervals = make([]int64, 0, maxIntervalSamples)
	}
	if cap(p.prepareUs) < maxIntervalSamples {
		p.prepareUs = make([]int64, 0, maxIntervalSamples)
	}
}

func (p *frameProfile) reset() {
	p.clearUs, p.setupUs, p.blitUs, p.rasterUs, p.loadUs, p.frames = 0, 0, 0, 0, 0, 0
	p.fullFrames, p.dirtyFrames = 0, 0
	p.first, p.last = time.Time{}, time.Time{}
	p.intervals = p.intervals[:0]
	p.prepareUs = p.prepareUs[:0]
	for i := range p.widgetUs {
		p.widgetUs[i] = 0
	}
}

// sample records one presented frame on the main thread.
func (p *frameProfile) sample(now time.Time, paintUs int64) {
	if !p.last.IsZero() {
		d := now.Sub(p.last).Microseconds()
		// Ignore >100ms gaps (startup, pause, window dragging): they are not
		// frame jitter and would destroy the percentile statistics.
		if d > 0 && d < 100000 {
			if len(p.intervals) >= maxIntervalSamples {
				copy(p.intervals, p.intervals[1:])
				p.intervals[len(p.intervals)-1] = d
			} else {
				p.intervals = append(p.intervals, d)
			}
		}
	}
	if p.first.IsZero() {
		p.first = now
	}
	p.last = now
	if len(p.prepareUs) >= maxIntervalSamples {
		copy(p.prepareUs, p.prepareUs[1:])
		p.prepareUs[len(p.prepareUs)-1] = paintUs
	} else {
		p.prepareUs = append(p.prepareUs, paintUs)
	}
}

func (p *frameProfile) prepareStats() (p50, p95, maxV float64) {
	n := len(p.prepareUs)
	if n == 0 {
		return
	}
	tmp := append([]int64(nil), p.prepareUs...)
	sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j] })
	return float64(tmp[n/2]) / 1000,
		float64(tmp[minIdx(n*95/100, n-1)]) / 1000,
		float64(tmp[n-1]) / 1000
}

func (p *frameProfile) begin() { p.reset() }

// stability returns frame-interval percentiles, jitter (std-dev) and the
// number of jank frames (>20ms). It allocates only in --bench reporting.
func (p *frameProfile) stability() (p50, p95, p99, maxV, sd float64, jank int) {
	n := len(p.intervals)
	if n == 0 {
		return
	}
	tmp := append([]int64(nil), p.intervals...)
	sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j] })
	ms := func(v int64) float64 { return float64(v) / 1000 }
	p50 = ms(tmp[n/2])
	p95 = ms(tmp[minIdx(n*95/100, n-1)])
	p99 = ms(tmp[minIdx(n*99/100, n-1)])
	maxV = ms(tmp[n-1])
	var sum, sum2 float64
	for _, v := range p.intervals {
		x := float64(v) / 1000
		sum += x
		sum2 += x * x
		if v > 20000 {
			jank++
		}
	}
	mean := sum / float64(n)
	sd = math.Sqrt(math.Max(0, sum2/float64(n)-mean*mean))
	return
}

func minIdx(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Report returns a ranked, per-frame average breakdown.
func (s *Scene) ProfileReport() string {
	p := &s.prof
	if p.frames == 0 {
		return "[PROFILE] no frames sampled"
	}
	n := float64(p.frames)
	dur := p.last.Sub(p.first).Seconds()
	steady := 0.0
	if dur > 0.5 && p.frames > 1 {
		steady = float64(p.frames-1) / dur
	}
	total := p.clearUs + p.setupUs + p.widgetTotal() + p.loadUs + p.blitUs
	var b []byte
	b = append(b, fmt.Sprintf("[PROFILE] %d 帧 (整帧 %d / 脏区 %d)  背景恢复=%.2fms  setup=%.2fms  widgets(含AA光栅化 %.2fms)=%.2fms  像素写回=%.2fms  PRESENT=%.2fms  PREPARE+PRESENT=%.2fms  帧率=%.1f FPS  (静态层烘焙一次 %.2fms)\n",
		p.frames, p.fullFrames, p.dirtyFrames,
		float64(p.clearUs)/n/1000, float64(p.setupUs)/n/1000,
		float64(p.rasterUs)/n/1000, float64(p.widgetTotal())/n/1000,
		float64(p.loadUs)/n/1000, float64(p.blitUs)/n/1000,
		float64(total)/n/1000, steady,
		float64(s.staticBakeUs+s.aaBakeUs)/1000)...)
	p50, p95, p99, mx, sd, jank := p.stability()
	if len(p.intervals) > 0 {
		q50, q95, qmax := p.prepareStats()
		b = append(b, fmt.Sprintf("[STABILITY] 帧间隔 %d 样本  p50=%.2fms  p95=%.2fms  p99=%.2fms  max=%.2fms  σ=%.2fms  jank(>20ms)=%d (%.1f%%)\n",
			len(p.intervals), p50, p95, p99, mx, sd, jank,
			float64(jank)*100/float64(len(p.intervals)))...)
		b = append(b, fmt.Sprintf("[STABILITY] PREPARE 耗时 p50=%.2fms  p95=%.2fms  max=%.2fms  (帧预算 16.67ms)\n",
			q50, q95, qmax)...)
	}
	frame := s.frameWidgets()
	type row struct {
		name string
		us   float64
	}
	rows := make([]row, 0, len(p.widgetUs))
	for i, v := range p.widgetUs {
		name := fmt.Sprintf("#%d", i)
		if i < len(frame) {
			q := frame[i].Rect()
			name = fmt.Sprintf("%s (%d,%d %dx%d)", frame[i].Name(),
				q.Left, q.Top, q.Width(), q.Height())
		}
		rows = append(rows, row{name, float64(v) / n / 1000})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].us > rows[j].us })
	for _, r := range rows {
		b = append(b, fmt.Sprintf("  %7.3fms  %s\n", r.us, r.name)...)
	}
	return string(b)
}

func (p *frameProfile) widgetTotal() int64 {
	var t int64
	for _, v := range p.widgetUs {
		t += v
	}
	return t
}

// measure is a tiny helper so the hot path stays readable.
func measure(start time.Time) int64 { return time.Since(start).Microseconds() }
