package widget

import (
	"sort"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/overlay"
)

// PortalOptions controls a widget rendered in the top-level overlay host.
type PortalOptions struct {
	ID             string
	Kind           overlay.LayerKind
	ZIndex         int
	Anchor         math.Rect
	Bounds         math.Rect
	Placement      overlay.Placement
	Offset         math.Vec2
	Flip           bool
	Clamp          bool
	CloseOnOutside bool
	FocusTrap      bool
	HasMask        bool
	MaskColor      math.Color
	OnDismiss      func(id string)
}

// Portal binds overlay layer metadata to a widget subtree.
type Portal struct {
	Layer     overlay.Layer
	Content   Widget
	Flip      bool
	Clamp     bool
	MaskColor math.Color
	OnDismiss func(id string)
}

// PortalHost owns top-level widget portals that escape parent clipping.
type PortalHost struct {
	manager           *overlay.Manager
	portals           map[string]*Portal
	focus             *FocusManager
	hoverPortalID     string
	pointerCaptureID  string
	pointerCapture    Widget
	pointerCaptureHit Widget
	pointerStart      math.Vec2
	pointerDragging   bool
	previousFocus     Widget // Widget that had focus before portal opened
}

// NewPortalHost creates a portal host backed by an overlay manager.
func NewPortalHost(manager *overlay.Manager) *PortalHost {
	if manager == nil {
		manager = overlay.NewManager()
	}
	return &PortalHost{
		manager: manager,
		portals: make(map[string]*Portal),
		focus:   NewFocusManager(),
	}
}

// Manager returns the backing overlay manager.
func (h *PortalHost) Manager() *overlay.Manager {
	if h == nil {
		return nil
	}
	return h.manager
}

// FocusManager returns the portal focus manager.
func (h *PortalHost) FocusManager() *FocusManager {
	if h == nil {
		return nil
	}
	return h.focus
}

// Children returns portal content widgets ordered by layer z-index.
func (h *PortalHost) Children() []Widget {
	if h == nil {
		return nil
	}
	layers := h.layers()
	children := make([]Widget, 0, len(layers))
	for _, layer := range layers {
		portal := h.portals[layer.ID]
		if portal != nil && portal.Content != nil {
			children = append(children, portal.Content)
		}
	}
	return children
}

// FocusTrapActive reports whether the topmost portal traps keyboard focus.
func (h *PortalHost) FocusTrapActive() bool {
	if h == nil {
		return false
	}
	layers := h.layers()
	if len(layers) == 0 {
		return false
	}
	return layers[len(layers)-1].Options.FocusTrap
}

// Add inserts or replaces a portal.
func (h *PortalHost) Add(content Widget, options PortalOptions) {
	if h == nil || content == nil || options.ID == "" {
		return
	}
	if h.manager == nil {
		h.manager = overlay.NewManager()
	}
	if h.portals == nil {
		h.portals = make(map[string]*Portal)
	}
	if h.focus == nil {
		h.focus = NewFocusManager()
	}

	if existing := h.portals[options.ID]; existing != nil {
		h.unregisterFocusable(existing.Content)
		existing.Content.SetParent(nil)
	}

	// Save previous focus when opening a focus-trapping portal
	if options.FocusTrap && h.focus != nil {
		h.previousFocus = h.focus.Current()
	}

	if owned, ok := content.(interface{ SetOwner(Widget) }); ok {
		owned.SetOwner(content)
	}
	content.SetParent(nil)

	layer := overlay.Layer{
		ID:        options.ID,
		Kind:      options.Kind,
		ZIndex:    options.ZIndex,
		Bounds:    options.Bounds,
		Anchor:    options.Anchor,
		Placement: options.Placement,
		Offset:    options.Offset,
		Options: overlay.Options{
			CloseOnOutside: options.CloseOnOutside,
			FocusTrap:      options.FocusTrap,
			HasMask:        options.HasMask,
		},
	}
	h.portals[options.ID] = &Portal{
		Layer:     layer,
		Content:   content,
		Flip:      options.Flip,
		Clamp:     options.Clamp,
		MaskColor: options.MaskColor,
		OnDismiss: options.OnDismiss,
	}
	h.manager.Add(layer)
	h.registerFocusable(content)
}

