package fastgraph

import (
	"iter"
	"slices"
)

// Direction represents the direction of an edge.
type Direction int

const (
	Incoming Direction = iota
	Outgoing
)

// Edge represents a directed edge with a target node and a direction.
type Edge[N comparable] struct {
	To        N
	Direction Direction
}

// EdgeKey represents a key for uniquely identifying an edge in the graph.
type EdgeKey[N comparable] struct {
	From N
	To   N
}

// DirectedGraph represents a directed graph with generic node and edge weights.
type DirectedGraph[N comparable, E any] struct {
	nodes map[N][]Edge[N]
	edges map[EdgeKey[N]]E
}

// New creates a new DirectedGraph instance.
func New[N comparable, E any]() DirectedGraph[N, E] {
	return DirectedGraph[N, E]{
		nodes: make(map[N][]Edge[N]),
		edges: make(map[EdgeKey[N]]E),
	}
}

// AddNode adds a node to the graph.
func (g *DirectedGraph[N, E]) AddNode(node N) {
	if _, exists := g.nodes[node]; !exists {
		g.nodes[node] = []Edge[N]{}
	}
}

// RemoveNode removes a node and all edges associated with it from the graph.
func (g *DirectedGraph[N, E]) RemoveNode(node N) {
	links, exists := g.nodes[node]
	if !exists {
		return
	}
	for _, link := range links {
		g.removeSingleEdge(link.To, node, Incoming)
		delete(g.edges, EdgeKey[N]{From: node, To: link.To})
	}
	delete(g.nodes, node)
}

// AddEdge adds an edge connecting two nodes to the graph with associated weight.
func (g *DirectedGraph[N, E]) AddEdge(from, to N, weight E) {
	if _, exists := g.edges[EdgeKey[N]{From: from, To: to}]; !exists {
		g.nodes[from] = append(g.nodes[from], Edge[N]{To: to, Direction: Outgoing})
		if from != to {
			g.nodes[to] = append(g.nodes[to], Edge[N]{To: from, Direction: Incoming})
		}
	}
	g.edges[EdgeKey[N]{From: from, To: to}] = weight
}

// RemoveEdge removes an edge from the graph and returns its weight.
func (g *DirectedGraph[N, E]) RemoveEdge(from, to N) {
	g.removeSingleEdge(from, to, Outgoing)
	if from != to {
		g.removeSingleEdge(to, from, Incoming)
	}

	delete(g.edges, EdgeKey[N]{From: from, To: to})
}

// removeSingleEdge removes a single directed edge between two nodes in the specified direction.
func (g *DirectedGraph[N, E]) removeSingleEdge(from, to N, direction Direction) {
	edges, exists := g.nodes[from]
	if !exists {
		return
	}

	idx := slices.IndexFunc(edges, func(e Edge[N]) bool {
		return e.To == to && e.Direction == direction
	})
	if idx != -1 {
		g.nodes[from] = slices.Delete(edges, idx, idx+1)
	}
}

// Neighbors returns an iterator over the neighbors of a node in the specified direction.
func (g *DirectedGraph[N, E]) Neighbors(node N, direction Direction) iter.Seq[N] {
	return func(yield func(N) bool) {
		edges, exists := g.nodes[node]
		if !exists {
			return
		}
		for _, edge := range edges {
			if edge.Direction == direction || edge.To == node {
				if !yield(edge.To) {
					return
				}
			}
		}
	}
}

// EdgeWeight returns the weight of an edge between two nodes.
func (g *DirectedGraph[N, E]) EdgeWeight(from, to N) (E, bool) {
	weight, exists := g.edges[EdgeKey[N]{From: from, To: to}]
	return weight, exists
}

// SetEdgeWeight sets the weight of an edge between two nodes.
func (g *DirectedGraph[N, E]) SetEdgeWeight(from, to N, weight E) {
	g.edges[EdgeKey[N]{From: from, To: to}] = weight
}
