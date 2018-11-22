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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

type vertexAtDepth struct {
	Vertex *Vertex
	Depth  int
}

// This internal method provides the option of not sorting the vertices during
// the walk, which we use for the Transitive reduction.
// Some configurations can lead to fully-connected subgraphs, which makes our
// transitive reduction algorithm O(n^3). This is still passable for the size
// of our graphs, but the additional n^2 sort operations would make this
// uncomputable in a reasonable amount of time.
func (d *DAG) depthFirstWalk(start []*Vertex, f DepthWalkFunc) error {

	seen := make(map[string]struct{})
	frontier := make([]*vertexAtDepth, len(start))
	for i, v := range start {
		frontier[i] = &vertexAtDepth{
			Vertex: v,
			Depth:  0,
		}
	}
	for len(frontier) > 0 {
		// Pop the current vertex
		n := len(frontier)
		current := frontier[n-1]
		frontier = frontier[:n-1]

		// Check if we've seen this already and return...
		if _, ok := seen[current.Vertex.ID]; ok {
			continue
		}

		seen[current.Vertex.ID] = struct{}{}

		// Visit the current node
		if err := f(current.Vertex, current.Depth); err != nil {
			return err
		}

		// Visit targets of this in a consistent order.
		targets := current.Vertex.Children.Values()

		for _, t := range targets {
			frontier = append(frontier, &vertexAtDepth{
				Vertex: t,
				Depth:  current.Depth + 1,
			})
		}
	}

	return nil
}

// DAG type implements a Directed Acyclic Graph data structure.
type DAG struct {
	mu       sync.Mutex
	vertices OrderedMap
}

// DepthWalkFunc is a walk function that also receives the current depth of the
// walk as an argument
type DepthWalkFunc func(*Vertex, int) error

// NewDAG creates a new Directed Acyclic Graph or DAG.
func NewDAG() *DAG {
	d := &DAG{
		vertices: *NewOrderedMap(),
	}

	return d
}

// AddVertex adds a vertex to the graph.
func (d *DAG) AddVertex(v *Vertex) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.vertices.Put(v.ID, v)

	return nil
}

// DeleteVertex deletes a vertex and all the edges referencing it from the
// graph.
func (d *DAG) DeleteVertex(vertex *Vertex) error {
	existsVertex := false

	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if vertices exists.
	for _, v := range d.vertices.Values() {
		if v == vertex {
			existsVertex = true
		}
	}
	if !existsVertex {
		return fmt.Errorf("Vertex with ID %v not found", vertex.ID)
	}

	d.vertices.Remove(vertex.ID)

	return nil
}

// AddEdge adds a directed edge between two existing vertices to the graph.
func (d *DAG) AddEdge(tailVertex *Vertex, headVertex *Vertex) error {
	tailExists := false
	headExists := false

	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if vertices exists.
	for _, vertex := range d.vertices.Values() {
		if vertex == tailVertex {
			tailExists = true
		}
		if vertex == headVertex {
			headExists = true
		}
	}
	if !tailExists {
		return fmt.Errorf("Vertex with ID %v not found", tailVertex.ID)
	}
	if !headExists {
		return fmt.Errorf("Vertex with ID %v not found", headVertex.ID)
	}

	// Check if edge already exists.
	for _, childVertex := range tailVertex.Children.Values() {
		if childVertex == headVertex {
			return fmt.Errorf("Edge (%v,%v) already exists", tailVertex.ID, headVertex.ID)
		}
	}

	// Add edge.
	tailVertex.Children.Add(headVertex)
	headVertex.Parents.Add(tailVertex)

	return nil
}

// DeleteEdge deletes a directed edge between two existing vertices from the
// graph.
func (d *DAG) DeleteEdge(tailVertex *Vertex, headVertex *Vertex) error {
	for _, childVertex := range tailVertex.Children.Values() {
		if childVertex == headVertex {
			tailVertex.Children.Remove(headVertex)
			headVertex.Parents.Remove(tailVertex)
		}
	}

	return nil
}

// GetVertex return a vertex from the graph given a vertex ID.
func (d *DAG) GetVertex(id string) (*Vertex, error) {
	var vertex *Vertex

	v, found := d.vertices.Get(id)
	if !found {
		return vertex, fmt.Errorf("vertex %s not found in the graph", id)
	}

	return v, nil
}

// Order return the number of vertices in the graph.
func (d *DAG) Order() int {
	numVertices := d.vertices.Size()

	return numVertices
}

// Size return the number of edges in the graph.
func (d *DAG) Size() int {
	numEdges := 0
	for _, vertex := range d.vertices.Values() {
		numEdges = numEdges + vertex.Children.Size()
	}

	return numEdges
}

