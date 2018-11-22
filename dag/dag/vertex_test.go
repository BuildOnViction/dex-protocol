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
	"testing"
)

func TestVertex(t *testing.T) {
	v := NewVertex("1", true, nil)

	if v.ID == "" {
		t.Fatalf("Vertex ID expected to be not empty string.\n")
	}
	if v.Value != nil {
		t.Fatalf("Vertex Value expected to be nil.\n")
	}
}

func TestVertex_Parents(t *testing.T) {
	v := NewVertex("1", true, nil)

	numParents := v.Parents.Size()
	if numParents != 0 {
		t.Fatalf("Vertex Parents expected to be 0 but got %d", v.Parents.Size())
	}
}

func TestVertex_Children(t *testing.T) {
	v := NewVertex("1", true, nil)

	numParents := v.Children.Size()
	if numParents != 0 {
		t.Fatalf("Vertex Children expected to be 0 but got %d", v.Children.Size())
	}
}

func TestVertex_Degree(t *testing.T) {
	dag1 := NewDAG()

	vertex1 := NewVertex("1", true, nil)
	vertex2 := NewVertex("2", true, nil)
	vertex3 := NewVertex("3", true, nil)

	degree := vertex1.Degree()
	if degree != 0 {
		t.Fatalf("Vertex1 Degree expected to be 0 but got %d", vertex1.Degree())
	}

	err := dag1.AddVertex(vertex1)
	if err != nil {
		t.Fatalf("Can't add vertex to DAG: %s", err)
	}
	err = dag1.AddVertex(vertex2)
	if err != nil {
		t.Fatalf("Can't add vertex to DAG: %s", err)
	}
	err = dag1.AddVertex(vertex3)
	if err != nil {
		t.Fatalf("Can't add vertex to DAG: %s", err)
	}

	err = dag1.AddEdge(vertex1, vertex2)
	if err != nil {
		t.Fatalf("Can't add edge to DAG: %s", err)
	}

	err = dag1.AddEdge(vertex2, vertex3)
	if err != nil {
		t.Fatalf("Can't add edge to DAG: %s", err)
	}

	degree = vertex1.Degree()
	if degree != 1 {
		t.Fatalf("Vertex1 Degree expected to be 1 but got %d", vertex1.Degree())
	}

	degree = vertex2.Degree()
	if degree != 2 {
		t.Fatalf("Vertex2 Degree expected to be 2 but got %d", vertex2.Degree())
	}
}

func TestVertex_InDegree(t *testing.T) {
	dag1 := NewDAG()

	vertex1 := NewVertex("1", true, nil)
	vertex2 := NewVertex("2", true, nil)
	vertex3 := NewVertex("3", true, nil)

	inDegree := vertex1.InDegree()
	if inDegree != 0 {
		t.Fatalf("Vertex1 InDegree expected to be 0 but got %d", vertex1.InDegree())
	}

	err := dag1.AddVertex(vertex1)
	if err != nil {
		t.Fatalf("Can't add vertex to DAG: %s", err)
	}
	err = dag1.AddVertex(vertex2)
	if err != nil {
		t.Fatalf("Can't add vertex to DAG: %s", err)
	}
	err = dag1.AddVertex(vertex3)
	if err != nil {
		t.Fatalf("Can't add vertex to DAG: %s", err)
	}

	err = dag1.AddEdge(vertex1, vertex2)
	if err != nil {
		t.Fatalf("Can't add edge to DAG: %s", err)
	}

	err = dag1.AddEdge(vertex2, vertex3)
	if err != nil {
		t.Fatalf("Can't add edge to DAG: %s", err)
	}

	inDegree = vertex1.InDegree()
	if inDegree != 0 {
		t.Fatalf("Vertex1 InDegree expected to be 0 but got %d", vertex1.InDegree())
	}

	inDegree = vertex2.InDegree()
	if inDegree != 1 {
		t.Fatalf("Vertex2 InDegree expected to be 1 but got %d", vertex2.InDegree())
	}
}

func TestVertex_OutDegree(t *testing.T) {
	dag1 := NewDAG()

	vertex1 := NewVertex("1", true, nil)
	vertex2 := NewVertex("2", true, nil)
	vertex3 := NewVertex("3", true, nil)

	outDegree := vertex1.OutDegree()
	if outDegree != 0 {
		t.Fatalf("Vertex1 OutDegree expected to be 0 but got %d", vertex1.OutDegree())
	}

	err := dag1.AddVertex(vertex1)
	if err != nil {
		t.Fatalf("Can't add vertex to DAG: %s", err)
	}
	err = dag1.AddVertex(vertex2)
	if err != nil {
		t.Fatalf("Can't add vertex to DAG: %s", err)
	}
	err = dag1.AddVertex(vertex3)
	if err != nil {
		t.Fatalf("Can't add vertex to DAG: %s", err)
	}

	err = dag1.AddEdge(vertex1, vertex2)
	if err != nil {
		t.Fatalf("Can't add edge to DAG: %s", err)
	}

	err = dag1.AddEdge(vertex2, vertex3)
	if err != nil {
		t.Fatalf("Can't add edge to DAG: %s", err)
	}

	outDegree = vertex1.OutDegree()
	if outDegree != 1 {
		t.Fatalf("Vertex1 OutDegree expected to be 1 but got %d", vertex1.OutDegree())
	}

	outDegree = vertex2.OutDegree()
	if outDegree != 1 {
		t.Fatalf("Vertex2 OutDegree expected to be 1 but got %d", vertex2.OutDegree())
	}

	outDegree = vertex3.OutDegree()
	if outDegree != 0 {
		t.Fatalf("Vertex2 OutDegree expected to be 0 but got %d", vertex3.OutDegree())
	}
}

func TestVertex_WithStringValue(t *testing.T) {
	v := NewVertex("1", true, []byte("one"))

	if v.ID == "" {
		t.Fatalf("Vertex ID expected to be not empty string.\n")
	}
	if string(v.Value) != "one" {
		t.Fatalf("Vertex Value expected to be one.\n")
	}
}
