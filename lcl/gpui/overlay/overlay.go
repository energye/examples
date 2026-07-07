package overlay

import (
	"sort"

	"github.com/energye/examples/lcl/gpui/core/math"
)

// Placement describes popup placement relative to an anchor.
type Placement int

const (
	BottomLeft Placement = iota
	BottomRight
	TopLeft
	TopRight
	LeftTop
	RightTop
	Center
)

// LayerKind identifies a portal layer role.
type LayerKind int

const (
	LayerPopup LayerKind = iota
	LayerTooltip
	LayerModal
	LayerMessage
)

// Options controls layer behavior.
type Options struct {
	CloseOnOutside bool
	FocusTrap      bool
	HasMask        bool
}

// Layer is a managed overlay layer.
type Layer struct {
	ID        string
	Kind      LayerKind
	ZIndex    int
	Bounds    math.Rect
	Anchor    math.Rect
	Placement Placement
	Offset    math.Vec2
	Options   Options
}

// Manager stores overlay layers ordered by z-index.
type Manager struct {
	layers []Layer
}

// NewManager creates an empty overlay manager.
func NewManager() *Manager {
	return &Manager{layers: make([]Layer, 0)}
}

// Add inserts or replaces a layer.
func (m *Manager) Add(layer Layer) {
	if m == nil {
		return
	}
	m.Remove(layer.ID)
	m.layers = append(m.layers, layer)
	m.sort()
}

// Remove removes a layer by ID.
func (m *Manager) Remove(id string) {
	if m == nil {
		return
	}
	for i, layer := range m.layers {
		if layer.ID == id {
			m.layers = append(m.layers[:i], m.layers[i+1:]...)
			return
		}
	}
}

// Layers returns layers sorted by z-index ascending.
func (m *Manager) Layers() []Layer {
	if m == nil {
		return nil
	}
	out := make([]Layer, len(m.layers))
	copy(out, m.layers)
	return out
}

// TopmostAt returns the highest layer containing a point.
func (m *Manager) TopmostAt(x, y float32) (Layer, bool) {
	if m == nil {
		return Layer{}, false
	}
	for i := len(m.layers) - 1; i >= 0; i-- {
		if m.layers[i].Bounds.Contains(x, y) {
			return m.layers[i], true
		}
	}
	return Layer{}, false
}

// DismissTargets returns close-on-outside layers above the clicked point.
func (m *Manager) DismissTargets(x, y float32) []Layer {
	if m == nil {
		return nil
	}
	targets := make([]Layer, 0)
	for i := len(m.layers) - 1; i >= 0; i-- {
		layer := m.layers[i]
		if layer.Bounds.Contains(x, y) {
			break
		}
		if layer.Options.CloseOnOutside {
			targets = append(targets, layer)
		}
	}
	return targets
}

func (m *Manager) sort() {
	if m == nil {
		return
	}
	sort.SliceStable(m.layers, func(i, j int) bool {
		if m.layers[i].ZIndex == m.layers[j].ZIndex {
			return m.layers[i].ID < m.layers[j].ID
		}
		return m.layers[i].ZIndex < m.layers[j].ZIndex
	})
}
