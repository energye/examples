package utils

import (
	"github.com/energye/lcl/tool/exec"
	"path/filepath"
)

func RootPath() string {
	return filepath.Join(exec.CurrentDir, "cef")
}

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

func (m *ArrayMap[T]) ContainsValue(value T) bool {
	for _, val := range m.values {
		if val == value {
			return true
		}
	}
	return false
}

func (m *ArrayMap[T]) Add(key string, value T) {
	if m.values == nil {
		m.values = make(map[string]T)
	}
	if _, ok := m.values[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.values[key] = value
}

func (m *ArrayMap[T]) Get(key string) T {
	return m.values[key]
}

func (m *ArrayMap[T]) Keys() []string {
	return m.keys
}

func (m *ArrayMap[T]) Values() (result []T) {
	for _, key := range m.keys {
		result = append(result, m.values[key])
	}
	return
}
func (m *ArrayMap[T]) Iterate(fn func(key string, value T)) {
	if fn == nil {
		return
	}
	for _, key := range m.keys {
		fn(key, m.values[key])
	}
}

func (m *ArrayMap[T]) Count() int {
	return len(m.keys)
}
