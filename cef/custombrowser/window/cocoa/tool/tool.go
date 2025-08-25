package tool

type ArrayMap[T comparable] struct {
	keys   []string
	values map[string]T
}

func (m *ArrayMap[T]) Del(key string) {
	if _, ok := m.values[key]; ok {
		delete(m.values, key)
		for idx, tmpkey := range m.keys {
			if key == tmpkey {
				m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
				break
			}
		}
	}
}

func NewArrayMap[T comparable]() *ArrayMap[T] {
	return &ArrayMap[T]{keys: make([]string, 0), values: make(map[string]T)}
}

func (m *ArrayMap[T]) ContainsKey(key string) bool {
	if _, ok := m.values[key]; ok {
		return true
	}
	return false
}

func (m *ArrayMap[T]) Count() int {
	return len(m.values)
}

func (m *ArrayMap[T]) Add(key string, value T) {
	if m.values == nil {
		m.values = make(map[string]T)
	}
	m.keys = append(m.keys, key)
	m.values[key] = value
}

func (m *ArrayMap[T]) Get(key string) T {
	return m.values[key]
}

func (m *ArrayMap[T]) Keys() []string {
	return m.keys
}

func (m *ArrayMap[T]) Iterate(fn func(key string, value T) bool) {
	if fn == nil {
		return
	}
	for _, key := range m.keys {
		if fn(key, m.values[key]) {
			break
		}
	}
}

type HashMap[T comparable] struct {
	values map[string]T
}

func NewHashMap[T comparable]() *HashMap[T] {
	return &HashMap[T]{values: make(map[string]T)}
}

func (m *HashMap[T]) ContainsKey(key string) bool {
	if _, ok := m.values[key]; ok {
		return true
	}
	return false
}

func (m *HashMap[T]) ContainsValue(value T) bool {
	for _, val := range m.values {
		if val == value {
			return true
		}
	}
	return false
}

func (m *HashMap[T]) Count() int {
	return len(m.values)
}

func (m *HashMap[T]) Add(key string, value T) {
	if m.values == nil {
		m.values = make(map[string]T)
	}
	m.values[key] = value
}

func (m *HashMap[T]) Get(key string) T {
	return m.values[key]
}

func (m *HashMap[T]) Values() map[string]T {
	return m.values
}

func (m *HashMap[T]) Iterate(fn func(key string, value T) bool) {
	if fn == nil {
		return
	}
	for key, val := range m.values {
		if fn(key, val) {
			break
		}
	}
}

func ToInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	default:
		return 0
	}
}

func ToDouble(value any) float64 {
	switch v := value.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}