// SinkVertices return vertices with no children defined by the graph edges.
func (d *DAG) SinkVertices() []*Vertex {
	var sinkVertices []*Vertex

	for _, vertex := range d.vertices.Values() {
		if vertex.Children.Size() == 0 {
			sinkVertices = append(sinkVertices, vertex)
		}
	}

	return sinkVertices
}

// SourceVertices return vertices with no parent defined by the graph edges.
func (d *DAG) SourceVertices() []*Vertex {
	var sourceVertices []*Vertex

	for _, vertex := range d.vertices.Values() {
		if vertex.Parents.Size() == 0 {
			sourceVertices = append(sourceVertices, vertex)
		}
	}

	return sourceVertices
}

// Successors return vertices that are children of a given vertex.
func (d *DAG) Successors(vertex *Vertex) ([]*Vertex, error) {
	var successors []*Vertex

	_, found := d.GetVertex(vertex.ID)
	if found != nil {
		return successors, fmt.Errorf("vertex %s not found in the graph", vertex.ID)
	}

	for _, v := range vertex.Children.Values() {
		successors = append(successors, v)
	}

	return successors, nil
}

// Predecessors return vertices that are parent of a given vertex.
func (d *DAG) Predecessors(vertex *Vertex) ([]*Vertex, error) {
	var predecessors []*Vertex

	_, found := d.GetVertex(vertex.ID)
	if found != nil {
		return predecessors, fmt.Errorf("vertex %s not found in the graph", vertex.ID)
	}

	for _, v := range vertex.Parents.Values() {
		predecessors = append(predecessors, v)
	}

	return predecessors, nil
}

// String implements stringer interface.
//
// Prints an string representation of this instance.
func (d *DAG) String() string {
	return fmt.Sprintf("DAG Vertices: %d, Edges: %d, Vertices: %s\n",
		d.Order(), d.Size(), d.vertices.Values())
}

// Ancestors returns a Set that includes every Vertex yielded by walking down from the
// provided starting Vertex v.
func (d *DAG) AncestorsWalk(vertex *Vertex, f DepthWalkFunc) error {
	// get all chilren from this vertex
	start := vertex.Children.Values()
	return d.depthFirstWalk(start, f)
}

// TransitiveReduction performs the transitive reduction of graph g in place.
// The transitive reduction of a graph is a graph with as few edges as
// possible with the same reachability as the original graph. This means
// that if there are three nodes A => B => C, and A connects to both
// B and C, and B connects to C, then the transitive reduction is the
// same graph with only a single edge between A and B, and a single edge
// between B and C.
//
// Complexity: O(V(V+E)), or asymptotically O(VE)
func (d *DAG) TransitiveReduction() {
	// For each vertex u in graph g, do a DFS starting from each vertex
	// v such that the edge (u,v) exists (v is a direct descendant of u).

	for _, u := range d.vertices.Values() {
		uTargets := u.Children
		d.depthFirstWalk(uTargets.Values(), func(v *Vertex, depth int) error {
			shared := uTargets.Intersection(v.Children)
			for _, vPrime := range shared.Values() {
				d.DeleteEdge(u, vPrime)
			}
			return nil
		})
	}
}

// Save save the dag to leveldb file
func (d *DAG) Save(filepath string) error {
	db, err := leveldb.OpenFile(filepath, nil)
	defer db.Close()
	if err == nil {
		batch := new(leveldb.Batch)
		for _, vertex := range d.vertices.Values() {

			bytes, _ := json.Marshal([]string{
				string(vertex.Value),
				strconv.FormatBool(vertex.Flag),
				strings.Join(vertex.Parents.Keys(), ","),
			})
			// save
			batch.Put([]byte(vertex.ID), bytes)

		}
		err = db.Write(batch, nil)
	}
	return err
}

// Load restore the dag from level db file
func (d *DAG) Load(filepath string) error {
	db, err := leveldb.OpenFile(filepath, nil)
	defer db.Close()
	if err == nil {
		iter := db.NewIterator(nil, nil)
		edgeSet := make(map[string][]string)
		for iter.Next() {

			key := string(iter.Key())
			bytes := iter.Value()

			var values [3]string

			json.Unmarshal(bytes, &values)

			vertex0 := NewVertex(key, values[1] == "true", []byte(values[0]))
			d.AddVertex(vertex0)
			if len(values[2]) > 0 {
				edgeSet[key] = strings.Split(values[2], ",")
			}
		}
		iter.Release()
		err = iter.Error()

		if err == nil {
			for id, parentIDs := range edgeSet {
				vertex, _ := d.GetVertex(id)
				for _, parentID := range parentIDs {
					parentVertex, _ := d.GetVertex(parentID)
					// add edge
					d.AddEdge(parentVertex, vertex)
				}
			}
		}

	}
	return err
}
