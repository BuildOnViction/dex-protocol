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

package main

import (
	"encoding/hex"
	"io/ioutil"
	"math/rand"
	"time"

	"./dag"
)

func randomID() string {
	key := make([]byte, 32)
	rand.Read(key)

	return hex.EncodeToString(key)
}

func ExampleEdges() *dag.DAG {
	dag1 := dag.NewDAG()
	vertex0 := dag.NewVertex("0", false, []byte("vertex0"))
	vertex1 := dag.NewVertex("1", true, []byte("vertex1"))
	vertex2 := dag.NewVertex("2", true, []byte("vertex2"))
	vertex3 := dag.NewVertex("3", false, []byte("vertex3"))
	vertex4 := dag.NewVertex("4", true, []byte("vertex4"))
	vertex5 := dag.NewVertex("5", true, []byte("vertex5"))

	dag1.AddVertex(vertex0)
	dag1.AddVertex(vertex1)
	dag1.AddVertex(vertex2)
	dag1.AddVertex(vertex3)
	dag1.AddVertex(vertex4)
	dag1.AddVertex(vertex5)

	// Edges
	dag1.AddEdge(vertex0, vertex1)
	dag1.AddEdge(vertex1, vertex2)
	dag1.AddEdge(vertex2, vertex3)
	dag1.AddEdge(vertex3, vertex4)
	dag1.AddEdge(vertex3, vertex5)
	dag1.AddEdge(vertex4, vertex5)

	return dag1
}

func ExampleTransReduction() *dag.DAG {
	dag1 := dag.NewDAG()
	vertex1 := dag.NewVertex("1", true, []byte("vertex1"))
	vertex2 := dag.NewVertex("2", true, []byte("vertex2"))
	vertex3 := dag.NewVertex("3", false, []byte("vertex3"))

	dag1.AddVertex(vertex1)
	dag1.AddVertex(vertex2)
	dag1.AddVertex(vertex3)

	dag1.AddEdge(vertex1, vertex2)
	dag1.AddEdge(vertex1, vertex3)
	dag1.AddEdge(vertex2, vertex3)

	dag1.TransitiveReduction()

	return dag1
}

func ExampleLoad(filepath string) (*dag.DAG, error) {
	dag1 := dag.NewDAG()
	err := dag1.Load(filepath)
	return dag1, err
}

func ExampleSave(dag1 *dag.DAG, filepath string) error {
	err := dag1.Save(filepath)
	return err
}

func ExampleRandom(number int) *dag.DAG {
	dag1 := dag.NewDAG()
	list := make([]int, number)
	vertexList := make([]*dag.Vertex, number)
	for i := 0; i < number; i = i + 1 {
		list[i] = i
		// key := strconv.Itoa(i)
		key := randomID()
		vertex := dag.NewVertex(key, true, nil)
		vertexList[i] = vertex
		dag1.AddVertex(vertex)
	}
	source := rand.NewSource(time.Now().UnixNano())
	shuffleRand := rand.New(source)

	shuffleRand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})

	for i := 0; i < number-1; i = i + 1 {
		dag1.AddEdge(vertexList[list[i]], vertexList[list[i+1]])
	}

	return dag1
}

func main() {

	// dag1 := ExampleTransReduction()
	// fmt.Println(dag1.String())

	// filepath := "data/dag1"
	// dag1, err := ExampleLoad(filepath)
	// fmt.Printf("Loaded from %s, err: %s\n", filepath, err)

	// fmt.Println(dag1.String())

	// vertex3, _ := dag1.GetVertex("2")
	// var ancestors []*dag.Vertex

	// filterFlagFunc := func(v *dag.Vertex, depth int) error {
	// 	if v.Flag {
	// 		ancestors = append(ancestors, v)
	// 	}
	// 	return nil
	// }
	// err = dag1.AncestorsWalk(vertex3, filterFlagFunc)

	// fmt.Printf("Node: %s, Ancestors: %s, err: %s\n", vertex3.ID, ancestors, err)

	// dag1 := ExampleEdges()
	// err := ExampleSave(dag1)
	// fmt.Printf("Saved to %s, err: %s\n", filepath, err)

	dag1 := ExampleRandom(1000)
	ioutil.WriteFile("output.json", []byte(dag1.String()), 0644)
	// fmt.Println(dag1)
}
