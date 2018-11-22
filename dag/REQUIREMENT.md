Task:

Construct a directed-acyclic-graph (DAG) with approximately 100,000 randomly generated vertices and 99,999 randomly generated edges.

This DAG should be stored via. LevelDB or whatever external database seems appropriate, and the goal is to be able to query statistics about certain vertices in this DAG as fast as possible.
Parallelism where it seems appropriate is perfectly fine.

There is a specific data format in which vertices are represented in this DAG. The format is:

```json
{
  "id": "a randomly, uniquely generated length-32 byte array that is then hex-encoded into a string. Thus, the ID is a length-64 string.",
  "parents": [an array of parent vertices string IDs],
  "flag": boolean (true or false)
}
```

Additional columns/data may be appended to this format should they be used for improving the runtime of querying the DAG, but "id" and "parents" must strictly be preserved.
Space complexity however is something that should be minimized as much as possible.

These are the following DAG-related algorithms that need to be done:

1. Reach(vertex ID): Count number of ancestors/progeny (children of children) of any given vertex.
2. ConditionalReach(vertex ID, flagCondition): Count number of ancestors/progeny that have their "flag" set to true, or false for any given vertex.
3. List(vertex ID) + ConditionalList(vertex ID, flagCondition): List the ancestors/progeny with the requirements denoted in algorithms 1) and 2).
4. Insert(vertex): Insert vertex to DAG and automatically construct necessary parent/children edges. Note that this DAG may be "incomplete" at any moment in time of its construction.

A DAG is considered "complete" if it is well-connected, and all vertices and edges have been inserted. Given a set of vertices and edges, there is only 1 unique instance of a "complete" DAG.
We consider a DAG incomplete based on the following: say you've added a vertex whose "parents" are not yet added to the database. Until the parent vertices have been added to the database, we are
unable to construct the necessary edges from said vertex to its parents. Hence, no matter what insertion order vertices are added to the graph, after calling Insert(vertex) on all vertices,
the DAG must converge to becoming the singly unique instance of a "complete" DAG.

Make sure to create benchmarks for every single of these algorithms. The time limit for each operation is 250,000 nanoseconds tested on a relatively fast consumer laptop (Razer Blade).
