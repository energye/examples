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
// This scheduler reproduces that with a wall-clock-aligned 60Hz cadence and a
// private WM_APP message posted to the main thread. GDI has no VSync API in
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
//
// ---------------------------------------------------------------------------
// WIN32 INPUT CONTRACT (hard-won, do not "optimise" back)
//
// GetMessage retrieves messages in this order: sent messages, POSTED messages,
// hardware input (mouse/keyboard/NC), WM_PAINT, WM_TIMER. A private frame
// message posted with PostMessageW is therefore HIGHER priority than every
// mouse click, key press and title-bar button.
//
// Once a frame costs more than the 16.667ms budget (a maximized 2880x1557
// client on a 200% DPI display renders in ~30ms), posting frames faster than
// the main thread can paint means the post queue is NEVER empty. Windows then
// never retrieves input: the title-bar buttons / restore / all mouse ops look
// frozen while the animation keeps rendering. Measured before the fix: a click
// took 382ms to dispatch with the pump on, 14ms with the pump off, even though
// GetCapture()==0 and WM_NCHITTEST returned HTCAPTION the whole time.
//
// Rules that keep input alive:
//   * At most ONE frame message may be outstanding. A slow frame is NOT a
//     "lost message": the old 3-slot watchdog reset inflight to 0 during a
//     long static re-bake and flooded the queue with frame messages. That was
//     the resize/maximize hang.
//   * The pump waits for FrameDone before posting the next slot (closed loop).
//     A slow frame shifts the timeline instead of queuing a catch-up burst.
//   * After an over-budget frame the post queue is deliberately left EMPTY
//     for inputBreath so the main thread's GetMessage can pull pending input.
//   * The stale-frame recovery timeout must be far longer than any legitimate
//     frame (a full static re-bake at 2880x1557 is ~190ms), never 3 slots.

const (
	wmRequestFrame = 0x8000 + 1 // WM_APP + 1, private frame message
	frameInterval  = time.Second / 60

	// 1.5ms is enough for the main thread to enter GetMessage and retrieve any
	// pending mouse/keyboard/NC message while the post queue is empty. Only
	// paid for over-budget frames, so the normal 60Hz cadence is unchanged.
	inputBreath = 1500 * time.Microsecond
	// A posted message to a live HWND is never lost. Only assume otherwise
	// after a genuinely pathological stall (window destroyed, debugger break).
	staleFrame = 3 * time.Second
)

type FrameScheduler struct {
	hwnd uintptr

	mu      sync.Mutex
	active  bool // animation running -> continuous 60Hz frames
	oneShot bool // idle dirty region -> a single frame
	started bool

	wake chan struct{}
	stop chan struct{}
	done chan struct{} // signalled by FrameDone (closed-loop pacing)

	inflight   int32 // 1 = a frame message is posted but not yet consumed
	inflightAt int64 // UnixNano of the post, used only for stale recovery
	frameCost  int64 // microseconds spent in the last PREPARE+PRESENT
	posted     int64
	skipped    int64
	lost       int64
}

