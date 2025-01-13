package cfg

import "iter"

// Graph defines the interface required for Tarjan's SCC algorithm.
type Graph[N comparable] interface {
	Nodes() iter.Seq[N]
	Neighbors(node N) iter.Seq[N]
}

// TarjanSCC represents the state of the algorithm.
type TarjanSCC[N comparable] struct {
	graph    Graph[N]
	index    int
	stack    []N
	onStack  map[N]bool
	indexMap map[N]int
	lowLink  map[N]int
	sccs     [][]N
}

// New creates a new TarjanSCC instance for the given graph.
func NewTarjanSCC[N comparable](graph Graph[N]) *TarjanSCC[N] {
	return &TarjanSCC[N]{
		graph:    graph,
		onStack:  make(map[N]bool),
		indexMap: make(map[N]int),
		lowLink:  make(map[N]int),
		sccs:     [][]N{},
	}
}

// StronglyConnectedComponents computes and returns the strongly connected components of the graph.
func (t *TarjanSCC[N]) StronglyConnectedComponents() [][]N {
	for node := range t.graph.Nodes() {
		if _, exists := t.indexMap[node]; !exists {
			t.strongConnect(node)
		}
	}
	return t.sccs
}

func (t *TarjanSCC[N]) strongConnect(node N) {
	// Set the depth index for node to the smallest unused index
	t.indexMap[node] = t.index
	t.lowLink[node] = t.index
	t.index++
	t.stack = append(t.stack, node)
	t.onStack[node] = true

	// Consider successors of node
	for neighbor := range t.graph.Neighbors(node) {
		if _, exists := t.indexMap[neighbor]; !exists {
			// Successor has not yet been visited; recurse on it
			t.strongConnect(neighbor)
			t.lowLink[node] = min(t.lowLink[node], t.lowLink[neighbor])
		} else if t.onStack[neighbor] {
			// Successor is in the stack and hence in the current SCC
			t.lowLink[node] = min(t.lowLink[node], t.indexMap[neighbor])
		}
	}

	// If node is a root node, pop the stack and generate an SCC
	if t.lowLink[node] == t.indexMap[node] {
		var scc []N
		for {
			top := t.stack[len(t.stack)-1]
			t.stack = t.stack[:len(t.stack)-1]
			t.onStack[top] = false
			scc = append(scc, top)
			if top == node {
				break
			}
		}
		t.sccs = append(t.sccs, scc)
	}
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
