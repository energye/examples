package main

import (
	"sync"
	"sync/atomic"
	"time"
)

// FrameScheduler is the "passive frame pump".
//
// Mature UI frameworks (Flutter SchedulerBinding / Chromium BeginFrame /
// WPF CompositionTarget.Rendering / Android Choreographer) do NOT let every
// input event or a free-running timer paint whenever it likes. They:
//
//   1. mark something dirty (invalidate / requestFrame)
//   2. wait for the display's frame boundary (VSync, ~16.667ms)
//   3. on that boundary: PREPARE the whole frame (layout + paint into an
//      offscreen surface), then PRESENT it atomically
//   4. if an animation is still active, request the NEXT frame; otherwise
//      stop completely and burn zero CPU
//
// This scheduler reproduces that with a wall-clock-aligned 60Hz cadence and
// a private WM_APP message posted to the main thread. GDI has no VSync API in
// this backend, so the cadence is an absolute-time 16.667ms timeline; a
// production GPU backend would replace the timeline with DwmFlush /
// IDXGIOutput::WaitForVBlank.
//
// Two frame sources, exactly as requested:
//   - "有动态的场景": active animation -> a frame at every 60Hz boundary
//   - "要渲染刷新场景": a dirty region while idle -> exactly ONE frame
//
// Anything else (mouse move, key, hover) only mutates state / marks a dirty
// rect. It never schedules an extra frame, so input can never raise the FPS.

const (
	wmRequestFrame = 0x8000 + 1 // WM_APP + 1, private frame message
	frameInterval  = time.Second / 60
)

type FrameScheduler struct {
	hwnd uintptr

	mu      sync.Mutex
	active  bool // animation running -> continuous 60Hz frames
	oneShot bool // idle dirty region -> a single frame
	started bool

	wake chan struct{}
	stop chan struct{}

	inflight   int32 // 1 = a frame message is posted but not yet consumed
	inflightAt int64 // UnixNano of the post, used to recover a lost message
	posted     int64
	skipped    int64
	lost       int64
}

func NewFrameScheduler(hwnd uintptr) *FrameScheduler {
	return &FrameScheduler{
		hwnd: hwnd,
		wake: make(chan struct{}, 1),
		stop: make(chan struct{}),
	}
}

func (f *FrameScheduler) notify() {
	select {
	case f.wake <- struct{}{}:
	default:
	}
}

func (f *FrameScheduler) Start() {
	f.mu.Lock()
	if f.started {
		f.mu.Unlock()
		return
	}
	f.started = true
	f.mu.Unlock()
	go f.loop()
}

func (f *FrameScheduler) Stop() {
	f.mu.Lock()
	if !f.started {
		f.mu.Unlock()
		return
	}
	f.mu.Unlock()
	select {
	case <-f.stop:
	default:
		close(f.stop)
	}
}

func (f *FrameScheduler) isActive() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.active
}

// Active reports whether continuous 60Hz animation frames are being pumped.
func (f *FrameScheduler) Active() bool { return f.isActive() }

// SetActive switches continuous animation frames on/off.
func (f *FrameScheduler) SetActive(v bool) {
	f.mu.Lock()
	if f.active == v {
		f.mu.Unlock()
		return
	}
	f.active = v
	if v {
		// Continuous animation takes over the frame clock; any pending idle
		// one-shot is obsolete. Leaving it set would make FrameDone() signal
		// the wake channel after EVERY frame and reset the absolute timeline.
		f.oneShot = false
	}
	f.mu.Unlock()
	f.notify()
}

// RequestFrame asks for ONE frame. While animating this is deliberately a
// no-op: frames already arrive at 60Hz and extra requests would only cause
// the "FPS rises when the mouse moves" artefact.
func (f *FrameScheduler) RequestFrame() {
	f.mu.Lock()
	if f.active {
		f.mu.Unlock()
		return
	}
	f.oneShot = true
	f.mu.Unlock()
	f.notify()
}