// Remove removes a portal by ID.
func (h *PortalHost) Remove(id string) {
	if h == nil || id == "" {
		return
	}
	portal := h.portals[id]
	if portal != nil {
		h.unregisterFocusable(portal.Content)
		portal.Content.SetParent(nil)

		// Restore previous focus when closing a focus-trapping portal
		if portal.Layer.Options.FocusTrap && h.previousFocus != nil {
			if h.focus != nil {
				h.focus.SetFocus(h.previousFocus)
			}
			h.previousFocus = nil
		}

		delete(h.portals, id)
	}
	if h.manager != nil {
		h.manager.Remove(id)
	}
}

// Portal returns a portal by ID.
func (h *PortalHost) Portal(id string) (*Portal, bool) {
	if h == nil {
		return nil, false
	}
	portal, ok := h.portals[id]
	return portal, ok
}

// Layout measures and positions all portals in viewport coordinates.
func (h *PortalHost) Layout(ctx *Context, viewport math.Rect) {
	if h == nil {
		return
	}
	for _, layer := range h.layers() {
		portal := h.portals[layer.ID]
		if portal == nil || portal.Content == nil {
			continue
		}
		bounds := portal.Layer.Bounds
		size := math.NewVec2(bounds.W, bounds.H)
		if size.X <= 0 || size.Y <= 0 {
			measured := portal.Content.Measure(ctx, Constraints{Max: math.NewVec2(viewport.W, viewport.H)})
			if size.X <= 0 {
				size.X = measured.X
			}
			if size.Y <= 0 {
				size.Y = measured.Y
			}
		}
		if size.X <= 0 {
			size.X = viewport.W
		}
		if size.Y <= 0 {
			size.Y = viewport.H
		}
		anchor := portal.Layer.Anchor
		if anchor.W <= 0 && anchor.H <= 0 && anchor.X == 0 && anchor.Y == 0 {
			anchor = viewport
		}
		if portal.usesPlacement() {
			bounds = overlay.Place(anchor, size, viewport, portal.Layer.Placement, overlay.PlacementOptions{
				Offset: portal.Layer.Offset,
				Flip:   portal.Flip,
				Clamp:  portal.Clamp,
			})
		} else {
			bounds.W = size.X
			bounds.H = size.Y
		}
		portal.Layer.Bounds = bounds
		portal.Content.Layout(ctx, math.NewRect(0, 0, bounds.W, bounds.H))
		if h.manager != nil {
			h.manager.Add(portal.Layer)
		}
	}
}

// Render draws all portals above the normal widget tree.
func (h *PortalHost) Render(ctx *Context) {
	if h == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	for _, layer := range h.layers() {
		portal := h.portals[layer.ID]
		if portal == nil || portal.Content == nil || !portal.Content.Visible() {
			continue
		}
		if portal.Layer.Options.HasMask {
			mask := portal.MaskColor
			if mask.A == 0 {
				mask = ctx.Tokens.Global.ColorBgMask
			}
			ctx.Renderer.FillRect(ctx.Viewport, mask)
		}
		bounds := portal.Layer.Bounds
		ctx.Renderer.PushTransform(math.TranslationMatrix(bounds.X, bounds.Y, 0))
		portal.Content.Render(ctx)
		ctx.Renderer.PopTransform()
	}
}

// HandleEvent routes input to the topmost portal before the normal widget tree.
func (h *PortalHost) HandleEvent(ctx *Context, event Event) bool {
	if h == nil || h.manager == nil {
		return false
	}
	switch event.Type {
	case EventMouseDown, EventMouseUp, EventMouseMove, EventMouseWheel, EventDoubleClick:
		return h.handlePointer(ctx, event)
	case EventKeyDown:
		// Handle Escape key to close topmost portal
		if event.Key == 27 { // Escape
			layers := h.layers()
			if len(layers) > 0 {
				topmost := layers[len(layers)-1]
				if topmost.Options.CloseOnOutside {
					portal := h.portals[topmost.ID]
					if portal != nil {
						onDismiss := portal.OnDismiss
						h.Remove(topmost.ID)
						if onDismiss != nil {
							onDismiss(topmost.ID)
						}
						return true
					}
				}
			}
		}
		if h.focus == nil {
			return false
		}
		focused := h.focus.Current()
		if focused != nil {
			return focused.HandleEvent(ctx, event)
		}
		return h.FocusTrapActive()
	case EventCharInput:
		if h.focus == nil {
			return false
		}
		focused := h.focus.Current()
		if focused != nil {
			return focused.HandleEvent(ctx, event)
		}
		return h.FocusTrapActive()
	default:
		return false
	}
}

