// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"cmp"
	"slices"
	"sort"
)

type (
	// Node is the comparable type for each node of Tree.
	Node = cmp.Ordered

	// Comparator is returned by the Order function for each of the top Nodes of a Tree for sorting.
	Comparator = cmp.Ordered

	// Tree defines a hierarchy of Nodes of comparable type.
	Tree[N Node] map[N]Tree[N]

	// Table provides an optional dictionary of values for the Nodes. If used, insert new Nodes here first to ensure uniqueness in Tree.
	Table[N Node, V any] map[N]V

	// Order provides a unique Comparator for each Node/Value for sorting.
	Order[N Node, V any, C Comparator] func(N, V) C

	Meta[N Node, V any, C Comparator] struct {
		Tree[N]
		Table[N, V]
		Order[N, V, C]
	}

	// Display shows the value of a Node at particular depth in Tree.
	Display[N Node, V any] func(int, N, V)
)

// Add adds new Nodes as a branch to the Tree.
func (tr Tree[N]) Add(nodes ...N) {
	if len(nodes) > 0 {
		if _, ok := tr[nodes[0]]; !ok {
			tr[nodes[0]] = Tree[N]{}
		}
		tr[nodes[0]].Add(nodes[1:]...)
	}
}

// All walks the tree and returns an ordered map of the nodes.
func (meta Meta[N, V, C]) All() func(yield func(int, N) bool) {
	return func(yield func(int, N) bool) {
		meta.push(0, yield)
	}
}

// push pushes all elements to the yield function.
func (meta Meta[N, V, C]) push(depth int, yield func(int, N) bool) bool {
	for _, node := range meta.order() {
		if !yield(depth, node) {
			return false
		}
		if !(Meta[N, V, C]{Tree: meta.Tree[node], Table: meta.Table, Order: meta.Order}).push(depth+1, yield) {
			return false
		}
	}
	return true
}

// FindTree finds the subtree anchored by a specific node.
func (tr Tree[N]) FindTree(node N) Tree[N] {
	for n, tr := range tr {
		if n == node {
			return Tree[N]{node: tr}
		}
		if tr = tr.FindTree(node); tr != nil {
			return tr
		}
	}
	return nil
}

// order sorts the top nodes of a tree and returns as an ordered slice.
func (meta Meta[N, V, C]) order() []N {
	var nodes []N
	for node := range meta.Tree {
		nodes = append(nodes, node)
	}
	if meta.Order == nil {
		slices.Sort(nodes)
	} else {
		sort.Slice(nodes, func(i, j int) bool {
			nodei := meta.Order(nodes[i], meta.Table[nodes[i]])
			nodej := meta.Order(nodes[j], meta.Table[nodes[j]])
			return nodei < nodej ||
				nodei == nodej && nodes[i] < nodes[j]
		})
	}
	return nodes
}
