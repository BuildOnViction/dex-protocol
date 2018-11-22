# DAG

A DAG, Directed acyclic graph implementation in golang.  
[Requirement](./REQUIREMENT.md)

## Install

```bash
go get github.com/syndtr/goleveldb/leveldb
```

## Example of use

```go

// Create the dag
dag1 := dag.NewDAG()

// Create the vertices. Value is []byte.
vertex0 := dag.NewVertex("0", false, []byte("vertex0"))
vertex1 := dag.NewVertex("1", true, []byte("vertex1"))
vertex2 := dag.NewVertex("2", true, []byte("vertex2"))
vertex3 := dag.NewVertex("3", false, []byte("vertex3"))
vertex4 := dag.NewVertex("4", true, []byte("vertex4"))
vertex5 := dag.NewVertex("5", true, []byte("vertex5"))

// Add the vertices to the dag.
dag1.AddVertex(vertex0)
dag1.AddVertex(vertex1)
dag1.AddVertex(vertex2)
dag1.AddVertex(vertex3)
dag1.AddVertex(vertex4)
dag1.AddVertex(vertex5)

// Add the edges (Note that given vertices must exist before adding an
// edge between them).
dag1.AddEdge(vertex0, vertex1)
dag1.AddEdge(vertex1, vertex2)
dag1.AddEdge(vertex2, vertex3)
dag1.AddEdge(vertex3, vertex4)
dag1.AddEdge(vertex3, vertex5)
dag1.AddEdge(vertex4, vertex5)
```

## Save to and load from level db

```go
filepath := "data/dag1"
dag1 := dag.NewDAG()
// ...

// save
err := dag1.Save(filepath)
fmt.Printf("Saved to %s, err: %s\n", filepath, err)

// load
err := dag1.Load(filepath)
fmt.Printf("Loaded from %s, err: %s\n", filepath, err)
fmt.Println(dag1.String())

```

## Count and list ancestors/progeny

```go
var ancestors []*dag.Vertex
// var count int = 0
filterFlagFunc := func(v *dag.Vertex, depth int) error {
  if v.Flag {
    ancestors = append(ancestors, v)
    // count++
  }
  return nil
}
err = dag1.AncestorsWalk(vertex3, filterFlagFunc)

fmt.Printf("Node: %s, Ancestors: %s, err: %s\n", vertex3.ID, ancestors, err)
```

## Transitive reduction

```go
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

fmt.Println(dag1)
```

## Demo using random 64 length hex string key

```go
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

dag1 := ExampleRandom(10)
fmt.Println(dag1)

```

**the result:**

```sh
$ go run example_dag.go
DAG Vertices: 10, Edges: 9, Vertices: [{
	"id": "52fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c649",
	"parents": ["5fb90badb37c5821b6d95526a41a9504680b4e7c8b763a1b1d49d4955c848621"],
	"flag": true,
	"value":
} {
	"id": "81855ad8681d0d86d1e91e00167939cb6694d2c422acd208a0072939487f6999",
	"parents": ["29b0223beea5f4f74391f445d15afd4294040374f6924b98cbf8713f8d962d7c"],
	"flag": true,
	"value":
} {
	"id": "eb9d18a44784045d87f3c67cf22746e995af5a25367951baa2ff6cd471c483f1",
	"parents": ["ff094279db1944ebd7a19d0f7bbacbe0255aa5b7d44bec40f84c892b9bffd436"],
	"flag": true,
	"value":
} {
	"id": "5fb90badb37c5821b6d95526a41a9504680b4e7c8b763a1b1d49d4955c848621",
	"parents": [""],
	"flag": true,
	"value":
} {
	"id": "6325253fec738dd7a9e28bf921119c160f0702448615bbda08313f6a8eb668d2",
	"parents": ["6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f"],
	"flag": true,
	"value":
} {
	"id": "0bf5059875921e668a5bdf2c7fc4844592d2572bcd0668d2d6c52f5054e2d083",
	"parents": ["6325253fec738dd7a9e28bf921119c160f0702448615bbda08313f6a8eb668d2"],
	"flag": true,
	"value":
} {
	"id": "6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f",
	"parents": ["8d019192c24224e2cafccae3a61fb586b14323a6bc8f9e7df1d929333ff99393"],
	"flag": true,
	"value":
} {
	"id": "ff094279db1944ebd7a19d0f7bbacbe0255aa5b7d44bec40f84c892b9bffd436",
	"parents": ["52fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c649"],
	"flag": true,
	"value":
} {
	"id": "29b0223beea5f4f74391f445d15afd4294040374f6924b98cbf8713f8d962d7c",
	"parents": ["0bf5059875921e668a5bdf2c7fc4844592d2572bcd0668d2d6c52f5054e2d083"],
	"flag": true,
	"value":
} {
	"id": "8d019192c24224e2cafccae3a61fb586b14323a6bc8f9e7df1d929333ff99393",
	"parents": ["eb9d18a44784045d87f3c67cf22746e995af5a25367951baa2ff6cd471c483f1"],
	"flag": true,
	"value":
}]

```
