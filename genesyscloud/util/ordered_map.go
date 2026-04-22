package util

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// OrderedMap is a generic map that preserves the insertion order of keys.
// It correctly maintains key order when unmarshalling from JSON, unlike Go's built-in map type.
type OrderedMap[V any] struct {
	keys   []string
	values map[string]V
}

func NewOrderedMap[V any]() *OrderedMap[V] {
	return &OrderedMap[V]{values: make(map[string]V)}
}

func (m *OrderedMap[V]) Set(key string, value V) {
	if _, exists := m.values[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.values[key] = value
}

func (m *OrderedMap[V]) Get(key string) (V, bool) {
	v, ok := m.values[key]
	return v, ok
}

func (m *OrderedMap[V]) Delete(key string) {
	if _, exists := m.values[key]; !exists {
		return
	}
	delete(m.values, key)
	for i, k := range m.keys {
		if k == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

func (m *OrderedMap[V]) Keys() []string {
	return m.keys
}

func (m *OrderedMap[V]) Len() int {
	return len(m.keys)
}

func (m *OrderedMap[V]) UnmarshalJSON(data []byte) error {
	m.keys = nil
	m.values = make(map[string]V)

	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected '{', got %v", t)
	}

	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return err
		}
		key := t.(string)

		var val V
		if err := dec.Decode(&val); err != nil {
			return err
		}
		m.Set(key, val)
	}
	return nil
}

func (m OrderedMap[V]) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range m.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')
		valBytes, err := json.Marshal(m.values[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