// DismissOutside removes close-on-outside portals above a point.
func (h *PortalHost) DismissOutside(x, y float32) []string {
	if h == nil || h.manager == nil {
		return nil
	}
	targets := h.manager.DismissTargets(x, y)
	ids := make([]string, 0, len(targets))
	for _, layer := range targets {
		portal := h.portals[layer.ID]
		if portal == nil {
			continue
		}
		ids = append(ids, layer.ID)
		onDismiss := portal.OnDismiss
		h.Remove(layer.ID)
		if onDismiss != nil {
			onDismiss(layer.ID)
		}
	}
	return ids
}

func (h *PortalHost) handlePointer(ctx *Context, event Event) bool {
	if h.pointerCaptureID != "" && (event.Type == EventMouseMove || event.Type == EventMouseUp) {
		return h.dispatchCapturedPointer(ctx, event)
	}
	if event.Type == EventMouseDown {
		if dismissed := h.DismissOutside(event.X, event.Y); len(dismissed) > 0 {
			return true
		}
	}
	layer, ok := h.manager.TopmostAt(event.X, event.Y)
	if !ok {
		if event.Type == EventMouseMove {
			h.updateHover(ctx, "", event)
		}
		return h.consumeTopMask(event.X, event.Y)
	}
	portal := h.portals[layer.ID]
	if portal == nil || portal.Content == nil {
		return false
	}
	if event.Type == EventMouseMove {
		h.updateHover(ctx, layer.ID, event)
	}
	local := math.NewVec2(event.X-layer.Bounds.X, event.Y-layer.Bounds.Y)
	hit := portal.Content.HitTest(local)
	if hit == nil {
		return portal.Layer.Options.HasMask
	}
	if event.Type == EventMouseDown && hit.Focusable() && h.focus != nil {
		h.focus.SetFocus(hit)
	}
	portalEvent := event
	portalEvent.X = local.X
	portalEvent.Y = local.Y
	portalEvent.LocalX = local.X
	portalEvent.LocalY = local.Y
	if event.Type == EventMouseDown {
		hit.SetStateFlag(StateActive, true)
		h.pointerCaptureID = layer.ID
		h.pointerCapture = portal.Content
		h.pointerCaptureHit = hit
		h.pointerStart = math.NewVec2(event.X, event.Y)
		h.pointerDragging = false
	}
	if event.Type == EventMouseUp {
		hit.SetStateFlag(StateActive, false)
	}
	if portal.Content.HandleEvent(ctx, portalEvent) {
		return true
	}
	return portal.Layer.Options.HasMask
}

func (h *PortalHost) dispatchCapturedPointer(ctx *Context, event Event) bool {
	portal := h.portals[h.pointerCaptureID]
	if portal == nil || portal.Content == nil {
		h.clearPointerCapture()
		return false
	}
	bounds := portal.Layer.Bounds
	portalEvent := event
	portalEvent.X = event.X - bounds.X
	portalEvent.Y = event.Y - bounds.Y
	portalEvent.LocalX = portalEvent.X
	portalEvent.LocalY = portalEvent.Y
	handled := portal.Content.HandleEvent(ctx, portalEvent)
	if event.Type == EventMouseMove {
		handled = h.dispatchCapturedDrag(ctx, portal, event) || handled
	}
	if event.Type == EventMouseUp {
		if h.pointerDragging {
			dragEnd := portalEvent
			dragEnd.Type = EventDragEnd
			dragEnd.DeltaX = event.X - h.pointerStart.X
			dragEnd.DeltaY = event.Y - h.pointerStart.Y
			handled = portal.Content.HandleEvent(ctx, dragEnd) || handled
		}
		if h.pointerCaptureHit != nil {
			h.pointerCaptureHit.SetStateFlag(StateActive, false)
		}
		h.clearPointerCapture()
		return true
	}
	return handled
}