func NewFrameScheduler(hwnd uintptr) *FrameScheduler {
	return &FrameScheduler{
		hwnd: hwnd,
		wake: make(chan struct{}, 1),
		stop: make(chan struct{}),
		done: make(chan struct{}, 1),
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
// Releasing inflight here (and only here) is what makes the pump closed-loop:
// no second frame message can be queued while a frame is being painted.
func (f *FrameScheduler) FrameDone(costUs int64) {
	atomic.StoreInt32(&f.inflight, 0)
	atomic.StoreInt64(&f.inflightAt, 0)
	atomic.StoreInt64(&f.frameCost, costUs)
	select {
	case f.done <- struct{}{}:
	default:
	}
	// Only an IDLE one-shot needs re-notifying. While animating, the loop is
	// already timing the next 60Hz slot - waking it would corrupt the cadence.
	if !f.isActive() && f.hasOneShot() {
		f.notify()
	}
}

// waitConsumed is the closed loop: it blocks until the outstanding frame
// message has been consumed. There is deliberately NO short deadline - a slow
// frame (or a ~190ms static re-bake at 2880x1557) shifts the next slot instead
// of dropping it. Dropping a whole slot cascaded into a false "39 FPS" reading
// after one slow frame. A 250ms poll is only a safety valve for a frame whose
// WndProc never returns at all (window destroyed).
func (f *FrameScheduler) waitConsumed() bool {
	for atomic.LoadInt32(&f.inflight) != 0 {
		tm := time.NewTimer(250 * time.Millisecond)
		select {
		case <-f.done:
			if !tm.Stop() {
				select {
				case <-tm.C:
				default:
				}
			}
		case <-tm.C:
			if at := atomic.LoadInt64(&f.inflightAt); at > 0 &&
				time.Since(time.Unix(0, at)) > staleFrame {
				atomic.StoreInt32(&f.inflight, 0)
				atomic.StoreInt64(&f.inflightAt, 0)
			}
		case <-f.stop:
			tm.Stop()
			return false
		}
	}
	return true
}

// post queues exactly ONE frame message. It is the only place that may raise
// inflight, and it refuses to do so while the previous slot is unconsumed.
func (f *FrameScheduler) post() bool {
	if !atomic.CompareAndSwapInt32(&f.inflight, 0, 1) {
		// A slow paint / static re-bake is NOT a lost message. The old
		// "3 slots without FrameDone => reset inflight" rule fired during the
		// ~190ms re-bake and dumped dozens of frame messages into the post
		// queue, starving all input after a resize / maximize.
		if at := atomic.LoadInt64(&f.inflightAt); at > 0 &&
			time.Since(time.Unix(0, at)) > staleFrame {
			atomic.StoreInt32(&f.inflight, 0)
			atomic.StoreInt64(&f.inflightAt, 0)
		}
		atomic.AddInt64(&f.lost, 1)
		return false
	}
	atomic.StoreInt64(&f.inflightAt, time.Now().UnixNano())
	if !postFrameMessage(f.hwnd) {
		atomic.StoreInt32(&f.inflight, 0)
		atomic.StoreInt64(&f.inflightAt, 0)
		return false
	}
	atomic.AddInt64(&f.posted, 1)
	return true
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
			// If a frame message is still in flight, leave oneShot set:
			// FrameDone() will wake us when the slot is consumed.
			if atomic.LoadInt32(&f.inflight) == 0 && f.takeOneShot() {
				f.post()
			}
			continue
		}

		next = next.Add(frameInterval)
		now := time.Now()
		if next.Before(now) {
			// Behind schedule: do NOT queue catch-up frames and do NOT wait
			// another whole slot. The next slot is "as soon as the previous
			// frame has been consumed" (see waitConsumed below).
			atomic.AddInt64(&f.skipped, 1)
			next = now
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

		// Closed loop: a slow frame shifts this slot instead of queueing a
		// second frame message. The post queue stays empty between frames.
		if !f.waitConsumed() {
			continue
		}
		// Over-budget frame: leave the post queue EMPTY for a moment so
		// GetMessage can retrieve pending mouse/keyboard/NC input. Input has
		// LOWER priority than posted messages on Win32 - this gap is why the
		// title bar and all mouse ops stay alive on a large/slow window.
		if atomic.LoadInt64(&f.frameCost) > int64(frameInterval) {
			time.Sleep(inputBreath)
		}
		// SetActive(false) may have arrived while we were waiting for the
		// previous frame / breathing. Recheck IMMEDIATELY before posting:
		// posting into an idle pump repaints once after "pause" and can chain
		// one-shot notifications (the "24 idle frames" schedtest failure).
		if !f.isActive() {
			next = time.Now()
			f.notify() // let the idle branch consume any pending one-shot
			continue
		}
		f.post()
	}
}

func (f *FrameScheduler) Stats() (posted, skipped, lost int64) {
	return atomic.LoadInt64(&f.posted), atomic.LoadInt64(&f.skipped), atomic.LoadInt64(&f.lost)
}
