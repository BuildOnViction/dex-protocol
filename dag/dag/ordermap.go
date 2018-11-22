// Copyright 2018, John Pham. All rights reserved.
//
// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with this
// work for additional information regarding copyright ownership.  The ASF
// licenses this file to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  See the
// License for the specific language governing permissions and limitations
// under the License.

package dag

import (
	"fmt"
	"sync"
)

// OrderedMap represents an associative array or map abstract data type.
// The underlying storage is in leveldb
type OrderedMap struct {
	// mu Mutex protects data structures below.
	mu sync.Mutex

	// keys is the Set list of keys.
	keys []string

	// store is the Set underlying store of values.
	store map[string]*Vertex
}

// NewOrderedMap creates a new empty OrderedMap.
func NewOrderedMap() *OrderedMap {
	m := &OrderedMap{
		keys:  make([]string, 0),
		store: make(map[string]*Vertex),
	}

	return m
}

// Put adds items to the map.
//
// If a key is found in the map it replaces it value.
func (m *OrderedMap) Put(key string, value *Vertex) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.store[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.store[key] = value
}

// Get returns the value of a key from the OrderedMap.
func (m *OrderedMap) Get(key string) (value *Vertex, found bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	value, found = m.store[key]
	return value, found
}

// Remove deletes a key-value pair from the OrderedMap.
//
// If a key is not found in the map it doesn't fails, just does nothing.
func (m *OrderedMap) Remove(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check key exists
	if _, found := m.store[key]; !found {
		return
	}

	// Remove the value from the store
	delete(m.store, key)

	// Remove the key
	for i := range m.keys {
		if m.keys[i] == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

// Size return the map number of key-value pairs.
func (m *OrderedMap) Size() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.store)
}

// Empty return if the map in empty or not.
func (m *OrderedMap) Empty() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.store) == 0
}

// Keys return the keys in the map in insertion order.
func (m *OrderedMap) Keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.keys
}

// Values return the values in the map in insertion order.
func (m *OrderedMap) Values() []*Vertex {
	m.mu.Lock()
	defer m.mu.Unlock()

	values := make([]*Vertex, len(m.store))
	for i, key := range m.keys {
		values[i] = m.store[key]
	}
	return values
}

// String implements Stringer interface.
//
// Prints the map string representation, a concatenated string of all its
// string representation values in insertion order.
func (m *OrderedMap) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []string
	for i, key := range m.keys {
		result = append(result, fmt.Sprintf("%s:%s", m.keys[i], m.store[key]))
	}

	return fmt.Sprintf("%s", result)
}