func (h *PortalHost) dispatchCapturedDrag(ctx *Context, portal *Portal, event Event) bool {
	if h == nil || portal == nil || portal.Content == nil {
		return false
	}
	dx := event.X - h.pointerStart.X
	dy := event.Y - h.pointerStart.Y
	if !h.pointerDragging && dx*dx+dy*dy < 16 {
		return false
	}
	bounds := portal.Layer.Bounds
	dragEvent := event
	dragEvent.X = event.X - bounds.X
	dragEvent.Y = event.Y - bounds.Y
	dragEvent.LocalX = dragEvent.X
	dragEvent.LocalY = dragEvent.Y
	dragEvent.DeltaX = dx
	dragEvent.DeltaY = dy
	if !h.pointerDragging {
		h.pointerDragging = true
		dragStart := dragEvent
		dragStart.Type = EventDragStart
		portal.Content.HandleEvent(ctx, dragStart)
	}
	dragEvent.Type = EventDragMove
	return portal.Content.HandleEvent(ctx, dragEvent)
}

func (h *PortalHost) clearPointerCapture() {
	if h == nil {
		return
	}
	h.pointerCaptureID = ""
	h.pointerCapture = nil
	h.pointerCaptureHit = nil
	h.pointerDragging = false
}

func (h *PortalHost) updateHover(ctx *Context, id string, event Event) {
	if h == nil || h.hoverPortalID == id {
		return
	}
	if h.hoverPortalID != "" {
		if portal := h.portals[h.hoverPortalID]; portal != nil && portal.Content != nil {
			h.dispatchPortalHover(ctx, portal, event, EventMouseLeave)
		}
	}
	h.hoverPortalID = id
	if h.hoverPortalID != "" {
		if portal := h.portals[h.hoverPortalID]; portal != nil && portal.Content != nil {
			h.dispatchPortalHover(ctx, portal, event, EventMouseEnter)
		}
	}
}

func (h *PortalHost) dispatchPortalHover(ctx *Context, portal *Portal, event Event, eventType EventType) {
	bounds := portal.Layer.Bounds
	hoverEvent := event
	hoverEvent.Type = eventType
	hoverEvent.X = event.X - bounds.X
	hoverEvent.Y = event.Y - bounds.Y
	hoverEvent.LocalX = hoverEvent.X
	hoverEvent.LocalY = hoverEvent.Y
	portal.Content.HandleEvent(ctx, hoverEvent)
}

func (h *PortalHost) consumeTopMask(x, y float32) bool {
	layers := h.layers()
	for i := len(layers) - 1; i >= 0; i-- {
		if portal := h.portals[layers[i].ID]; portal != nil && portal.Layer.Options.HasMask {
			return true
		}
	}
	return false
}

func (h *PortalHost) layers() []overlay.Layer {
	if h == nil || h.manager == nil {
		return nil
	}
	layers := h.manager.Layers()
	sort.SliceStable(layers, func(i, j int) bool {
		if layers[i].ZIndex == layers[j].ZIndex {
			return layers[i].ID < layers[j].ID
		}
		return layers[i].ZIndex < layers[j].ZIndex
	})
	return layers
}

func (h *PortalHost) registerFocusable(widget Widget) {
	if h == nil || h.focus == nil || widget == nil {
		return
	}
	if widget.Focusable() {
		h.focus.Add(widget)
	}
	// Use ParentWidget interface for consistent traversal
	if parent, ok := widget.(ParentWidget); ok {
		for _, child := range parent.Children() {
			h.registerFocusable(child)
		}
	}
}

func (h *PortalHost) unregisterFocusable(widget Widget) {
	if h == nil || h.focus == nil || widget == nil {
		return
	}
	h.focus.Remove(widget)
	// Use ParentWidget interface for consistent traversal
	if parent, ok := widget.(ParentWidget); ok {
		for _, child := range parent.Children() {
			h.unregisterFocusable(child)
		}
	}
}

func (p *Portal) usesPlacement() bool {
	if p == nil {
		return false
	}
	anchor := p.Layer.Anchor
	offset := p.Layer.Offset
	return anchor.X != 0 || anchor.Y != 0 || anchor.W != 0 || anchor.H != 0 ||
		offset.X != 0 || offset.Y != 0 ||
		p.Layer.Placement != overlay.BottomLeft
}
