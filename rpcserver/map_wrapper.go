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
	internal_map *protocol_types.Map
}

// NewMap creates a new Map Wrapper
func NewMap[Key capnp.Ptr, Value capnp.Ptr](s *capnp.Segment) (*Map[Key, Value], error) {
	internal_map, err := protocol_types.NewMap(s)
	if err != nil {
		return nil, err
	}

	return &Map[Key, Value]{
		internal_map: &internal_map,
	}, nil
}

// FromMap converts a capnp map to a wrapper
func FromMap[Key capnp.Ptr, Value capnp.Ptr](m *protocol_types.Map) *Map[Key, Value] {
	return &Map[Key, Value]{
		internal_map: m,
	}
}

// HasEntries returns true if the map has entries
func (m *Map[Key, Value]) HasEntries() bool {
	return m.internal_map.HasEntries()
}

// Get a Segment of the internal map
func (m *Map[Key, Value]) Segment() *capnp.Segment {
	return m.internal_map.Segment()
}

// Entries returns the entries of the map as a go map
func (m *Map[Key, Value]) Entries() (map[Key]Value, error) {
	// Check if we have entries. If not we return an empty map
	result := make(map[Key]Value)
	if !m.internal_map.HasEntries() {
		return result, nil
	}

	entries, err := m.internal_map.Entries()
	if err != nil {
		return nil, err
	}

	for i := 0; i < entries.Len(); i++ {
		entry := entries.At(i)
		key, err := entry.Key()
		if err != nil {
			return nil, err
		}

		value, err := entry.Value()
		if err != nil {
			return nil, err
		}

		result[Key(key)] = Value(value)
	}

	return result, nil
}

type ErrMapTooLarge struct{}

func (e ErrMapTooLarge) Error() string {
	return "Map supplied is larger than the internal map."
}

// SetEntries sets the entries of the map
func (m *Map[Key, Value]) SetEntries(entries map[Key]Value) error {
	// Check if we have any entries
	if !m.internal_map.HasEntries() {
		// Allocate enough entries
		_, err := m.internal_map.NewEntries(int32(len(entries)))
		if err != nil {
			return err
		}
	}

	// Ensure the map is not larger than the internal map
	internal_entries, err := m.internal_map.Entries()
	if err != nil {
		return err
	}
	if len(entries) > internal_entries.Len() {
		return ErrMapTooLarge{}
	}

	// Set the entries. Important: We cant use the Entries() method we defined earlier in this struct as that one is a copy.
	idx := 0
	for key, value := range entries {
		entry := internal_entries.At(idx)
		entry.SetKey(capnp.Ptr(key))
		entry.SetValue(capnp.Ptr(value))
		idx++
	}

	return nil
}
