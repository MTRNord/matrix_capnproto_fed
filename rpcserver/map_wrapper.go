package rpcserver

import (
	capnp "capnproto.org/go/capnp/v3"
	protocol_types "github.com/MTRNord/matrix_protobuf_fed/proto/federation/v1/types"
)

/*
 * This is a wrapper for the Map capnproto type which is defined as:
 *
 * ```
 * # A generic map from keys to values.
 * struct Map(Key, Value) {
 *   entries @0 :List(Entry);
 *   struct Entry @0xb000b19244fa63f4 {
 *     key @0 :Key;
 *     value @1 :Value;
 *   }
 * }
 * ```
 *
 * Contrary to normal go maps this map has a fixed size.
 */
type Map[Key capnp.Ptr, Value capnp.Ptr] struct {
	internalMap *protocol_types.Map
	maxSize     int32
}

// NewMap creates a new Map Wrapper
func NewMap[Key capnp.Ptr, Value capnp.Ptr](s *capnp.Segment, maxSize int32) (*Map[Key, Value], error) {
	internalMap, err := protocol_types.NewMap(s)
	if err != nil {
		return nil, err
	}

	return &Map[Key, Value]{
		internalMap: &internalMap,
		maxSize:     maxSize,
	}, nil
}

// FromMap converts a capnp map to a wrapper
func FromMap[Key capnp.Ptr, Value capnp.Ptr](m *protocol_types.Map, maxSize int32) *Map[Key, Value] {
	return &Map[Key, Value]{
		internalMap: m,
		maxSize:     maxSize,
	}
}

// HasEntries returns true if the map has entries
func (m *Map[Key, Value]) HasEntries() bool {
	return m.internalMap.HasEntries()
}

// Get a Segment of the internal map
func (m *Map[Key, Value]) Segment() *capnp.Segment {
	return m.internalMap.Segment()
}

// Entries returns the entries of the map as a go map
func (m *Map[Key, Value]) Entries() (map[Key]*Value, error) {
	// Check if we have entries. If not we return an empty map
	result := make(map[Key]*Value)
	if !m.HasEntries() {
		return result, nil
	}

	entries, err := m.internalMap.Entries()
	if err != nil {
		return nil, err
	}

	for i := 0; i < entries.Len(); i++ {
		entry := entries.At(i)
		key, err := entry.Key()
		if err != nil {
			return nil, err
		}

		value_raw, err := entry.Value()
		if err != nil {
			return nil, err
		}

		value := Value(value_raw)
		result[Key(key)] = &value
	}

	return result, nil
}

func (m *Map[Key, Value]) AddEntry(key Key, value Value) error {
	// Check if we have any entries
	if !m.internalMap.HasEntries() {
		// Allocate enough entries
		_, err := m.internalMap.NewEntries(m.maxSize)
		if err != nil {
			return err
		}
	}
	internalEntries, err := m.internalMap.Entries()
	if err != nil {
		return err
	}

	entry := internalEntries.At(internalEntries.Len() - 1)
	err = entry.SetKey(capnp.Ptr(key))
	if err != nil {
		return err
	}
	return entry.SetValue(capnp.Ptr(value))
}
