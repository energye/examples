package pipeline

import (
	"fmt"
	"sync"

	"github.com/energye/examples/lcl/gpui/core/math"
)

// SVGIconCache caches parsed SVG icon geometry.
type SVGIconCache struct {
	mu    sync.RWMutex
	items map[string]*SVGIcon
}

// NewSVGIconCache creates an empty icon cache.
func NewSVGIconCache() *SVGIconCache {
	return &SVGIconCache{items: make(map[string]*SVGIcon)}
}

// Get returns a parsed icon from cache or parses and stores it.
func (c *SVGIconCache) Get(pathData string, viewBox math.Rect, fillRule FillRule) (*SVGIcon, error) {
	key := svgIconCacheKey(pathData, viewBox, fillRule)

	c.mu.RLock()
	icon := c.items[key]
	c.mu.RUnlock()
	if icon != nil {
		return icon, nil
	}

	parsed, err := NewSVGIcon(pathData, viewBox, fillRule)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	if existing := c.items[key]; existing != nil {
		c.mu.Unlock()
		return existing, nil
	}
	c.items[key] = parsed
	c.mu.Unlock()
	return parsed, nil
}

// Clear removes all cached icons.
func (c *SVGIconCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*SVGIcon)
}

// Len returns the number of cached icons.
func (c *SVGIconCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

var defaultSVGIconCache = NewSVGIconCache()

// CachedSVGIcon returns an icon from the default geometry cache.
func CachedSVGIcon(pathData string, viewBox math.Rect, fillRule FillRule) (*SVGIcon, error) {
	return defaultSVGIconCache.Get(pathData, viewBox, fillRule)
}

// ClearSVGIconCache clears the default geometry cache.
func ClearSVGIconCache() {
	defaultSVGIconCache.Clear()
}

func svgIconCacheKey(pathData string, viewBox math.Rect, fillRule FillRule) string {
	return fmt.Sprintf("%d:%g,%g,%g,%g:%s", fillRule, viewBox.X, viewBox.Y, viewBox.W, viewBox.H, pathData)
}
