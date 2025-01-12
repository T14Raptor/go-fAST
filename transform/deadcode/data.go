package deadcode

import (
	"slices"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/tools/fastgraph"
	"github.com/t14raptor/go-fast/tools/tarjan"
)

type VarInfo struct {
	// This does not include self-references in a function.
	Usage int
	// This does not include self-references in a function.
	Assign int
}

type Data struct {
	usedNames map[ast.Id]VarInfo
	graph     fastgraph.DirectedGraph[ast.Id, VarInfo]
	entries   map[ast.Id]struct{}
}

func (d *Data) AddDependencyEdge(from, to ast.Id, assign bool) {
	if info, ok := d.graph.EdgeWeight(from, to); ok {
		if assign {
			info.Assign++
		} else {
			info.Usage++
		}
		d.graph.SetEdgeWeight(from, to, info)
	} else {
		info := VarInfo{}
		if assign {
			info.Assign = 1
		} else {
			info.Usage = 1
		}
		d.graph.AddEdge(from, to, info)
	}
}

func (d *Data) SubtractCycles() {
	cycles := tarjan.New(d.graph).StronglyConnectedComponents()

outer:
	for _, cycle := range cycles {
		if len(cycle) == 1 {
			continue
		}

		// We have to exclude cycle from remove list if an outer node refences an item
		// of cycle.
		for _, node := range cycle {
			// It's referenced by an outer node.
			if _, ok := d.entries[node]; ok {
				continue outer
			}

			for neighbor := range d.graph.NeighborsDirected(node, fastgraph.Incoming) {
				// Neighbour in cycle does not matter
				if !slices.Contains(cycle, neighbor) {
					continue outer
				}
			}
		}

		for _, i := range cycle {
			for _, j := range cycle {
				if i == j {
					continue
				}

				// Adjust usage and assignment
				if weight, exists := d.graph.EdgeWeight(i, j); exists {
					entry := d.usedNames[j]
					entry.Usage -= weight.Usage
					entry.Assign -= weight.Assign
					d.usedNames[j] = entry
				}
			}
		}
	}
}