func (f *FrameScheduler) takeOneShot() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	v := f.oneShot
	f.oneShot = false
	return v
}

func (f *FrameScheduler) hasOneShot() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.oneShot
}

// FrameDone is called by the main thread after it has consumed a frame.
// Releasing inflight here (and only here) prevents a slow frame from piling
// up catch-up messages in the queue.
func (f *FrameScheduler) FrameDone() {
	// Never let the counter go negative (e.g. a one-shot frame that arrived
	// outside the pump's accounting).
	for {
		n := atomic.LoadInt32(&f.inflight)
		if n <= 0 {
			break
		}
		if atomic.CompareAndSwapInt32(&f.inflight, n, n-1) {
			break
		}
	}
	atomic.StoreInt64(&f.inflightAt, 0)
	// Only an IDLE one-shot needs re-notifying. While animating, the loop is
	// already timing the next 60Hz slot - waking it would corrupt the cadence.
	if !f.isActive() && f.hasOneShot() {
		f.notify()
	}
}

func (f *FrameScheduler) post() {
	// Allow at most TWO outstanding frame messages. A single slow frame then
	// costs no slot (the queued one is consumed right after), while a persistently
	// stuck consumer still cannot pile up an unbounded catch-up burst.
	for {
		n := atomic.LoadInt32(&f.inflight)
		if n < 2 {
			if atomic.CompareAndSwapInt32(&f.inflight, n, n+1) {
				break
			}
			continue
		}
		// No FrameDone for >3 slots means the message was lost: recover.
		if at := atomic.LoadInt64(&f.inflightAt); at > 0 &&
			time.Since(time.Unix(0, at)) > 3*frameInterval {
			atomic.StoreInt32(&f.inflight, 0)
			continue
		}
		atomic.AddInt64(&f.lost, 1)
		return
	}
	atomic.StoreInt64(&f.inflightAt, time.Now().UnixNano())
	if !postFrameMessage(f.hwnd) {
		atomic.AddInt32(&f.inflight, -1)
		atomic.StoreInt64(&f.inflightAt, 0)
		return
	}
	atomic.AddInt64(&f.posted, 1)
}

func (f *FrameScheduler) loop() {
	// Absolute timeline: every slot is a multiple of frameInterval. A late
	// paint shifts NOTHING - we skip the missed slot instead of catching up.
	next := time.Now()
	// One reusable Timer: time.NewTimer(wait) on every slot allocates a runtime
	// timer + channel 60x/second and shows up as GC-assisted frame jitter.
	tm := time.NewTimer(0)
	if !tm.Stop() {
		select {
		case <-tm.C:
		default:
		}
	}
	defer tm.Stop()
	for {
		if !f.isActive() {
			// Idle: no frames at all. Wake only on a state change or a dirty
			// region request. Zero CPU while nothing is animating / changing.
			select {
			case <-f.wake:
			case <-f.stop:
				return
			}
			if f.isActive() {
				next = time.Now()
				continue
			}
			if f.takeOneShot() {
				f.post()
			}
			continue
		}

		next = next.Add(frameInterval)
		now := time.Now()
		if next.Before(now) {
			atomic.AddInt64(&f.skipped, 1)
			next = now.Add(frameInterval)
		}
		wait := time.Until(next)
		if wait > 0 {
			tm.Reset(wait)
			select {
			case <-tm.C:
			case <-f.wake:
				if !tm.Stop() {
					select {
					case <-tm.C:
					default:
					}
				}
				next = time.Now()
				continue
			case <-f.stop:
				tm.Stop()
				return
			}
		}
		f.post()
	}
}

func (f *FrameScheduler) Stats() (posted, skipped, lost int64) {
	return atomic.LoadInt64(&f.posted), atomic.LoadInt64(&f.skipped), atomic.LoadInt64(&f.lost)
}
